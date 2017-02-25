package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func MyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Fprintf(w, "Hello %v\n", vars["id"])
}

func main() {
	mux := mux.NewRouter()
	mux.HandleFunc("/some/page/{id}", MyHandler).Methods("GET")
	http.ListenAndServe(":8080", mux)
}
