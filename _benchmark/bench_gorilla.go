package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func DynamicHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Fprintf(w, "Hello %v\n", vars["id"])
}

func StaticHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello")
}

func main() {
	mux := mux.NewRouter()
	mux.HandleFunc("/some/page/{id}", DynamicHandler).Methods("GET")
	mux.HandleFunc("/other/page/path", StaticHandler).Methods("GET")
	http.ListenAndServe(":8080", mux)
}
