package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type ServerMsg struct {
	Type  int      `json:"type"`
	ID    string   `json:"id"`
	SDP   string   `json:"sdp"`
	Games []string `json:"games"`
}

type ClientMsg struct {
	ID       string `json:"id"`
	ServerID string `json:"serverid"`
	SDP      string `json:"sdp"`
	Conn     *websocket.Conn
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var serverWSMap map[string]*websocket.Conn
var clientWSMap map[string]ClientMsg

// ws endpoint to talk to the game server
func serverEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	readerServer(ws)
}

func readerServer(conn *websocket.Conn) {
	var msg ServerMsg
	for {
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println(err)
			return
		}

		switch msg.Type {
		case 1:
			// Init message - server gives out a name and list of games
			fmt.Println(msg.ID)
			serverWSMap[msg.ID] = conn

		case 2:
			// Ack message from the server - server sends out its sdp
			// sdp field will be set
			serverSDP := msg.SDP
			for _, v := range clientWSMap {
				if v.ServerID == msg.ID {
					v.Conn.WriteMessage(1, []byte(serverSDP))
				}
			}

		}

		// if err := conn.WriteMessage(1, []byte(msg.Games[0])); err != nil {
		// 	log.Println(err)
		// }
	}
}

// client endpoint that is hit up to give client id and sdp
func clientEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	readerClient(ws)
}

func readerClient(conn *websocket.Conn) {
	var msg ClientMsg
	for {
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println(err)
			return
		}

		msg.Conn = conn
		serverConn := serverWSMap[msg.ServerID]
		if serverConn == nil {
			conn.WriteMessage(1, []byte("Check the server id"))
		} else {
			clientWSMap[msg.ID] = msg
			serverConn.WriteMessage(1, []byte(msg.SDP))
		}

	}
}

func main() {
	serverWSMap = make(map[string]*websocket.Conn)
	clientWSMap = make(map[string]ClientMsg)

	http.HandleFunc("/server", serverEndpoint)
	http.HandleFunc("/client", clientEndpoint)

	log.Fatal(http.ListenAndServe(":9000", nil))
}
