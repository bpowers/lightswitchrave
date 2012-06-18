package main

import (
	"code.google.com/p/go.net/websocket"
	"github.com/bpowers/seshcookie"
	"log"
	"net/http"
	"strings"
)

func SocketServer(c *websocket.Conn) {
	log.Print("ok")
	websocket.JSON.Send(c, "hello")
	var err error
	var req map[string]interface{}
	for ; err == nil; err = websocket.JSON.Receive(c, &req) {
		log.Print("some request: ", req)
		//websocket.JSON.Send(c, msg)
	}
	log.Print("disconnected")
}

func restrictSocketPath(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		session := seshcookie.Session.Get(req)

		username, _ := session["user"].(string)

		// if an authorized user is trying to open a websocket
		// for some model that isn't his, return a 404
		if !strings.HasPrefix(req.URL.Path, "/ws/"+username+"/") {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			forceRevalidate(w)
			w.Write([]byte("not found."))
			return
		}

		handler.ServeHTTP(w, req)
	})
}
