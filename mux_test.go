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
	assertNotFound(t, m, "/")
	assertNotFound(t, m, "/blog")
	assertNotFound(t, m, "/blog/")
	assertNotFound(t, m, "/blog/2016")
	assertNotFound(t, m, "/blog/2016/")
}

func TestCatchAll(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	register(m, "/", "Catch all", nil)
	assertOK(t, m, "/", "Catch all")
	assertOK(t, m, "/not-registered", "Catch all")
	assertOK(t, m, "/a/b", "Catch all")
}

func TestWithoutSlash(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	register(m, "/help", "Help without slash", nil)
	assertOK(t, m, "/help", "Help without slash")
	assertNotFound(t, m, "/helpMe")
	assertNotFound(t, m, "/he")
	assertNotFound(t, m, "/help/")
}

func TestWithSlash(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	register(m, "/login/", "Login with slash", nil)
	assertOK(t, m, "/login/", "Login with slash")
	assertNotFound(t, m, "/log")
}

func TestLongPattern(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	register(m, "/category/first", "First category", nil)
	assertNotFound(t, m, "/cat")
	assertNotFound(t, m, "/category")
	assertNotFound(t, m, "/category/")
	assertOK(t, m, "/category/first", "First category")
	assertNotFound(t, m, "/category/first/")
	assertNotFound(t, m, "/category/second/third")
	assertNotFound(t, m, "/category/second/third/")
	assertNotFound(t, m, "/category/second/third/2016")
}

func TestShorterCatchAll(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	register(m, "/users/", "Users", nil)
	register(m, "/users/admin", "Admin", nil)
	assertOK(t, m, "/users/", "Users")
	assertOK(t, m, "/users/admin", "Admin")
	assertOK(t, m, "/users/administrator", "Users")
	assertOK(t, m, "/users/admin/", "Users")
	assertOK(t, m, "/users/admin/details", "Users")
	assertOK(t, m, "/users/admin/details/", "Users")
}

func TestDynamicPattern(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	pr := newParamsRecorder(params{"deckId": "123", "cardId": "99"})
	register(m, "/", "Home", nil)
	register(m, "/new", "New Deck", nil)
	register(m, "/:deckId/study/:cardId", "Deck", pr)
	assertOK(t, m, "/123/study/99", "Deck")
	pr.assertEquals(t)
}

func TestRegisterPatternTwice(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	register(m, "/new", "Home1", nil)
	register(m, "/new", "Home2", nil)
	assertOK(t, m, "/new", "Home2")
}

func TestRemoveHandler(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	register(m, "/", "Home", nil)
	register(m, "/new", "New Deck", nil)
	m.Handle("/new", nil)
	assertOK(t, m, "/new", "Home")
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

func register(m *Muxer, pattern string, body string, cr *paramsRecorder) {
	m.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)

		if cr != nil {
			for key := range cr.expected {
				v := r.Context().Value(key)
				cr.actual[key] = v.(string)
			}
		}
	})
}

func assertNotFound(t *testing.T, m *Muxer, path string) {
	rec := serve(t, m, path)

	if rec.Code != http.StatusNotFound {
		_, _, line, _ := runtime.Caller(1)
		t.Fatalf("expected code %d, but got: %d (line %d)",
			http.StatusNotFound, rec.Code, line)
	}
}

func assertOK(t *testing.T, m *Muxer, path string, body string) {
	rec := serve(t, m, path)

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

func serve(t *testing.T, m *Muxer, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	m.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
	return w
}
