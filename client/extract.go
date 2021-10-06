package client

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/axllent/myback/logger"
	"github.com/klauspost/compress/zstd"
)

var dumpMatch = regexp.MustCompile(`\-[a-z0-9]{8}\.sql(\.zst)?$`)
var scannedPaths = make(map[string]bool)
var scannedFiles = make(map[string]bool)

// ExtractPaths will extract compressed backups and write a SQL file
func ExtractPaths(writeTo string, paths []string) error {
	dumps := []string{}
	for _, p := range paths {
		matches, _ := filepath.Glob(p)
		for _, match := range matches {
			dumps = append(dumps, findDumps(match)...)
		}
	}

	if len(dumps) == 0 {
		return errors.New("no backups found")
	}

	f, err := os.Create(writeTo)
	if err != nil {
		return fmt.Errorf("error writing %s: %s", writeTo, err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			logger.Log().Errorf("Error closing file: %s\n", err.Error())
		}
	}()

	for _, dump := range dumps {
		src, err := os.Open(filepath.Clean(dump))
		if err != nil {
			logger.Log().Error(err.Error())
			continue
		}

		logger.Log().Infof("Reading from %s", dump)

		if isCompressed(dump) {
			reader, err := zstd.NewReader(src)
			if err != nil {
				logger.Log().Error(err.Error())
				continue
			}

			_, err = io.Copy(f, reader) // #nosec
			if err != nil {
				logger.Log().Error(err.Error())
				continue
			}

			reader.Close()

		} else {
			_, err = io.Copy(f, src)
			if err != nil {
				logger.Log().Error(err.Error())
				continue
			}
		}

		if err := src.Close(); err != nil {
			logger.Log().Error(err.Error())
			continue
		}
	}

	return nil
}

func findDumps(loc string) []string {
	loc = filepath.Clean(loc)
	files := []string{}
	if isFile(loc) {
		_, found := scannedFiles[loc]
		if !found {
			if dumpMatch.MatchString(loc) {
				scannedFiles[loc] = true
				files = append(files, loc)
			}
		}
	} else if isDir(loc) {
		_, found := scannedPaths[loc]
		if !found {
			scannedPaths[loc] = true
			err := filepath.Walk(loc, func(loc string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				files = append(files, findDumps(loc)...)
				return nil
			})
			if err != nil {
				logger.Log().Error(err.Error())
			}
		}
	}

	return files
}
