package server

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/axllent/myback/logger"
	"github.com/gorilla/mux"
)

// Dump returns mysqldump of a database or table
func dump(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	user, pass, _ := r.BasicAuth()

	database := decodeName(vars["database"])

	// optional table based on route
	table := decodeName(vars["table"])

	db, err := dbConnect(database, user, pass)
	if err != nil {
		httpError(w, err)
		return
	}
	defer db.Close()

	// start building mysqldump arguments
	args := createMySQLDumpAuthArgs(user)

	boolOptions := []string{"add-drop-database", "add-drop-table", "compact", "complete-insert", "hex-blob", "no-data"}

	for _, option := range boolOptions {
		val := r.URL.Query().Get(option)
		if isValidBool(val) {
			args = append(args, "--"+option+"="+strings.ToLower(val))
		}
	}

	stringOptions := []string{"where"}

	for _, option := range stringOptions {
		val := r.URL.Query().Get(option)
		if val != "" {
			args = append(args, "--"+option+"="+val)
		}
	}

	if database != "" {
		rows, err := db.Query("USE `" + escapeBackticks(database) + "`")
		if err != nil {
			// database does not exist
			pageNotFound(w, r)
			return
		}
		defer rows.Close()

		if table != "" {
			var engine sql.NullString
			err = db.QueryRow("SELECT ENGINE FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?", database, table).Scan(&engine)
			if err != nil {
				pageNotFound(w, r)
				return
			}

			if engine.String == "MEMORY" {
				args = append(args, "--no-data=1")
			}
		}

		args = append(args, database)

		if table != "" {
			args = append(args, table)
		}

	} else {
		args = append(args, "--all-databases")
	}

	cmd := exec.Command(Config.MySQLDump, args...) // #nosec

	// Set password in environment for mysqldump
	cmd.Env = append(os.Environ(), "MYSQL_PWD="+pass)

	// temporary sql file
	dumpFile := tempFileName("myback-", ".sql")
	defer deleteFile(dumpFile)

	// catch errors
	stdOut, err := os.OpenFile(filepath.Clean(dumpFile), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		httpError(w, err)
		logger.Log().Error(err.Error())
		return
	}

	cmd.Stdout = stdOut

	stderr := new(bytes.Buffer)
	cmd.Stderr = stderr

	defer func() {
		if err := stdOut.Close(); err != nil {
			logger.Log().Errorf("Error closing file: %s", err.Error())
		}
	}()

	if err := cmd.Run(); err != nil {
		if errResp, err := io.ReadAll(stderr); err == nil {
			response := strings.TrimSpace(fmt.Sprint(string(errResp)))
			httpError(w, errors.New(string(response)))
			logger.Log().Error(response)
			return
		}

		httpError(w, err)
		logger.Log().Error(err.Error())
		return
	}

	dump, err := os.Open(filepath.Clean(dumpFile))
	if err != nil {
		httpError(w, err)
		logger.Log().Errorf("Error opening dump file: %s", err.Error())
		return
	}

	defer func() {
		if err := dump.Close(); err != nil {
			logger.Log().Errorf("Error closing file: %s", err.Error())
		}
	}()

	fileName := encodeName(vars["database"])
	if table != "" {
		fileName += "-" + encodeName(table)
	}

	w.Header().Set("Content-Disposition", `inline; filename="`+fileName+`.sql"`)

	stat, _ := os.Stat(dumpFile)
	w.Header().Set("Content-Length", fmt.Sprintf("%v", stat.Size()))

	// stream dump to http
	if _, err := io.Copy(w, dump); err != nil {
		logger.Log().Errorf("Error writing stream: %s", err.Error())
	}

	logger.Log().Infof("%s %s", ip(r), r.RequestURI)
}

// DeleteFile deleted a file
func deleteFile(f string) {
	if err := os.Remove(f); err != nil {
		logger.Log().Errorf(err.Error())
	}
}

// TempFileName generates a temporary filename
func tempFileName(prefix, suffix string) string {
	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		logger.Log().Errorf(err.Error())
		return ""
	}
	return filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes)+suffix)
}

// CreateMySQLDumpAuthArgs creates the initial args
func createMySQLDumpAuthArgs(user string) []string {
	args := []string{"-u", user, "-h", Config.MySQLHost, "--skip-lock-tables", "--skip-add-locks"}

	if Config.MySQLSSL {
		args = append(args, "--ssl=true")
	} else {
		args = append(args, "--ssl=false")
	}

	return args
}
