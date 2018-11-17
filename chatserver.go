package main

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	clients := make(map[*client]bool)
	http.HandleFunc(
		"/home",
		func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "static/home.html")
		},
	)
	http.HandleFunc(
		"/ws",
		func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Println(err)
				return
			}
			cliento := &client{conn: conn, send: make(chan []byte, 256)}
			clients[cliento] = true
			go func() {
				ticker := time.NewTicker(pingPeriod)
				defer func() {
					ticker.Stop()
					cliento.conn.Close()
				}()
				for {
					<-ticker.C
					cliento.conn.SetWriteDeadline(time.Now().Add(writeWait))
					err := cliento.conn.WriteMessage(websocket.PingMessage, nil)
					if err != nil {
						return
					}
				}
			}()
			go func() {
				defer func() {
					delete(clients, cliento)
					close(cliento.send)
					cliento.conn.Close()
				}()
				cliento.conn.SetReadLimit(maxMessageSize)
				cliento.conn.SetReadDeadline(time.Now().Add(pongWait))
				cliento.conn.SetPongHandler(func(string) error { cliento.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
				for {
					_, message, err := cliento.conn.ReadMessage()
					if err != nil {
						if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
							log.Printf("error: %v", err)
						}
						break
					}
					message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
					for clienty := range clients {
						// clienty.conn.SetWriteDeadline(time.Now().Add(writeWait))
						writer, _ := clienty.conn.NextWriter(websocket.TextMessage)
						writer.Write(message)
						writer.Close()
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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type client struct {
	conn *websocket.Conn
	send chan []byte
}
