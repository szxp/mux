[![Build Status](https://travis-ci.org/szxp/mux.svg?branch=master)](https://travis-ci.org/szxp/mux)
[![Build Status](https://ci.appveyor.com/api/projects/status/github/szxp/mux?branch=master&svg=true)](https://ci.appveyor.com/project/szxp/mux)
[![GoDoc](https://godoc.org/github.com/szxp/mux?status.svg)](https://godoc.org/github.com/szxp/mux)
[![Go Report Card](https://goreportcard.com/badge/github.com/szxp/mux)](https://goreportcard.com/report/github.com/szxp/mux)

# mux
A lightweight HTTP request router (multiplexer). [Documentation is available at GoDoc](https://godoc.org/github.com/szxp/mux).

## Releases
Master branch is the stable production-ready branch. 

## Features
 * Static and dynamic patterns supported. Dynamic parameter values are available in the [request's context](https://godoc.org/net/http#Request.Context).
 * Compatible with the built-in [http.Handler](https://godoc.org/net/http#Handler)
 * Only standard library dependencies.
 * Go 1.7+ supported.
 
## Benchmarks
Testing the examples in the benchmark directory running the `bench.sh` script at least
three times for each muxer. The result is:

```
              dynamic route (/some/page/:id)

httprouter    27229 Requests/sec  (+13.40 %)
bone          25679 Requests/sec  ( +6.95 %)
mux           25439 Requests/sec  ( +5.95 %)
gorrila/mux   24010 Requests/sec  (     0 %)
```

```
              static route (/other/page/path)

httprouter    27780 Requests/sec  (+10.71 %)
bone          27357 Requests/sec  ( +9,03 %)
mux           26792 Requests/sec  ( +6.77 %)
gorrila/mux   25091 Requests/sec  (     0 %)
```

The test machine was a Dell Latitude D630 laptop with Intel(R) Core2 Duo T7250 2.00 GHz processor.

## Example
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
	muxer.HandleFunc("/login", loginHandler, "POST")
	muxer.HandleFunc("/users/:username", userHandler)
	muxer.NotFound(notFoundHandler)
	http.ListenAndServe(":8080", muxer)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	body := []byte(`
		<h1>Home</h1>
		<p>
			<a href="/users/admin">Admin profile page</a> <br/>
			<a href="/login">Login page</a>
		</p>
		<form action="/" method="POST">
			<button type="submit">Post to Home URL</button>
		</form>
		<form action="/nonexisting" method="POST">
			<button type="submit">Post to non existing URL</button>
		</form>
	`)
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.Header().Add("Content-Length", fmt.Sprintf("%d", len(body)))
	w.Write(body)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Login")
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, r.Context().Value(mux.CtxKey("username")))
}

func notFoundHandler(w http.ResponseWriter, r *http.Request, methodMismatch bool) {
	if methodMismatch {
		http.Error(w, r.Method+" not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.Error(w, "not found", http.StatusNotFound)
}
```
