package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func DynamicHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "Hello %v\n", ps.ByName("id"))
}

func StaticHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "Hello")
}

func main() {
	mux := httprouter.New()
	mux.GET("/some/page/:id", DynamicHandler)
	mux.GET("/other/page/path", StaticHandler)
	http.ListenAndServe(":8080", mux)
}
