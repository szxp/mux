package mux

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
)

func TestEmpty(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	assertNotFound(t, m, "GET", "/")
	assertNotFound(t, m, "GET", "/blog")
	assertNotFound(t, m, "GET", "/blog/")
	assertNotFound(t, m, "GET", "/blog/2016")
	assertNotFound(t, m, "GET", "/blog/2016/")
}

func TestCatchAll(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	register(m, "/", nil, "Catch all", nil)
	assertOK(t, m, "GET", "/", "Catch all")
	assertOK(t, m, "GET", "/not-registered", "Catch all")
	assertOK(t, m, "GET", "/a/b", "Catch all")
}

func TestWithoutSlash(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	register(m, "/help", nil, "Help without slash", nil)
	assertOK(t, m, "GET", "/help", "Help without slash")
	assertNotFound(t, m, "GET", "/helpMe")
	assertNotFound(t, m, "GET", "/he")
	assertNotFound(t, m, "GET", "/help/")
}

func TestWithSlash(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	register(m, "/login/", nil, "Login with slash", nil)
	assertOK(t, m, "GET", "/login/", "Login with slash")
	assertNotFound(t, m, "GET", "/log")
}

func TestLongPattern(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	register(m, "/category/first", nil, "First category", nil)
	assertNotFound(t, m, "GET", "/cat")
	assertNotFound(t, m, "GET", "/category")
	assertNotFound(t, m, "GET", "/category/")
	assertOK(t, m, "GET", "/category/first", "First category")
	assertNotFound(t, m, "GET", "/category/first/")
	assertNotFound(t, m, "GET", "/category/second/third")
	assertNotFound(t, m, "GET", "/category/second/third/")
	assertNotFound(t, m, "GET", "/category/second/third/2016")
}

func TestShorterCatchAll(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	register(m, "/users/", nil, "Users", nil)
	register(m, "/users/admin", nil, "Admin", nil)
	assertOK(t, m, "GET", "/users/", "Users")
	assertOK(t, m, "GET", "/users/admin", "Admin")
	assertOK(t, m, "GET", "/users/administrator", "Users")
	assertOK(t, m, "GET", "/users/admin/", "Users")
	assertOK(t, m, "GET", "/users/admin/details", "Users")
	assertOK(t, m, "GET", "/users/admin/details/", "Users")
}

func TestDynamicPattern(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	pr := newParamsRecorder(params{"deckId": "123", "cardId": "99"})
	register(m, "/", nil, "Home", nil)
	register(m, "/new", nil, "New Deck", nil)
	register(m, "/:deckId/study/:cardId", nil, "Deck", pr)
	assertOK(t, m, "GET", "/123/study/99", "Deck")
	pr.assertEquals(t)
}

func TestRegisterPatternTwice(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	register(m, "/new", nil, "Home1", nil)
	register(m, "/new", nil, "Home2", nil)
	assertOK(t, m, "GET", "/new", "Home2")
}

func TestRemoveHandler(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	register(m, "/", nil, "Home", nil)
	register(m, "/new", nil, "New Deck", nil)
	m.Handle("/new", nil)
	assertOK(t, m, "GET", "/new", "Home")
}

func TestMethods(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	register(m, "/gift", []string{"GET"}, "Receiving a gift", nil)
	register(m, "/gift", []string{"POST"}, "Giving a gift", nil)
	register(m, "/bicycle", []string{"GET", "POST"}, "Bicycle", nil)
	register(m, "/car", []string{"PUT", "PATCH"}, "Car", nil)

	assertOK(t, m, "GET", "/gift", "Receiving a gift")
	assertOK(t, m, "POST", "/gift", "Giving a gift")
	assertOK(t, m, "GET", "/bicycle", "Bicycle")
	assertOK(t, m, "POST", "/bicycle", "Bicycle")
	assertOK(t, m, "PUT", "/car", "Car")
	assertOK(t, m, "PATCH", "/car", "Car")
	assertNotFound(t, m, "PUT", "/bicycle")
	assertNotFound(t, m, "PATCH", "/bicycle")
	assertNotFound(t, m, "GET", "/car")
	assertNotFound(t, m, "POST", "/car")
	assertNotFound(t, m, "DELETE", "/bicycle")
	assertNotFound(t, m, "DELETE", "/car")
}

func register(m *Muxer, pattern string, methods []string, body string, cr *paramsRecorder) {
	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)

		if cr != nil {
			for key := range cr.expected {
				v := r.Context().Value(key)
				cr.actual[key] = v.(string)
			}
		}
	}
	m.HandleFunc(pattern, handlerFunc, methods...)
}

func assertNotFound(t *testing.T, m *Muxer, method, path string) {
	rec := serve(t, m, method, path)

	if rec.Code != http.StatusNotFound {
		_, _, line, _ := runtime.Caller(1)
		t.Fatalf("expected code %d, but got: %d (line %d)",
			http.StatusNotFound, rec.Code, line)
	}
}

func assertOK(t *testing.T, m *Muxer, method, path, body string) {
	rec := serve(t, m, method, path)

	if rec.Code != http.StatusOK {
		_, _, line, _ := runtime.Caller(1)
		t.Fatalf("expected code %d, but got: %d (line %d)",
			http.StatusOK, rec.Code, line)
	}

	if rec.Body.String() != body {
		_, _, line, _ := runtime.Caller(1)
		t.Fatalf("expected body '%s', but got: '%s' (line %d)",
			body, rec.Body.String(), line)
	}
}

func serve(t *testing.T, m *Muxer, method, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	m.ServeHTTP(w, httptest.NewRequest(method, path, nil))
	return w
}

type paramsRecorder struct {
	expected params
	actual   params
}

func newParamsRecorder(expected params) *paramsRecorder {
	return &paramsRecorder{expected, params{}}
}

func (pr *paramsRecorder) assertEquals(t *testing.T) {
	if len(pr.expected) != len(pr.actual) {
		_, _, line, _ := runtime.Caller(1)
		t.Fatalf("expected params: %v, but got: %v (line: %d)",
			pr.expected, pr.actual, line)
	}

	for key := range pr.expected {
		if pr.expected[key] != pr.actual[key] {
			_, _, line, _ := runtime.Caller(1)
			t.Fatalf("expected params: %v, but got: %v (line: %d)",
				pr.expected, pr.actual, line)
		}
	}
}
