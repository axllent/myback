package server

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/axllent/myback/logger"
	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// Database struct
type Database struct {
	Name string `json:"name"`
}

// Table struct
type Table struct {
	Name       string `json:"name"`
	Type       string `json:"Type"`
	Rows       int64  `json:"rows"`
	CreateTime string `json:"create_time"`
	Checksum   int64  `json:"checksum"`
	Create     string `json:"create"`
	CreateHash string `json:"create_hash"`
}

// Tables struct
type Tables struct {
	Database string  `json:"database"`
	Create   string  `json:"create"`
	Tables   []Table `json:"tables"`
}

// DBConnect connects to the database with an optional table
// Note: db must be closed from calling function!
func dbConnect(table, user, pass string) (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, pass, Config.MySQLHost, Config.MySQLPort, table))

	if err != nil {
		return nil, err
	}

	return db, err
}

// ListDatabases controller
func listDatabases(w http.ResponseWriter, r *http.Request) {
	user, pass, _ := r.BasicAuth()

	db, err := dbConnect("", user, pass)
	if err != nil {
		httpError(w, err)
		return
	}
	defer db.Close()

	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		httpError(w, err)
		return
	}
	defer rows.Close()

	var name string

	var results = []Database{}

	for rows.Next() {
		err := rows.Scan(&name)
		if err != nil {
			httpError(w, err)
			return
		}
		// exclude information_schema & performance_schema databases
		if name == "information_schema" || name == "performance_schema" {
			continue
		}
		db := encodeName(name)
		results = append(results, Database{db})
	}

	logger.Log().Infof("%s %s", ip(r), r.RequestURI)

	jsonResponse(w, results)
}

// ListTables controller
func listTables(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	user, pass, _ := r.BasicAuth()

	database := decodeName(vars["database"])

	db, err := dbConnect(database, user, pass)
	if err != nil {
		httpError(w, err)
		return
	}
	defer db.Close()

	var ignore sql.NullString

	var results = Tables{}
	results.Database = vars["database"]
	results.Tables = []Table{}

	_ = db.QueryRow("SHOW CREATE DATABASE IF NOT EXISTS `"+escapeBackticks(database)+"`").Scan(&ignore, &results.Create)

	results.Create = fmt.Sprintf("%s;\n\nUSE `%s`;\n\n", results.Create, escapeBackticks(database))

	rows, err := db.Query("SELECT TABLE_NAME, ENGINE, TABLE_ROWS, CREATE_TIME FROM information_schema.tables WHERE table_schema = ?", escapeBackticks(database))

	if err != nil {
		pageNotFound(w, r)
		return
	}
	defer rows.Close()

	for rows.Next() {

		table := Table{}

		var engine sql.NullString
		var tblRows sql.NullInt64
		var tblChecksum sql.NullInt64
		var createTime mysql.NullTime

		err := rows.Scan(&table.Name, &engine, &tblRows, &createTime)
		if err != nil {
			httpError(w, err)
			return
		}

		if engine.Valid {
			table.Type = engine.String
		} else {
			table.Type = "View"
		}

		if tblRows.Valid {
			table.Rows = tblRows.Int64
		}

		if createTime.Valid {
			table.CreateTime = createTime.Time.Local().String()
		}

		// checksum
		if table.Type == "View" {
			// View table
			_ = db.QueryRow("SHOW CREATE TABLE `"+escapeBackticks(table.Name)+"`").Scan(&ignore, &table.Create, &ignore, &ignore)
		} else {
			// Data table
			err = db.QueryRow("CHECKSUM TABLE `"+escapeBackticks(table.Name)+"`").Scan(&ignore, &tblChecksum)
			if err != nil {
				httpError(w, err)
				return
			}
			// checksum
			if tblChecksum.Valid {
				table.Checksum = tblChecksum.Int64
			}

			_ = db.QueryRow("SHOW CREATE TABLE `"+escapeBackticks(table.Name)+"`").Scan(&ignore, &table.Create)
		}

		// generate the CREATE hash
		table.CreateHash = hashString(table.Create)[0:8]

		// return an encoded name for URL / file parsing
		table.Name = encodeName(table.Name)

		results.Tables = append(results.Tables, table)
	}

	logger.Log().Infof("%s %s", ip(r), r.RequestURI)

	jsonResponse(w, results)
}
