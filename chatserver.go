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
				log.Println(err)
				return
			}
			connections[conn] = true
			go func() {
				for {
					_, message, _ := conn.ReadMessage()
					if err != nil {
						conn.Close()
						delete(connections, conn)
						return
					}
					log.Println("read a msg")
					message = bytes.TrimSpace(bytes.Replace(message, []byte{'\n'}, []byte{' '}, -1))
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
