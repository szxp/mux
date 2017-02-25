package main

import (
	"fmt"
	"github.com/szxp/mux"
	"net/http"
)

func DynamicHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello %v\n", r.Context().Value(mux.CtxKey("id")))
}

func StaticHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello")
}

func main() {
	mux := mux.NewMuxer()
	mux.HandleFunc("/some/page/:id", http.HandlerFunc(DynamicHandler))
	mux.HandleFunc("/other/page/path", http.HandlerFunc(StaticHandler))
	http.ListenAndServe(":8080", mux)
}
