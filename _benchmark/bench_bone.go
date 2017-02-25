package main

import (
	"fmt"
	"github.com/go-zoo/bone"
	"net/http"
)

func DynamicHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello %v\n", bone.GetValue(r, "id"))
}

func StaticHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello")
}

func main() {
	mux := bone.New()
	mux.Get("/some/page/:id", http.HandlerFunc(DynamicHandler))
	mux.Get("/other/page/path", http.HandlerFunc(StaticHandler))
	http.ListenAndServe(":8080", mux)
}
