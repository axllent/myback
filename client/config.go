package client

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config struct
var Config struct {
	// type ConfigStruct struct {
	URL      string            `yaml:"url"`
	Repo     string            `yaml:"repo"`
	Compress bool              `yaml:"compress"`
	Username string            `yaml:"user"`
	Password string            `yaml:"pass"`
	Only     []string          `yaml:"only"`
	Ignore   []string          `yaml:"ignore"`
	NoData   []string          `yaml:"nodata"`
	Where    []string          `yaml:"where"`
	WhereMap map[string]string // for key => val lookups
}

// ParseConfig will set the encironment based on a config
func ParseConfig(configFile string) error {

	yamlData, err := ioutil.ReadFile(filepath.Clean(configFile))
	if err != nil {
		return fmt.Errorf("config file not found or readable: \"%s\"", configFile)
	}

	err = yaml.Unmarshal([]byte(yamlData), &Config)
	if err != nil {
		return fmt.Errorf("error parsing yaml config: %v", err)
	}

	Config.WhereMap = make(map[string]string)

	for _, v := range Config.Where {
		s := strings.SplitN(v, " ", 2)
		if len(s) == 2 {
			Config.WhereMap[s[0]] = s[1]
		}
	}

	for _, i := range Config.Only {
		replaced := strings.Replace(i, "*", "~~~~~~WILDCARD~~~~~", -1)
		quoted := regexp.QuoteMeta(replaced)
		quoted = strings.Replace(quoted, "~~~~~~WILDCARD~~~~~", "(.*)", -1)
		onlySlice = append(onlySlice, regexp.MustCompile("^"+quoted+"$"))
	}

	for _, i := range Config.Ignore {
		replaced := strings.Replace(i, "*", "~~~~~~WILDCARD~~~~~", -1)
		quoted := regexp.QuoteMeta(replaced)
		quoted = strings.Replace(quoted, "~~~~~~WILDCARD~~~~~", "(.*)", -1)
		ignoreSlice = append(ignoreSlice, regexp.MustCompile("^"+quoted+"$"))
	}

	for _, i := range Config.NoData {
		replaced := strings.Replace(i, "*", "~~~~~~WILDCARD~~~~~", -1)
		quoted := regexp.QuoteMeta(replaced)
		quoted = strings.Replace(quoted, "~~~~~~WILDCARD~~~~~", "(.*)", -1)
		nodataSlice = append(nodataSlice, regexp.MustCompile("^"+quoted+"$"))
	}

	// add default logging & stats mysql tables to ignore list
	for _, t := range []string{"mysql.*_log", "mysql.help_*", "mysql.*_stats"} {
		replaced := strings.Replace(t, "*", "~~~~~~WILDCARD~~~~~", -1)
		quoted := regexp.QuoteMeta(replaced)
		quoted = strings.Replace(quoted, "~~~~~~WILDCARD~~~~~", "(.*)", -1)
		ignoreSlice = append(ignoreSlice, regexp.MustCompile("^"+quoted+"$"))
	}

	return validConfig()
}

func validConfig() error {
	if Config.URL == "" {
		return errors.New("no server url set in config")
	}

	if Config.Username == "" {
		return errors.New("no MySQL user set in config")
	}

	return nil
}
