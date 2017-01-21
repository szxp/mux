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
	assertNotFound(t, m, "GET", "/", 404)
	assertNotFound(t, m, "GET", "/blog", 404)
	assertNotFound(t, m, "GET", "/blog/", 404)
	assertNotFound(t, m, "GET", "/blog/2016", 404)
	assertNotFound(t, m, "GET", "/blog/2016/", 404)
}

func TestWithoutSlash(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	register(m, "/help", nil, "Help without slash", nil)
	assertOK(t, m, "GET", "/help", "Help without slash")
	assertNotFound(t, m, "GET", "/helpMe", 404)
	assertNotFound(t, m, "GET", "/he", 404)
	assertNotFound(t, m, "GET", "/help/", 404)
}

func TestWithSlash(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	register(m, "/login/", nil, "Login with slash", nil)
	assertOK(t, m, "GET", "/login/", "Login with slash")
	assertNotFound(t, m, "GET", "/log", 404)
}

func TestLongPattern(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	register(m, "/category/first", nil, "First category", nil)
	assertNotFound(t, m, "GET", "/cat", 404)
	assertNotFound(t, m, "GET", "/category", 404)
	assertNotFound(t, m, "GET", "/category/", 404)
	assertOK(t, m, "GET", "/category/first", "First category")
	assertNotFound(t, m, "GET", "/category/first/", 404)
	assertNotFound(t, m, "GET", "/category/second/third", 404)
	assertNotFound(t, m, "GET", "/category/second/third/", 404)
	assertNotFound(t, m, "GET", "/category/second/third/2016", 404)
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
	pr := newParamsRecorder(args{"deckId": "123", "cardId": "99"})
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
	assertOK(t, m, "GET", "/new", "New Deck")
	m.Handle("/new", nil)
	assertNotFound(t, m, "GET", "/new", 404)
}

func TestHTTPMethods(t *testing.T) {
	t.Parallel()
	m := NewMuxer()
	register(m, "/", []string{"GET"}, "Home", nil)
	register(m, "/gift", []string{"GET"}, "Receiving a gift", nil)
	register(m, "/gift", []string{"POST"}, "Giving a gift", nil)
	register(m, "/bicycle", []string{"GET", "POST"}, "Bicycle", nil)
	register(m, "/car", []string{"PUT", "PATCH"}, "Car", nil)

	assertOK(t, m, "GET", "/", "Home")
	assertNotFound(t, m, "POST", "/", 405)
	assertOK(t, m, "GET", "/gift", "Receiving a gift")
	assertOK(t, m, "POST", "/gift", "Giving a gift")
	assertOK(t, m, "GET", "/bicycle", "Bicycle")
	assertOK(t, m, "POST", "/bicycle", "Bicycle")
	assertOK(t, m, "PUT", "/car", "Car")
	assertOK(t, m, "PATCH", "/car", "Car")
	assertNotFound(t, m, "PUT", "/bicycle", 405)
	assertNotFound(t, m, "PATCH", "/bicycle", 405)
	assertNotFound(t, m, "GET", "/car", 405)
	assertNotFound(t, m, "POST", "/car", 405)
	assertNotFound(t, m, "DELETE", "/bicycle", 405)
	assertNotFound(t, m, "DELETE", "/car", 405)
}

func register(m *Muxer, pattern string, methods []string, body string, cr *paramsRecorder) {
	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)

		if cr != nil {
			for key := range cr.expected {
				v := r.Context().Value(CtxKey(key))
				cr.actual[key] = v.(string)
			}
		}
	}
	m.HandleFunc(pattern, handlerFunc, methods...)
}

func assertNotFound(t *testing.T, m *Muxer, method, path string, status int) {
	rec := serve(t, m, method, path)

	if rec.Code != status {
		_, _, line, _ := runtime.Caller(1)
		t.Fatalf("expected status %d, but got: %d (line %d)",
			status, rec.Code, line)
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
	expected args
	actual   args
}

func newParamsRecorder(expected args) *paramsRecorder {
	return &paramsRecorder{expected, args{}}
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

func BenchmarkDynamic(b *testing.B) {
	handlerFunc := func(w http.ResponseWriter, r *http.Request) {}
	muxer := NewMuxer()
	muxer.HandleFunc("/a/:b/c/:d/e/:f", handlerFunc, "GET", "POST")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/a/b/c/d/e/f", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		muxer.ServeHTTP(w, r)
	}
}

func BenchmarkStatic(b *testing.B) {
	handlerFunc := func(w http.ResponseWriter, r *http.Request) {}
	muxer := NewMuxer()
	muxer.HandleFunc("/a/b/c/d/e/f", handlerFunc, "GET", "POST")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/a/b/c/d/e/f", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		muxer.ServeHTTP(w, r)
	}
}
