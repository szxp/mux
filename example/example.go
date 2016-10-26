package main

import (
	"fmt"
	"github.com/szxp/mux"
	"net/http"
)

func main() {
	muxer := mux.NewMuxer()
	muxer.HandleFunc("/login", loginHandler)
	muxer.HandleFunc("/users/:username", userHandler)
	http.ListenAndServe(":8080", muxer)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Login")
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, r.Context().Value("username"))
}
