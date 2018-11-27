package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func main() {
	var connections = make(map[*websocket.Conn]string)
	http.HandleFunc(
		"/home",
		func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "static/home.html")
		},
	)
	type socketMessage struct {
		Messagetype string
		Payload     []string
	}
	http.HandleFunc(
		"/ws",
		func(w http.ResponseWriter, r *http.Request) {
			name := r.URL.Query()["name"][0]
			upgrader := websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				fmt.Println(err)
				return
			}
			connections[conn] = name
			fmt.Println(name, " connected, current users ", connections)
			doeet := func() {
				users := []string{}
				for _, usr := range connections {
					users = append(users, usr)
				}
				sm := &socketMessage{
					Messagetype: "users",
					Payload:     users}
				jsm, _ := json.Marshal(sm)
				for conny := range connections {
					conny.WriteMessage(websocket.TextMessage, jsm)
				}
			}
			doeet()
			go func() {
				for {
					_, message, err := conn.ReadMessage()
					if err != nil {
						fmt.Println(err, name, " disconnected")
						delete(connections, conn)
						conn.Close()
						doeet()
						return
					}
					message = bytes.TrimSpace(bytes.Replace(message, []byte{'\n'}, []byte{' '}, -1))
					println(name, "sent ", string(message))
					for conny := range connections {
						conny.WriteMessage(websocket.TextMessage, message)
					}
				}
			}()
		},
	)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
