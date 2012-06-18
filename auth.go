// Copyright 2011 Bobby Powers. All rights reserved.
package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"github.com/bpowers/seshcookie"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"
	"time"
)

type Authorizer interface {
	Authorized(name, pass string) bool
}

type Decider interface {
	// returns true if a certain resource should be served to an
	// unauthorized user
	Allow(path string) bool
}

// variadic version of strings.HasPrefix
func containsPrefix(s string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

type decider struct{}
type socketDecider struct{}

func (d *decider) Allow(path string) bool {
	if containsPrefix(path, "/css/", "/img/", "/js/", "/favicon.ico") {
		return true
	}
	return false
}

func (d *socketDecider) Allow(path string) bool {
	return false
}

type authorizer struct {
	basePath string
}

func cleanUser(user string) string {
	user = strings.TrimSpace(user)
	// clean the name as if it was a path, removing the leading
	// '/' or '.', which is guaranteed to be there
	return path.Clean("/" + user)[1:]
}

// returns true if the sha256 hex password hash (of $name|$pass)
// stored at $basePath/$name/.password matches the user/pass
// arguements here.
func (a *authorizer) Authorized(name, pass string) bool {
	nameCleaned := cleanUser(name)
	// basically, if we detect any path manipulation stuff, bail
	if len(nameCleaned) == 0 || nameCleaned != name || strings.Index(nameCleaned, "/") != -1 {
		return false
	}

	passwordPath := path.Join(a.basePath, ".auth", nameCleaned)
	expectedHashHex, err := ioutil.ReadFile(passwordPath)
	if err != nil {
		log.Print("reading password file failed: ", err)
		return false
	}
	expectedHash := make([]byte, sha256.Size)
	_, err = hex.Decode(expectedHash, expectedHashHex)
	if err != nil {
		log.Printf("%s: hex.Decode: %s\n", passwordPath, err)
		return false
	}

	// hash the user supplied password
	inputHash := sha256.New()
	inputHash.Write([]byte(nameCleaned))
	inputHash.Write([]byte("|"))
	inputHash.Write([]byte(pass))

	// and check if it matches the one on disk
	return subtle.ConstantTimeCompare(expectedHash, inputHash.Sum(nil)) == 1
}

func serveLogin(rw http.ResponseWriter, isError bool) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	forceRevalidate(rw)
	login := template.Must(template.ParseFiles("./tmpl/login.html"))
	if err := login.Execute(rw, isError); err != nil {
		log.Printf("login tmpl.Execute: %v\n", err)
	}
}

// AuthHandler is an http.Handler which is meant to be sandwiched
// between the seshcookie session handler and the handler for
// resources you wish to require authentication to access.
type AuthHandler struct {
	http.Handler
	Auth            Authorizer
	Unauthenticated Decider // some resources might not need a login
}

func (h *AuthHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	session := seshcookie.Session.Get(req)
	log.Printf("using session: %#v\n", session)

	redirect := session["redirect"]
	delete(session, "redirect")
	log.Print("path: ", req.URL.Path)
	switch req.URL.Path {
	case "/login":
		if req.Method != "POST" {
			serveLogin(rw, false)
			return
		}
		err := req.ParseForm()
		if err != nil {
			log.Printf("error '%s' parsing form for %#v\n", err, req)
		}
		user := req.Form.Get("user")
		pass := req.Form.Get("pass")
		if !h.Auth.Authorized(user, pass) {
			log.Printf("authentication failed for %s (pass:%s)\n",
				user, pass)
			// prevent brute force login attempts, sleep
			// for 100 ms
			time.Sleep(100000000)
			session["redirect"] = redirect
			serveLogin(rw, true)
			return
		}

		log.Printf("authorized %s\n", user)
		session["user"] = user
		delete(session, "redirect")
		path, ok := redirect.(string)
		if !ok {
			path = "/"
		}
		http.Redirect(rw, req, path, http.StatusFound)
		return
	case "/logout":
		delete(session, "user")
		http.Redirect(rw, req, "/login", http.StatusFound)
		return
	}

	if _, ok := session["user"]; !ok {
		if h.Unauthenticated != nil && !h.Unauthenticated.Allow(req.URL.Path) {
			session["redirect"] = req.URL.Path
			http.Redirect(rw, req, "/login", http.StatusFound)
			return
		}
	}

	h.Handler.ServeHTTP(rw, req)
}
