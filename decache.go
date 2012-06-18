package main

import (
	"net/http"
	"time"
)

type decacheHandler struct {
	http.Handler
}

// An hour ago.
func aWhileAgo() time.Time {
	return time.Now().Add(-time.Hour)
}

func addExpiresHeaderFor(path string) bool {
	return true
}

func forceRevalidate(w http.ResponseWriter) {
	w.Header().Set("Expires", aWhileAgo().Format(http.TimeFormat))
}

// by setting the expires tag 30 minutes into the past, we ensure that
// browsers _always_ do a GET for resources with an If-Modified-Since
// header.  If they have the latest copy they get a 304 response and
// use their cache, but the point is they always check.
func (self *decacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if addExpiresHeaderFor(r.URL.Path) {
		forceRevalidate(w)
	}
	self.Handler.ServeHTTP(w, r)
}
