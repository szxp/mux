[![Build Status](https://travis-ci.org/szxp/mux.svg?branch=master)](https://travis-ci.org/szxp/mux)
[![Build Status](https://ci.appveyor.com/api/projects/status/github/szxp/mux?branch=master&svg=true)](https://ci.appveyor.com/project/szxp/mux)
[![GoDoc](https://godoc.org/github.com/szxp/mux?status.svg)](https://godoc.org/github.com/szxp/mux)
[![Go Report Card](https://goreportcard.com/badge/github.com/szxp/mux)](https://goreportcard.com/report/github.com/szxp/mux)

# mux
A lightweight HTTP request router (multiplexer). [Documentation is available at GoDoc](https://godoc.org/github.com/szxp/mux).

## Features
 * Static and dynamic patterns supported.
 * Dynamic parameter values are available in the request's context.
 * The router implements the http.Handler interface, so the standard library's HTTP server can use it as a handler (see the example below).
 * Safe for concurrent use by multiple goroutines.
 * Go 1.7+ supported.

## Working exmaple
```go
package main

import (
	"fmt"
	"github.com/szxp/mux"
	"net/http"
)

func main() {
	muxer := mux.NewMuxer()
	muxer.HandleFunc("/", indexHandler, "GET")
	muxer.HandleFunc("/login", loginHandler, "GET", "POST")
	muxer.HandleFunc("/users/:username", userHandler)
	http.ListenAndServe(":8080", muxer)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Home")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s %s", r.Method, "Login")
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, r.Context().Value(mux.CtxKey("username")))
}
```
