// Copyright (c) 2013 The latest-artifact AUTHORS
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
	"time"
)

func main() {
	// Parse the command line.
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

		// Only GET method is allowed.
		if req.Method != "GET" {
			http.Error(rw, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		urlPath := req.URL.Path

		// Only pseudo-files called "latest" can be requested.
		if base := path.Base(urlPath); base != "latest" {
			http.Error(rw, "Filename Not Allowed", http.StatusForbidden)
			return
		}

		// Find the most recently modified file.
		infos, err := ioutil.ReadDir(filepath.Join(root, path.Dir(urlPath)))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(infos) == 0 {
			http.Error(rw, "Directory Empty", http.StatusNotFound)
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

		// Return a redirect.
		newPath := path.Join(path.Base(urlPath), infos[k].Name())
		http.Redirect(rw, req, newPath, http.StatusTemporaryRedirect)
	})

	// Listen and serve until killed.
	if err := http.ListenAndServe(*fAddr, nil); err != nil {
		log.Fatal(err)
	}
}

func usage() {
	io.WriteString(os.Stderr, `USAGE
  latest-artifact [-verbose] [-addr=ADDR] ROOT

OPTIONS
`)
	flag.PrintDefaults()
	io.WriteString(os.Stderr, `
DESCRIPTION
  Start listening for artifacts GET requests using ROOT as the server root.

  This server does not serve any files itself, but when hit with a GET request
  for a pseudo-file called "latest", it returns a redirect to the most recently
  modified file in the relevant directory.

`)
	os.Exit(2)
}
