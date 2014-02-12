// Copyright (c) 2013 The artifacts-store AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const StatusUnprocessableEntity = 422

func main() {
	// Parse command line.
	fAddr := flag.String("addr", "localhost:9876", "network addres to listen on")
	fVerbose := flag.Bool("verbose", false, "print verbose output to stderr")

	flag.Parse()

	if flag.NArg() != 1 {
		usage()
	}
	root := flag.Arg(0)

	if !*fVerbose {
		log.SetOutput(ioutil.Discard)
	}

	// Set up the HTTP handler.
	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("%v %v from %v", req.Method, req.URL, req.RemoteAddr)

		if req.Method != "GET" {
			http.Error(rw, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		urlPath := req.URL.Path

		if len(urlPath) == 0 {
			http.Error(rw, "Empty URL Path", StatusUnprocessableEntity)
			return
		}

		if urlPath[len(urlPath)-1] == '/' {
			http.Error(rw, "Trailing Slash in URL Not Allowed", StatusUnprocessableEntity)
			return
		}

		// Handle the regular requests that do not need to return 302.
		if base := path.Base(urlPath); base != "latest" {
			if !strings.HasSuffix(urlPath, ".tar.gz") {
				http.Error(rw, "Not a Tar Archive", StatusUnprocessableEntity)
				return
			}

			file, err := os.Open(filepath.Join(root, urlPath))
			if err != nil {
				if os.IsNotExist(err) {
					http.Error(rw, "Not Found", http.StatusNotFound)
					return
				}
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
			defer file.Close()

			rw.Header().Set("Content-Type", "application/x-gtar-compressed")

			if _, err := io.Copy(rw, file); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

			return
		}

		// If the request path contains "latest" as the base name, we search
		// relevant directory and return 303 to the file modified most recently.
		dir := filepath.Join(root, path.Base(urlPath))
		infos, err := ioutil.ReadDir(dir)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(infos) == 0 {
			http.Error(rw, "Directory Is Empty", http.StatusNotFound)
			return
		}

		var (
			k int
			t time.Time
		)
		for i, info := range infos {
			if info.ModTime().After(t) {
				k = i
			}
		}

		newPath := path.Join(path.Base(urlPath), infos[k].Name())
		http.Redirect(rw, req, newPath, http.StatusSeeOther)
	})

	// Listen and serve until killed.
	if err := http.ListenAndServe(*fAddr, nil); err != nil {
		log.Fatal(err)
	}
}

func usage() {
	io.WriteString(os.Stderr, `USAGE
  artifacts-store [-verbose] [-addr=ADDR] ROOT

OPTIONS
`)
	flag.PrintDefaults()
	io.WriteString(os.Stderr, `
DESCRIPTION
  Start listening for artifacts GET requests using ROOT as the server root.

  This server is just serving static files unless "latest" is present
  as the base name in the URL path. If that is the case, the most recently
  changed file from the relevant directory is returned by sending 303 See Other.

`)
	os.Exit(2)
}
