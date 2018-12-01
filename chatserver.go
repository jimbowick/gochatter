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
	rooms := []string{"main roomy", "second roomy"}
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

	type basicsockMess struct {
		Messagetype string
		Payload     string
	}
	type messwithRoom struct {
		Messagetype string
		Payload     string
		Inroom      string
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
			cm := &socketMessage{
				Messagetype: "chatrooms",
				Payload:     rooms}
			jcm, _ := json.Marshal(cm)
			conn.WriteMessage(websocket.TextMessage, jcm)
			fmt.Println(name, " connected, current users ", connections)
			sendUserRefresh := func() {
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
			sendUserRefresh()
			go func() {
				for {
					_, message, err := conn.ReadMessage()
					if err != nil {
						fmt.Println(err, name, " disconnected")
						delete(connections, conn)
						conn.Close()
						sendUserRefresh()
						return
					}
					message = bytes.TrimSpace(bytes.Replace(message, []byte{'\n'}, []byte{' '}, -1))
					println(name, "sent ", string(message))
					result := &messwithRoom{}
					json.Unmarshal(message, result)
					if result.Messagetype == "roomname" {
						rooms = append(rooms, result.Payload)
						sm := &socketMessage{
							Messagetype: "chatrooms",
							Payload:     rooms}
						jsm, _ := json.Marshal(sm)
						for conny := range connections {
							conny.WriteMessage(websocket.TextMessage, jsm)
						}
					} else if result.Messagetype == "chat" {
						for conny, anm := range connections {
							finny := anm + ": " + string(result.Payload)
							jg := &messwithRoom{
								Messagetype: "chat",
								Payload:     finny,
								Inroom:      result.Inroom,
							}
							jgg, _ := json.Marshal(jg)
							conny.WriteMessage(websocket.TextMessage, jgg)
						}
					} else if result.Messagetype == "setName" {
						connections[conn] = result.Payload
						sendUserRefresh()
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
