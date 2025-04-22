package client

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/axllent/myback/logger"
	"github.com/klauspost/compress/zstd"
)

// UserAgent string
var UserAgent = "MyBack client"

// GetFile returns a GET response using basic auth
func getFile(url string) ([]byte, error) {

	logger.Log().Infof("Fetching %s", url)

	client := http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte{}, err
	}

	req.Header.Set("User-Agent", UserAgent)

	// This one line implements the authentication required for the task.
	if Config.Username != "" {
		req.SetBasicAuth(Config.Username, Config.Password)
	}

	res, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}

	b, err := io.ReadAll(res.Body)
	if err := res.Body.Close(); err != nil {
		return []byte{}, err
	}

	if err != nil {
		return b, err
	}

	if res.StatusCode == 401 {
		return b, errors.New("Unauthorised")
	}

	if res.StatusCode != 200 {
		b, _ := io.ReadAll(res.Body)
		return b, errors.New(string(b))
	}

	return b, nil
}

// DownloadToFile will download a url to a local file. It will compress the file with
// zstd if Config.Compress is true
func downloadToFile(url string, queryParams map[string]string, filepath string) error {
	client := http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	q := req.URL.Query()
	for k, v := range queryParams {
		q.Add(k, v)
	}

	req.Header.Set("User-Agent", UserAgent)

	req.URL.RawQuery = q.Encode()

	if err != nil {
		return err
	}

	// This one line implements the authentication required for the task.
	if Config.Username != "" {
		req.SetBasicAuth(Config.Username, Config.Password)
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode == 401 {
		return errors.New("Unauthorised")
	}

	if res.StatusCode != 200 {
		b, _ := io.ReadAll(res.Body)
		return errors.New(string(b))
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("error downloading %s", url)
	}

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}

	defer func() {
		if err := out.Close(); err != nil {
			logger.Log().Error(err.Error())
		}
	}()

	// we ignore errors here because the content length is unknown
	// when streaming and ReadAll will return an "unexpected EOF"
	b, _ := io.ReadAll(res.Body)

	if Config.Compress {
		w, err := zstd.NewWriter(out, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
		if err != nil {
			return err
		}
		if _, err := w.Write(b); err != nil {
			return err
		}
		return w.Close()
	}

	// Write the body to file
	_, err = out.Write(b)

	return err
}
