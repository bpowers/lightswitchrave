package main

import (
	"io"
	"log"
	"net/http"
	"strings"
)

func proxyToCnote(w http.ResponseWriter, r *http.Request) {
	// don't cache any API responses
	forceRevalidate(w)
	path := r.URL.Path[len("/api/"):]
	path = strings.Replace(path, "#", "%23", -1)
	path = strings.Replace(path, "?", "%3F", -1)
	uri := "http://127.0.0.1:1969/" + path
	resp, err := http.Get(uri)
	if err != nil {
		log.Printf("err proxying %s: %r\n", uri, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	if proxiedType := resp.Header.Get("Content-Type"); proxiedType != "" {
		w.Header().Add("Content-Type", proxiedType)
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func proxyToGodoc(w http.ResponseWriter, r *http.Request) {
	// don't cache any API responses
	forceRevalidate(w)
	path := r.URL.Path
	uri := "http://127.0.0.1:6060/" + path
	resp, err := http.Get(uri)
	if err != nil {
		log.Printf("err proxying %s: %r\n", uri, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	if proxiedType := resp.Header.Get("Content-Type"); proxiedType != "" {
		w.Header().Add("Content-Type", proxiedType)
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
