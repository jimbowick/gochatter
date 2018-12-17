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
	type user struct {
		username     string
		invitedrooms []string
	}
	var connections = make(map[*websocket.Conn]*user)
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
	type inInvite struct {
		Messagetype string
		Payload     string
		Roomname    string
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
			connections[conn] = &user{username: name, invitedrooms: []string{}}
			cm := &socketMessage{
				Messagetype: "chatrooms",
				Payload:     rooms}
			jcm, _ := json.Marshal(cm)
			conn.WriteMessage(websocket.TextMessage, jcm)
			fmt.Println(name, " connected, current users ", connections)
			sendUserRefresh := func() {
				usernames := []string{}
				for _, usr := range connections {
					nmy := usr.username
					usernames = append(usernames, nmy)
				}
				sm := &socketMessage{
					Messagetype: "users",
					Payload:     usernames}
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
						for conny := range connections {
							finny := connections[conn].username + ": " + string(result.Payload)
							jg := &messwithRoom{
								Messagetype: "chat",
								Payload:     finny,
								Inroom:      result.Inroom,
							}
							jgg, _ := json.Marshal(jg)
							conny.WriteMessage(websocket.TextMessage, jgg)
						}
					} else if result.Messagetype == "setName" {
						connections[conn].username = result.Payload
						sendUserRefresh()
					} else if result.Messagetype == "invite" {
						for conny, boy := range connections {
							if boy.username == result.Payload {
								connections[conny].invitedrooms = append(connections[conny].invitedrooms, result.Inroom)
								sm := &socketMessage{
									Messagetype: "invites",
									Payload:     connections[conny].invitedrooms}
								jsm, _ := json.Marshal(sm)
								conny.WriteMessage(websocket.TextMessage, jsm)
							}
						}

					} else if result.Messagetype == "Acceptinvite" {
						var newinvites = []string{}
						var oldinvs = connections[conn].invitedrooms
						for _, invroom := range oldinvs {
							if invroom != result.Payload {
								newinvites = append(newinvites, invroom)
							}
						}
						connections[conn].invitedrooms = newinvites
						sm := &socketMessage{
							Messagetype: "invites",
							Payload:     connections[conn].invitedrooms}
						jsm, _ := json.Marshal(sm)
						conn.WriteMessage(websocket.TextMessage, jsm)
					} else if result.Messagetype == "voice" {
						for conny := range connections {
							jg := &basicsockMess{
								Messagetype: "voice",
								Payload:     result.Payload,
							}
							jgg, _ := json.Marshal(jg)
							conny.WriteMessage(websocket.TextMessage, jgg)
						}
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
