// The view binary is the main one used, it displays comments and likes.
// The file name is passed as the query string.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/icza/backscanner"
	"github.com/makeworld-the-better-one/gemlikes/shared"
)

func numLikes(file string) (int, error) {
	f, err := os.Open(filepath.Join(shared.GetLikesDir(), file))
	if err != nil {
		// Most likely the file hasn't been created yet, meaning no likes.
		return 0, nil
	}
	defer f.Close()

	likes := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		likes++
	}
	if err := scanner.Err(); err != nil {
		return -1, err
	}
	return likes, nil
}

func main() {
	err := shared.SafeInit()
	shared.HandleErr(err)
	file, _, err := shared.GetQueryAndIP()
	shared.HandleErr(err)
	if !shared.IsFileValid(file) {
		shared.RespondError("File not valid, nothing can be shown. Sorry!")
		return
	}

	// Display likes
	likes, err := numLikes(file)
	shared.HandleErr(err)
	likesStr := "likes! 💖"
	if likes == 1 {
		likesStr = "like. 💖"
	}
	likesResponse := fmt.Sprintf("# %s\n\n%d %s\n=> like?%s Add yours\n\n", file, likes, likesStr, shared.PathEscape(file))

	// Display comments

	f, err := os.Open(shared.GetCommentsFile(file))
	if errors.Is(err, os.ErrNotExist) {
		// No comments yet, not a real error
		shared.Respond(likesResponse + fmt.Sprintf("=> add-comment?%s Add a comment 💬\n", shared.PathEscape(file)))
		return
	}
	shared.HandleErr(err)
	defer f.Close()

	fi, err := f.Stat()
	shared.HandleErr(err)

	// Read file in reverse, with a buffer for each 4 lines, to reconstruct comments from most recent to least
	lineNo := 0
	commentsResponse := ""
	buf := ""
	scanner := backscanner.New(f, int(fi.Size()))
	for {
		// Comment file format is: username, id, timestamp, comment
		// Each on their own line, with no empty lines

		line, _, err := scanner.Line()
		if err != nil {
			if err == io.EOF {
				break
			}
			shared.HandleErr(err)
		}
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		switch lineNo % 4 {
		case 0:
			// comment
			// Reset buf
			buf = line + "\n\n"
		case 1:
			// timestamp - RFC1123 format
			buf = fmt.Sprintf("%s:\n", line) + buf
		case 2:
			// id
			buf = fmt.Sprintf("%s) @ ", line) + buf
		case 3:
			// username
			// Final part
			commentsResponse += fmt.Sprintf("%s (id: ", line) + buf
		}
		lineNo++
	}
	// Add number of comments string, before all the comments
	numComments := lineNo / 4
	commentStr := "comments"
	if numComments == 1 {
		commentStr = "comment"
	}
	commentsResponse = fmt.Sprintf("%d %s 💬\n=> add-comment?%s Add yours\n\n", numComments, commentStr, shared.PathEscape(file)) + commentsResponse

	shared.Respond(likesResponse + commentsResponse)
}
