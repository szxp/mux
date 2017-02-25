// Copyright 2017 Péter Szakszon. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package mux provides a lightweight HTTP request router (multiplexer).
package mux

import (
	"context"
	"net/http"
	"path"
	"sort"
	"strings"
	"sync"
)

// Muxer represents an HTTP request multiplexer.
type Muxer struct {
	mu              sync.RWMutex
	registered      map[string]*route
	routes          []*route
	notFoundHandler func(w http.ResponseWriter, r *http.Request, methodMismatch bool)
}

// NewMuxer returns a new Muxer.
// The returned Muxer is safe for concurrent use by multiple goroutines.
func NewMuxer() *Muxer {
	return &Muxer{
		registered: make(map[string]*route, 10),
		routes:     make([]*route, 0, 10),
	}
}

// route represents a pattern with handlers.
type route struct {
	// the exploded pattern
	segments []string

	// the length of segments slice
	len int

	// supported method
	method string

	// paramateres names: segment index -> name
	params map[int]string

	// the handler for a pattern that ends in a slash
	slashHandler http.Handler

	// the handler for a pattern that NOT ends in a slash
	nonSlashHandler http.Handler
}

// methodSupported checks whether the given method
// is supported by this route.
func (p *route) methodSupported(method string) bool {
	return p.method == "" || p.method == method
}

// notMatch checks whether the segment at index i
// does not match the pathSeg path segment.
func (p *route) notMatch(pathSeg string, i int) bool {
	if /*p.len == 0 || */ p.len-1 < i {
		return false
	}

	s := p.segments[i]
	return (len(s) == 0 || s[0] != ':') && (s != pathSeg)
}

// args is a map for request parameter values.
type args map[string]string

// argsMap returns a map containing request parameter values.
func (p *route) argsMap(pathSegs []string) args {
	m := args{}
	slen := len(pathSegs)
	for i, name := range p.params {
		if i < slen {
			m[name] = pathSegs[i]
		}
	}
	return m
}

// priority computes the priority of the route.
//
// Every segment has a priority value:
// 2 = static segment
// 1 = dynamic segment
//
// The route priority is created by concatenating the priorities of the segments.
// The slash (/) route has the priority 0.
func (p *route) priority() string {
	if p.segments[0] == "" { // slash pattern
		return "0"
	}
	pri := make([]byte, 0, 3)
	for _, s := range p.segments {
		if s[0] == ':' {
			pri = append(pri, '1')
		} else {
			pri = append(pri, '2')
		}
	}
	return string(pri)
}

// byPriority implements sort.Interface for []*route based on
// the priority().
type byPriority []*route

func (a byPriority) Len() int           { return len(a) }
func (a byPriority) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byPriority) Less(i, j int) bool { return a[i].priority() > a[j].priority() }

// Handle registers the handler for the given pattern.
//
// Static and dynamic patterns are supported.
// Static pattern examples:
//   /
//   /product
//   /users/new/
//
// Dynamic patterns can contain paramterer names after the colon character.
// Dynamic pattern examples:
//   /blog/:year/:month
//   /users/:username/profile
//
// Parameter values for a dynamic pattern will be available
// in the request's context (http.Request.Context()) associated with
// the parameter name. Use the context's Value() method to retrieve a value:
//   value := req.Context().Value(mux.CtxKey("username")))
//
// The muxer will choose the most specific pattern that matches the request.
// A pattern with longer static prefix is more specific
// than a pattern with a shorter static prefix.
//
// If the request path is /a/e then:
//   /a      vs /:b       => /a       wins
//   /:x     vs /:x/e     => /:x/e    wins
//   /a/:b/c vs /:d/e/:f  => /a/:b/c  wins
//
// The slash pattern (/) does NOT act as a catch all pattern.
//
// If HTTP methods are given then only requests with those methods
// will be dispatched to the handler whose pattern matches the request path. For example:
//   muxer.HandleFunc("/login", loginHandler, "GET", "POST")
//
// If the Muxer didn't find a suitable handler for the request
// and "not found" handler is not set the Muxer will reply to the request
// with "404 Not found" or "405 Method not allowed" status code.
// Use the NotFound method to set a custom error handler.
func (m *Muxer) Handle(pattern string, handler http.Handler, methods ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if pattern == "" {
		panic("invalid pattern " + pattern)
	}

	host, path := split(pattern)
	endsInSlash := path[len(path)-1] == '/'
	path = strings.Trim(path, "/")

	if len(methods) == 0 {
		methods = []string{""}
	}
	for _, method := range methods {
		key := method + host + path
		r := m.registered[key]
		if r == nil {
			r = newRoute(method, path)
			m.routes = append(m.routes, r)
			m.registered[key] = r
		}

		if endsInSlash {
			r.slashHandler = handler
		} else {
			r.nonSlashHandler = handler
		}
	}
	sort.Sort(byPriority(m.routes))
}

func newRoute(method, path string) *route {
	r := &route{method: method}
	r.segments = strings.Split(path, "/")
	r.len = len(r.segments)

	for i, s := range r.segments {
		if len(s) > 0 && s[0] == ':' { // dynamic segment
			if r.params == nil {
				r.params = make(map[int]string)
			}
			r.params[i] = s[1:]
		}
	}
	return r
}

// split splits the pattern, separating it into host and path.
func split(pattern string) (host, path string) {
	pStart := strings.Index(pattern, "/")
	if pStart == -1 {
		panic("path must begin with slash")
	}

	path = pattern[pStart:]

	// the domain part of the url is case insensitive
	host = strings.ToLower(pattern[:pStart])
	return
}

// HandleFunc registers the handler function for the given pattern.
// See the Handle method for details on how to register a pattern.
func (m *Muxer) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request), methods ...string) {
	if handler != nil {
		m.Handle(pattern, http.HandlerFunc(handler), methods...)
		return
	}
	m.Handle(pattern, nil, methods...)
}

// ServeHTTP dispatches the request to the handler whose
// pattern most closely matches the request URL.
//
// If the path is not in its canonical form, the
// handler will be an internally-generated handler
// that redirects to the canonical path.
func (m *Muxer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if r.RequestURI == "*" {
		if r.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method != "CONNECT" {
		if p := cleanPath(r.URL.Path); p != r.URL.Path {
			url := *r.URL
			url.Path = p
			http.Redirect(w, r, url.String(), http.StatusMovedPermanently)
			return
		}
	}

	h, args, methodMismatch := m.match(r.Method, r.Host, r.URL.Path)
	if h != nil {
		if len(args) > 0 {
			ctx := r.Context()
			for key, value := range args {
				ctx = context.WithValue(ctx, CtxKey(key), value)
			}
			r = r.WithContext(ctx)
		}
		h.ServeHTTP(w, r)
		return
	}

	if m.notFoundHandler != nil {
		m.notFoundHandler(w, r, methodMismatch)
		return
	}

	status := http.StatusNotFound
	if methodMismatch {
		status = http.StatusMethodNotAllowed
	}
	text := http.StatusText(status)
	http.Error(w, text, status)
}

// Return the canonical path for p, eliminating . and .. elements.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}
	return np
}

func (m *Muxer) match(method, _, path string) (h http.Handler, args args, methodMismatch bool) {
	endsInSlash := path[len(path)-1] == '/'
	segments := strings.Split(strings.Trim(path, "/"), "/")
	slen := len(segments)

	candidates := m.possibleRoutes(slen, endsInSlash)
	candLen := len(candidates)

LOOP:
	for i := slen - 1; i >= 0; i-- {
		s := segments[i]

		for k, r := range candidates {
			if r != nil && r.notMatch(s, i) {
				candidates[k] = nil
				candLen -= 1
			}
		}
		if candLen == 0 {
			break LOOP
		}
	}

	if candLen > 0 {
		for _, c := range candidates {
			if c != nil && c.methodSupported(method) {
				args = c.argsMap(segments)
				if c.len < slen || endsInSlash {
					h = c.slashHandler
				} else {
					h = c.nonSlashHandler
				}
				return
			}
		}
		methodMismatch = true
	}
	return
}

func (m *Muxer) possibleRoutes(slen int, endsInSlash bool) []*route {
	routes := make([]*route, 0, len(m.routes))
	for _, r := range m.routes {
		if r.len == slen && ((endsInSlash && r.slashHandler != nil) || (!endsInSlash && r.nonSlashHandler != nil)) {
			routes = append(routes, r)
		} else if r.len < slen && r.slashHandler != nil {
			routes = append(routes, r)
		}
	}
	return routes
}

// NotFound registers a handler that will be called when
// the Muxer didn't find a suitable handler for the request.
// The handler can be used to reply to the request with a custom error.
//
// The handler will be passed the http.ResponseWriter, the original
// http.Request and a bool argument, which indicates whether
// the request path matches a pattern but the request method
// does not match the method associated with the pattern.
// It can be used to distinguish between 404 Not Found and
// 405 Method Not Allowed errors.
func (m *Muxer) NotFound(h func(w http.ResponseWriter, r *http.Request, methodMismatch bool)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notFoundHandler = h
}

// CtxKey is the type of the context keys at which named parameter
// values are stored.
//
// Use the request context's Value() method to retrieve a value:
//   value := req.Context().Value(mux.CtxKey("username")))
type CtxKey string
