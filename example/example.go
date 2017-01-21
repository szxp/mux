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
