// Copyright 2011 Bobby Powers. All rights reserved.
package main

import (
	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/goconf/conf"
	"flag"
	"github.com/bpowers/seshcookie"
	"log"
	"net/http"
	"os"
	"path"
)

var (
	apiKey       = "AIzaSyBO09P0AN4Fr5yLnQ0MkC3zeOekD4os79Y"
	clientId     = "1032050692895.apps.googleusercontent.com"
	clientSecret = "kY9ulXEdxY9hlQDBoCQuzO6e"
	oauthConfig  = &oauth.Config{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		Scope:        "https://www.googleapis.com/auth/tasks",
		AuthURL:      "https://accounts.google.com/o/oauth2/auth",
		TokenURL:     "https://accounts.google.com/o/oauth2/token",
		RedirectURL:  "https://boosd.org/oauth2callback",
	}
)

func main() {
	devMode := flag.Bool("dev", false, "run on port 8080, rather than 8443")
	flag.Parse()

	conf, err := conf.ReadConfigFile("config")
	if err != nil {
		log.Printf("reading config failed: %s\n", err)
		return
	}
	key, err := conf.GetString("", "key")
	if err != nil {
		log.Printf("reading config.key failed: %s", err)
		return
	}
	name, err := conf.GetString("", "cookie-name")
	if err != nil {
		log.Printf("reading config.cookie-name failed: %s", err)
		return
	}
	musicDir, err := conf.GetString("", "music-dir")
	if err != nil {
		log.Printf("reading config.models-dir failed: %s", err)
		return
	}
	if musicDir[:2] == "~/" {
		musicDir = path.Join(os.Getenv("HOME"), musicDir[2:])
	}

	log.Printf("using music in %s\n", musicDir)

	rootHandler := seshcookie.NewSessionHandler(
		&AuthHandler{
			http.FileServer(http.Dir("./static")),
			&authorizer{musicDir},
			&decider{},
		},
		key,
		nil)
	rootHandler.CookieName = name

	http.Handle("/", rootHandler)
	http.Handle("/err/", http.FileServer(http.Dir("./err")))
	http.Handle("/music/", http.StripPrefix("/music",
		http.FileServer(http.Dir(musicDir))))
	http.HandleFunc("/api/", proxyToCnote)
	http.HandleFunc("/pkg/", proxyToGodoc)
	http.HandleFunc("/doc/", proxyToGodoc)

	watch, err := NewDirwatch(musicDir)
	if err != nil {
		log.Printf("NewDirwatch: %r\n", err)
		return
	}
	_ = watch

	if *devMode {
		err = http.ListenAndServe(
			":8080",
			nil)
	} else {
		go func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
				http.Redirect(rw, r, "https://lightswitchrave.net/", 302)
			})
			http.ListenAndServe(":8080", mux)
		}()
		// if we're serving over https, set the secure flag
		// for cookies
		seshcookie.Session.Secure = true
		err = http.ListenAndServeTLS(
			":8443",
			"/home/bpowers/.tls/certchain.pem",
			"/home/bpowers/.tls/boosd.org_key.pem",
			nil)
	}
	if err != nil {
		log.Printf("ListenAndServe:", err)
	}
}
