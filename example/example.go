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
