package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/axllent/myback/logger"
	"github.com/klauspost/compress/zstd"
)

var (
	// a lookup array for databases to ignore
	onlySlice = []*regexp.Regexp{}
	// a lookup array for databases to ignore
	ignoreSlice = []*regexp.Regexp{}
	// a lookup array for tables to skip data
	nodataSlice = []*regexp.Regexp{}
)

// Backup will perform the backup
func Backup() []error {
	synced := make(map[string][]string)

	errors := []error{}

	// make sure repo directory exists
	if _, err := os.Stat(Config.Repo); os.IsNotExist(err) {
		logger.Log().Infof("Creating repo directory: %s", Config.Repo)
		err := os.MkdirAll(Config.Repo, os.ModePerm)
		if err != nil {
			errors = append(errors, fmt.Errorf("cannot create repo directory: %s", Config.Repo))

			return errors
		}
	}

	logger.Log().Debugf("Fetching list of databases from %s", Config.URL)

	dbresponse, err := getFile(Config.URL + "/db")
	if err != nil {
		errors = append(errors, fmt.Errorf("error: %s", err))
		return errors
	}

	var jsondbs = []GHMDDatabase{}

	err = json.Unmarshal(dbresponse, &jsondbs)
	if err != nil {
		errors = append(errors, fmt.Errorf("unexpected result from %s: %s", Config.URL+"/db", err))
		return errors
	}

	for _, database := range jsondbs {

		if len(onlySlice) > 0 && !inRegexpArray(database.Name, onlySlice) {
			logger.Log().Debugf("Skipping database: %s", database.Name)
			continue
		}

		if inRegexpArray(database.Name, ignoreSlice) {
			logger.Log().Debugf("Skipping database: %s", database.Name)
			continue
		}

		if inArray(database.Name, Config.Ignore) {
			// skip database
			continue
		}
		// ensure database directory exists
		dbdir := filepath.Join(Config.Repo, database.Name)
		if _, err := os.Stat(dbdir); os.IsNotExist(err) {
			err := os.MkdirAll(dbdir, os.ModePerm)
			if err != nil {
				errors = append(errors, fmt.Errorf("cannot create database directory: %s", dbdir))
				continue
			}
		}

		tables, err := dumpModifiedTables(database.Name, dbdir)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		synced[database.Name] = tables
	}

	if len(errors) > 0 {
		errors = append(errors, fmt.Errorf("%d errors encountered, skipping deletion", len(errors)))
	} else {
		deleteOldData(Config.Repo, synced)
	}

	return errors
}

// DumpModifiedTables dumps only changed tables
func dumpModifiedTables(database string, dbdir string) ([]string, error) {
	var db = Database{}
	var tables []string
	dbresponse, err := getFile(Config.URL + "/db/" + database)
	if err != nil {
		return tables, err
	}

	if err := json.Unmarshal(dbresponse, &db); err != nil {
		logger.Log().Fatalf("Unexpected result from %s: %s", Config.URL+"/db"+database, err)
		return tables, err
	}

	ext := ".sql"
	if Config.Compress {
		ext = ext + ".zst"
	}

	dbFilename := fmt.Sprintf("database-%s%s", hashString(db.Create), ext)

	dbfile := filepath.Join(dbdir, dbFilename)

	if !isFile(dbfile) {
		f, err := os.Create(dbfile)
		if err != nil {
			logger.Log().Fatalf("Error writing %s: %s", dbfile, err)
			return tables, err
		}

		if Config.Compress {
			w, err := zstd.NewWriter(f, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
			if err != nil {
				return tables, err
			}
			if _, err := w.Write([]byte(db.Create)); err != nil {
				return tables, err
			}
			if err := w.Close(); err != nil {
				return tables, err
			}
		} else {
			_, err = f.WriteString(db.Create)

			if err != nil {
				logger.Log().Fatalf("Error writing %s: %s", dbfile, err)
				err := f.Close()

				return tables, err
			}
		}

		if err := f.Close(); err != nil {
			return tables, err
		}
	}

	// add DB to list of files
	tables = append(tables, dbFilename)

	for _, table := range db.Tables {
		// Optional dump parameters
		queryParams := make(map[string]string)

		// lookup name
		lookupName := fmt.Sprintf("%s.%s", database, table.Name)

		if inRegexpArray(lookupName, ignoreSlice) {
			logger.Log().Debugf("Skipping table: %s", lookupName)
			continue
		}

		if inRegexpArray(lookupName, nodataSlice) {
			logger.Log().Debugf("Skipping data: %s", lookupName)
			table.Checksum = 0
			queryParams["no-data"] = "1"
		}

		// selective where?
		if where, ok := Config.WhereMap[lookupName]; ok {
			queryParams["where"] = where
		}

		// maps are not guaranteed to be the same from one iteration to the next
		// so multiple queries can result in different results
		keys := make([]string, 0, len(queryParams))
		for k := range queryParams {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		paramsStr := ""
		for _, k := range keys {
			paramsStr = fmt.Sprintf("%s[%s=%s]", paramsStr, k, queryParams[k])
		}

		// Check the table is not ignored
		// generate the table name
		ext := ".sql"
		if Config.Compress {
			ext = ext + ".zst"
		}
		tblFilename := fmt.Sprintf("t-%s-%d-%s%s", table.Name, table.Checksum, hashString(table.Create+paramsStr), ext)

		tables = append(tables, tblFilename)
		tblSave := filepath.Join(dbdir, tblFilename)

		if !isFile(tblSave) {
			err := downloadToFile(Config.URL+"/dump/"+database+"/"+table.Name, queryParams, tblSave)
			if err != nil {
				logger.Log().Errorf("Unable to download %s/%s: %s", database, table.Name, err.Error())
			} else {
				logger.Log().Infof("Saved %s/%s", database, tblFilename)
			}
		}
	}

	return tables, nil
}

// DeleteOldData deletes old databases and files
func deleteOldData(repo string, tables map[string][]string) {
	directories, err := ioutil.ReadDir(repo)
	if err != nil {
		logger.Log().Fatalf("Error reading %s: %s", repo, err)
		return
	}

	for _, dir := range directories {
		dbdir := filepath.Join(repo, dir.Name())
		tables, ok := tables[dir.Name()]
		if !ok {
			logger.Log().Infof("Deleted database: %s", dbdir)
			err := os.RemoveAll(dbdir)
			if err != nil {
				logger.Log().Fatalf("Error deleting %s: %s", dbdir, err)
			}
			continue
		}
		// list files inside of database directory
		files, err := ioutil.ReadDir(dbdir)
		if err != nil {
			logger.Log().Fatalf("Unexpected result from %s", err)
			return
		}
		for _, file := range files {
			if !inArray(file.Name(), tables) {
				delFile := filepath.Join(dbdir, file.Name())
				err := os.RemoveAll(delFile)
				if err != nil {
					logger.Log().Fatalf("Error deleting %s: %s", delFile, err)
					return
				}
				logger.Log().Infof("Deleted: %s/%s", dir.Name(), file.Name())
			}
		}
	}
}
