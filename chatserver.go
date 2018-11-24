package main

import (
	"bytes"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func main() {
	var connections = make(map[*websocket.Conn]bool)
	http.HandleFunc(
		"/home",
		func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "static/home.html")
		},
	)
	http.HandleFunc(
		"/ws",
		func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				println(err)
				return
			}
			println(conn, " connected")
			connections[conn] = true
			go func() {
				for {
					_, message, err := conn.ReadMessage()
					if err != nil {
						println(conn, " connection closed")
						delete(connections, conn)
						conn.Close()
						return
					}
					message = bytes.TrimSpace(bytes.Replace(message, []byte{'\n'}, []byte{' '}, -1))
					println("someone sent: ", string(message))
					for conny := range connections {
						conny.WriteMessage(websocket.TextMessage, message)
					}
				}
			}()
		},
	)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
