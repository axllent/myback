package server

import (
	"crypto/sha256"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"
)

// CharacterReplacements map for URL & file friendly MySQL database & table names
var characterReplacements = map[string]string{
	"'":  "@0027",
	"/":  "@002F",
	"\\": "@005C",
	" ":  "@0020",
	"\"": "@0022",
	"*":  "@002A",
	"+":  "@002B",
	",":  "@002C",
	".":  "@002E",
	":":  "@003A",
	";":  "@003B",
	"%":  "@0025",
	"#":  "@0023",
	"?":  "@003f",
	"`":  "@0060",
}

// EncodeName encodes a database or table name
func encodeName(str string) string {
	// change any @
	str = strings.Replace(str, "@", "@0040", -1)
	for key, value := range characterReplacements {
		str = strings.Replace(str, key, value, -1)
	}
	return str
}

// DecodeName decodes a database table name
func decodeName(str string) string {
	for key, value := range characterReplacements {
		str = strings.Replace(str, value, key, -1)
	}
	// change back any @
	str = strings.Replace(str, "@0040", "@", -1)
	return str
}

// EscapeBackticks escapes ` in the database / table name
func escapeBackticks(str string) string {
	str = strings.Replace(str, "`", "``", -1)

	return str
}

// IsValidBool returns whether a string matched valid bool results
func isValidBool(val string) bool {
	val = strings.ToLower(val)
	valid := []string{"true", "false", "0", "1"}
	for _, n := range valid {
		if val == n {
			return true
		}
	}
	return false
}

// hashString will return a sha256 hash of a string
func hashString(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

// Which locates a binary in the current $PATH.
// It will append ".exe" to the filename if the platform is Windows.
func which(binName string) (string, error) {
	return exec.LookPath(binName)
}

func ip(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return ip
	}

	return "unknown"
}
