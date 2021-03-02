// The view binary links to this binary, passing the filename as the query string.
// The binary stores the filename in a file at <data>/tmp/<ip>, to use later. <- TODO
// Requesting that link will return a response asking for input.
// That's where the user enters their username and comment.
// Their username is just the first word of their comment.
// It checks for name re-use, and if that IP has already hit the max comments amount.
// It knows what file the comment is for by looking at the tmp file from earlier
// Comment is stored and the tmp file is deleted.
// Usernames are stored in this file, one per line: <data>/comments/<file>/<ip>
// All comments are stored by time in <data>/comments/<file>/comments
// Lines alternate between username, id, timestamp, and comment.
// The id is a truncated hash of the user's IP address, and the timestamp is in RFC1123 format.
//
// Colons are replaced with underscores for IPv6 IPs.
//
// Issues:
// Filenames and comments are both used as query strings here. Currently the only
// way to tell the difference is to just assume that an invalid filename is a username.
// This is a hack.
package main

import (
	"bufio"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/makeworld-the-better-one/gemlikes/shared"
	toml "github.com/pelletier/go-toml"
)

func nameTaken(file, name, ip string) (bool, error) {
	// Get file list
	dir := filepath.Join(shared.GetCommentsDir(), file)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return true, err
	}
	d, _ := os.Open(dir)
	ips, err := d.Readdirnames(0)
	d.Close()
	if err != nil {
		return true, err
	}

	// Search each file to see if the username is taken
	// TODO: Parallelize this using goroutines
	for _, ipfile := range ips {
		if ip == ipfile {
			// Don't search your own comment file to see if the name is taken
			// An IP address can reuse a name it's already used
			continue
		}
		if ipfile == "comments" {
			// General comments file, shouldn't be searched
			continue
		}

		f, err := os.Open(filepath.Join(dir, ipfile))
		if err != nil {
			return true, errors.New("error opening " + ipfile + " in nameTaken func: " + err.Error())
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if name == strings.TrimSpace(scanner.Text()) {
				f.Close()
				return true, nil
			}
		}
		if err := scanner.Err(); err != nil {
			f.Close()
			return true, err
		}
		f.Close()
	}
	// The name wasn't found to be used for that file, by any non-self IP addresses
	return false, nil
}

func maxComments() int64 {
	config, _ := toml.LoadFile(shared.GetConfigPath())
	max := config.Get("max_comments").(int64)
	if max == 0 {
		// Missing or not an int, etc
		return 5 // Default value
	}
	if max == -1 {
		// Comments disabled
		return 0
	}
	return max
}

func hasCommentsLeft(file, ip string) (bool, error) {
	max := maxComments()
	if max == 0 {
		// No comments allowed
		return false, nil
	}
	f, err := os.Open(filepath.Join(shared.GetCommentsDir(), file, ip))
	if errors.Is(err, os.ErrNotExist) {
		// They haven't commented before
		return true, nil
	} else if err != nil {
		return false, errors.New("error opening " + ip + " in hasCommentsLeft func: " + err.Error())
	}
	defer f.Close()
	// Count usernames in file to determine number of comments already made
	scanner := bufio.NewScanner(f)
	lines := 0
	for scanner.Scan() {
		lines++
	}
	if err := scanner.Err(); err != nil {
		return false, err
	}

	if int64(lines) >= max {
		return false, nil
	}
	return true, nil
}

func storeFilename(file, ip string) error {
	// File is overwritten every time
	f, err := os.OpenFile(filepath.Join(shared.GetTmpDir(), ip), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(file)
	if err != nil {
		return err
	}
	return nil
}

func getFilename(ip string) (string, error) {
	path := filepath.Join(shared.GetTmpDir(), ip)
	if !shared.PathExists(path) {
		return "", errors.New("file not known")
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	file := string(bytes)
	if !shared.IsFileValid(file) {
		return "", errors.New("stored filename is invalid")
	}
	return file, nil
}

func removeTmp(ip string) {
	os.Remove(filepath.Join(shared.GetTmpDir(), ip))
}

func getId(ip string) string {
	ip = strings.ReplaceAll(ip, "_", ":") // Use real IP address, not sanitized
	h := sha256.New()
	h.Write([]byte(ip))
	shared.WriteIPSalt(h)
	// First 8 chars
	return fmt.Sprintf("%x", h.Sum(nil))[:8]
}

func addComment(file, ip, username, comment string) error {
	// Save username to username list
	f, err := os.OpenFile(filepath.Join(shared.GetCommentsDir(), file, ip), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	_, err = f.WriteString(username + "\n")
	if err != nil {
		f.Close()
		return err
	}
	f.Close()

	// Save username, id, timestamp, and comment - each on their own line
	f, err = os.OpenFile(shared.GetCommentsFile(file), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(username + "\n" + getId(ip) + "\n" + time.Now().UTC().Round(time.Second).Format(time.RFC1123) + "\n" + comment + "\n")
	if err != nil {
		return err
	}
	return nil
}

func main() {
	err := shared.SafeInit()
	shared.HandleErr(err)

	query, ip, err := shared.GetQueryAndIP()
	shared.HandleErr(err)
	ip = shared.SanitizeIP(ip)

	if shared.IsFileValid(query) {
		// It's a filename to comment on
		hasLeft, err := hasCommentsLeft(query, ip)
		shared.HandleErr(err)
		if !hasLeft {
			shared.RespondError(fmt.Sprintf("You have already commented %d times on this file.", maxComments()))
			return
		}
		// Store filename in temp file
		shared.HandleErr(storeFilename(query, ip))
		shared.RespondInput("Enter your username, a space, and your comment:")
		return
	}

	// It's a comment, this is a second request

	file, err := getFilename(ip)
	shared.HandleErr(err)
	removeTmp(ip)

	query = strings.TrimSpace(query)
	if len(query) < 3 {
		// 3 chars is enough for a username, space, and one-char comment
		shared.RespondError("Your input was too short.")
		return
	}
	// Get username, by finding the first space
	idx := strings.Index(query, " ")
	if idx <= -1 {
		shared.RespondError("No username found. Your username should be the first part of your input, followed by a space, and then your comment text.")
		return
	}
	// Validate username, comment, etc
	username := strings.TrimSpace(query[:idx])
	if len([]rune(username)) > 40 || len([]rune(username)) <= 0 { // TODO: Make this configurable?
		shared.RespondError("Your username is invalid. It must be <= 40 characters and > 0.")
		return
	}
	if strings.Contains(username, "?") || strings.Contains(username, "%") || strings.Contains(username, " ") {
		shared.RespondError("Your username must not contain a question mark, a percent sign, or any spaces.")
		return
	}
	comment := strings.TrimSpace(query[idx+1:])
	if len(comment) > 2000 {
		shared.RespondError("Comment must be under 2000 bytes.")
		return
	}
	if len(comment) <= 0 {
		shared.RespondError("Your comment was empty, it was not added.")
		return
	}

	// Remove newlines
	comment = strings.ReplaceAll(comment, "\r\n", " ")
	comment = strings.ReplaceAll(comment, "\n", " ")

	taken, err := nameTaken(file, username, ip)
	shared.HandleErr(err)
	if taken {
		shared.RespondError("A different IP address has already used that username on this article.")
		return
	}

	shared.HandleErr(addComment(file, ip, username, comment))
	shared.Respond(fmt.Sprintf("Comment by '%s' on '%s' added!\n=> view?%s View all comments", username, file, shared.PathEscape(file)))
}
