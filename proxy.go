package main

import (
	"io"
	"log"
	"net/http"
)

func proxyToCnote(w http.ResponseWriter, r *http.Request) {
	// don't cache any API responses
	forceRevalidate(w)
	path := r.URL.Path[len("/api/"):]
	uri := "http://127.0.0.1:1969/" + path
	resp, err := http.Get(uri)
	if err != nil {
		log.Printf("err proxying %s: %r\n", uri, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
