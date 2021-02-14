// The view binary links to this binary, passing the filename as the query string.
// Requesting that link will add a like to the post/file. It also displays a page
// detailing what happened, if the file was liked, or if there was some error,
// or if you've already liked it.
//
// Likes for each file are stored <data>/likes/file-name, as just a newline-delimited
// list of IP addresses. The number of likes is the number of newlines, the number of
// IP addresses.
package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/makeworld-the-better-one/gemlikes/shared"
)

func hasLikedAlready(file, ip string) (bool, error) {
	f, err := os.Open(filepath.Join(shared.GetLikesDir(), file))
	if err != nil {
		// Most likely file doesn't exist, no one has liked it then
		return false, nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == ip {
			return true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return true, err
	}
	return false, nil
}

func addLike(file, ip string) error {
	f, err := os.OpenFile(filepath.Join(shared.GetLikesDir(), file), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Save the IP
	_, err = f.WriteString(ip + "\n")
	if err != nil {
		return err
	}
	return nil
}

func main() {
	err := shared.SafeInit()
	shared.HandleErr(err)
	// It's all valid, ready to go
	file, ip, err := shared.GetQueryAndIP()
	shared.HandleErr(err)
	ip = shared.SanitizeIP(ip)

	if !shared.IsFileValid(file) {
		shared.RespondError("File not valid for liking.")
		return
	}

	if shared.LikesDisabled() {
		shared.RespondError("Likes have been disabled.")
		return
	}

	already, err := hasLikedAlready(file, ip)
	shared.HandleErr(err)
	if already {
		shared.Respond("Back so soon? You (or your IP) has already liked this file.")
		return
	}
	// Add a like
	shared.HandleErr(addLike(file, ip))
	shared.Respond("# 💖 Like Added 💖")
}
