// yapi
// Copyright (c) 2014 Fatih Cetinkaya (http://github.com/cmfatih/yapi)
// For the full copyright and license information, please view the LICENSE.txt file.

// Package stdin provides stdin related functions.
//
// References:
//  `ModeNamedPipe`         : http://golang.org/pkg/os/#FileMode
//  `http.DetectContentType`: http://golang.org/src/pkg/net/http/sniff.go
//  File signatures         : http://www.garykessler.net/library/file_sigs.html
//  Media Types             : http://www.iana.org/assignments/media-types/media-types.xhtml
package stdin

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

var (
	stdinReader    = bufio.NewReader(os.Stdin)
	stdinHasPipe   bool
	stdinErr       error
	contentSize    int64
	contentType    string
	contentTypeErr error
)

// StdinReader returns the stdin reader
func StdinReader() *bufio.Reader {
	return stdinReader
}

// StdinHasPipe returns whether is there a stream on stdin or not
func StdinHasPipe() bool {
	return stdinHasPipe
}

// StdinErr returns the stdin error if any
func StdinErr() error {
	return stdinErr
}

// ContentType returns the content type
func ContentType() string {
	return contentType
}

// contentTypeError returns the content type error
func ContentTypeError() error {
	return contentTypeErr
}

func init() {

	// TODO: Add more file signatures

	// Posix
	//
	// None    : fi.Size(): 0  / fi.Mode(): Dcrw--w----
	// |       : fi.Size(): 0  / fi.Mode(): prw-------
	// < file  : fi.Size(): >0 / fi.Mode(): -rw-rw-r--
	// << EOF  : fi.Size(): >0 / fi.Mode(): -rw-------
	// <<< EOF : fi.Size(): >0 / fi.Mode(): -rw-------
	//
	//
	// Windows
	//
	// None    : `os.Stdin.Stat` error; GetFileInformationByHandle /dev/stdin: The handle is invalid.
	// |       : fi.Size(): >0 / fi.Mode(): -rw-rw-rw-
	// < file  : fi.Size(): >0 / fi.Mode(): -rw-rw-rw-

	fi, err := os.Stdin.Stat()

	if runtime.GOOS == "windows" {
		if err != nil {
			// no pipe
			// if there is no pipe than os.Stdin.Stat() returns error on Windows
			stdinHasPipe = false
		} else {
			stdinHasPipe = true
		}
	} else {
		if err != nil {
			stdinErr = err
			stdinHasPipe = false
		} else if fi.Mode()&os.ModeNamedPipe == 0 && fi.Size() == 0 {
			// no pipe
			stdinHasPipe = false
		} else {
			stdinHasPipe = true
		}
	}

	if stdinHasPipe == true {
		contentSize = fi.Size()
		buf, err := stdinReader.Peek(512)
		//print("buf: ", string(buf[0:10]), "\n") // for debug

		if err != nil && err != io.EOF {
			contentTypeErr = err
		} else {
			dct := http.DetectContentType(buf)

			if strings.HasPrefix(dct, "text/") == true {
				dctf := strings.Split(dct, ";")
				contentType = dctf[0]
			} else {
				contentType = dct
			}
		}
	}

	//print("shp|ct|size: ", stdinHasPipe, "|", contentType, "|", contentSize, "\n") // for debug
}
