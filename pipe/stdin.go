// yapi
// Copyright (c) 2014 Fatih Cetinkaya (http://github.com/cmfatih/yapi)
// For the full copyright and license information, please view the LICENSE.txt file.

// This file contains stdin related implementations.
//
// References:
//  `ModeNamedPipe`         : http://golang.org/pkg/os/#FileMode
//  `http.DetectContentType`: http://golang.org/src/pkg/net/http/sniff.go
//  File signatures         : http://www.garykessler.net/library/file_sigs.html
//  Media Types             : http://www.iana.org/assignments/media-types/media-types.xhtml

package pipe

import (
	"bufio"
	"errors"
	//"log"
	"net/http"
	"os"
	"runtime"
	"strings"
)

var (
	contentType string
)

func init() {

	// Init vars
	contentType, _ = contentTypeDetect()
}

// ContentType returns the content type
func ContentType() string {
	return contentType
}

// contentTypeDetect detects and returns the content type
func contentTypeDetect() (string, error) {

	// TODO: Add more file signatures

	// *nix
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
			return "", nil
		}
	} else {
		if err != nil {
			return "", errors.New("failed to determine content type: " + err.Error())
		} else if fi.Mode()&os.ModeNamedPipe == 0 && fi.Size() == 0 {
			// no pipe
			return "", nil
		}
	}

	//log.Printf("name: %s - size: %d - mode: %s\n", os.Stdin.Name(), fi.Size(), fi.Mode())

	// Init vars
	ct := ""
	rd := bufio.NewReader(os.Stdin)
	buf, err := rd.Peek(512)
	dct := http.DetectContentType(buf)

	if strings.HasPrefix(dct, "text/") == true {
		dctf := strings.Split(dct, ";")
		ct = dctf[0]
	} else {
		ct = dct
	}

	return ct, nil
}
