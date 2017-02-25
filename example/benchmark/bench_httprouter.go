package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func MyHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "Hello %v\n", ps.ByName("id"))
}

func main() {
	mux := httprouter.New()
	mux.GET("/some/page/:id", MyHandler)
	http.ListenAndServe(":8080", mux)
}
