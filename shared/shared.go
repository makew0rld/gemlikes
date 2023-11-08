// This package contains functions useful to two or more of the other binaries/pkgs.
package shared

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml"
)

var ErrConfigDir = errors.New("config dir invalid or not set")
var LikesDisabled = false

var ipSaltValue []byte

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	// Some other error, like perms for example
	return false
}

// SafeInit makes sure all the required data directories are created.
// It can be called multiple times safely.
//
// It should be called and checked before doing any other operation.
func SafeInit() error {
	// Create required files and dirs if they don't exist
	// err := os.MkdirAll(getConfigDir(), 0755)
	// if err != nil {
	// 	return err
	// }
	err := os.MkdirAll(GetLikesDir(), 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(GetCommentsDir(), 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(GetTmpDir(), 0755)
	if err != nil {
		return err
	}
	os.OpenFile(GetConfigPath(), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	// Check config parsing
	config, err := toml.LoadFile(GetConfigPath())
	if err != nil {
		return err
	}
	// Validate config "dir" variable
	pathsNil := config.Get("dirs")
	if pathsNil == nil {
		return ErrConfigDir
	}
	paths := config.Get("dirs").([]interface{})
	if len(paths) == 0 {
		// var not set, or empty
		return ErrConfigDir
	}
	for _, path := range paths {
		if !PathExists(path.(string)) {
			return ErrConfigDir
		}
	}
	// Validate data dir variable
	data := config.Get("data")
	if data == nil {
		return ErrConfigDir
	}
	if !PathExists(data.(string)) {
		return ErrConfigDir
	}
	// Check CGI vars
	file, ok := os.LookupEnv("QUERY_STRING")
	if !ok || strings.TrimSpace(file) == "" {
		// No query
		return errors.New("no file specified")
	}
	ip, ok := os.LookupEnv("REMOTE_ADDR")
	if !ok || strings.TrimSpace(ip) == "" {
		// No remote addr
		return errors.New("no REMOTE_ADDR specified, server CGI error")
	}
	// Get config "disable_likes" variable
	if config.Get("disable_likes") != nil {
		LikesDisabled = config.Get("disable_likes").(bool)
	}
	// Load IP->ID salt
	ipSaltValue, err = loadIPSalt(config, data.(string))
	if err != nil {
		return err
	}

	return nil
}

func GetQueryAndIP() (string, string, error) {
	query, err := url.PathUnescape(os.Getenv("QUERY_STRING"))
	if err != nil {
		return "", "", errors.New("your client might not be escaping query strings properly")
	}
	return query, os.Getenv("REMOTE_ADDR"), nil
}

// IsFileValid returns a true if that file is actionable.
// ie, it can be commented on and liked.
func IsFileValid(file string) bool {
	if strings.TrimSpace(file) == "" {
		return false
	}
	if strings.Contains(file, "?") {
		// For query string reasons
		return false
	}
	// Check if it's actually a path and not a file
	// Because queries like `dir/file.gmi` are not allowed.
	if filepath.Base(file) != file {
		// On relative files with no directory, filepath.Base just returns
		// the filename again. Therefore if anything different is returned,
		// the filepath contains some directories and is invalid.
		return false
	}
	// Check if it's a malicious, directory-traversing, path
	// Like: ../../otherfile.gmi
	// This check is just in case - in theory no malicious path
	// should be able to get past the previous check
	if filepath.Clean(string(os.PathSeparator)+file) != string(os.PathSeparator)+file {
		// The Clean func will remove all relative path tricks
		// If the result of cleaning is not the same as the uncleaned path, then there are definitely some directory
		// tricks in the file passed
		return false

	}

	config, _ := toml.LoadFile(GetConfigPath())
	dirs := config.Get("dirs").([]interface{})
	found := 0 // How many times it's found
	for _, dir := range dirs {
		if PathExists(filepath.Join(dir.(string), file)) {
			found++
		}
	}
	// If found != 1 then it wasn't found, or there are multiple files with that name
	return found == 1
}

func Respond(gmi string) {
	fmt.Print("20 text/gemini\r\n")
	fmt.Print(gmi + "\r\n")
}

func RespondError(reason string) {
	fmt.Print("40 " + strings.TrimSpace(reason) + "\r\n")
}

func RespondInput(prompt string) {
	fmt.Print("10 " + strings.TrimSpace(prompt) + "\r\n")
}

func getDataDir() string {
	config, _ := toml.LoadFile(GetConfigPath())
	return config.Get("data").(string)
}

func GetLikesDir() string {
	return filepath.Join(getDataDir(), "likes")
}

func GetCommentsDir() string {
	return filepath.Join(getDataDir(), "comments")
}

func GetTmpDir() string {
	return filepath.Join(getDataDir(), "tmp")
}

func WriteIPSalt(w io.Writer) (int, error) {
	return w.Write(ipSaltValue)
}

func loadIPSalt(config *toml.Tree, dataDir string) ([]byte, error) {
	ipSaltMethod := "auto"
	if key := config.Get("ip_salt"); key != nil {
		ipSaltMethod = key.(string)
	}

	ipSaltPath := ipSaltMethod
	switch ipSaltMethod {
	case "disabled":
		return nil, nil
	case "auto":
		ipSaltPath = filepath.Join(dataDir, "ip_salt")
		f, err := os.OpenFile(ipSaltPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o644)
		if err == nil {
			// First time, create random data
			defer f.Close()
			if _, err := io.CopyN(f, rand.Reader, 16); err != nil {
				return nil, err
			}
		}
	}

	return ioutil.ReadFile(ipSaltPath)
}

func getConfigDir() string {
	e, _ := os.Executable()
	return filepath.Dir(e)
}

func GetConfigPath() string {
	return filepath.Join(getConfigDir(), "gemlikes.toml")
}

func GetCommentsFile(file string) string {
	return filepath.Join(GetCommentsDir(), file, "comments")
}

// SanitizeIP makes IPs filename safe.
// It changes IPv6 colons to underscores.
func SanitizeIP(ip string) string {
	return strings.ReplaceAll(ip, ":", "_")
}

func HandleErr(err error) {
	if err == nil {
		return
	}
	text := err.Error()
	// Capitalize first letter
	text = strings.ToUpper(string(text[0])) + text[1:]
	RespondError(text)
	os.Exit(0)
}

func PathEscape(path string) string {
	return strings.ReplaceAll(url.PathEscape(path), "+", "%2B")
}
