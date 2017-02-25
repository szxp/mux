package main

import (
	"fmt"
	"github.com/szxp/mux"
	"net/http"
)

func MyHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello %v\n", r.Context().Value(mux.CtxKey("id")))
}

func main() {
	mux := mux.NewMuxer()
	mux.HandleFunc("/some/page/:id", http.HandlerFunc(MyHandler))
	http.ListenAndServe(":8080", mux)
}
