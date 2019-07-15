package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Init struct {
	ID    string   `json:"id"`
	Games []string `json:"games"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var serverWSMap map[string]*websocket.Conn

// ws endpoint to talk to the game server
func serverEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	reader(ws)
}

func reader(conn *websocket.Conn) {
	var msg Init
	for {
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(msg.ID)
		serverWSMap[msg.ID] = conn
		// if err := conn.WriteMessage(1, []byte(msg.Games[0])); err != nil {
		// 	log.Println(err)
		// }
	}
}

// client endpoint that is hit up to get server details
func connectEndpoint(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	serverId := r.Form.Get("server")
	sdp := r.Form.Get("sdp")
	fmt.Println(serverId)

	conn := serverWSMap[serverId]
	if conn == nil {
		w.Write([]byte("Check the server id"))
		return
	}
	err := conn.WriteMessage(1, []byte(sdp))
	if err != nil {
		log.Println(err)
	}
}

func main() {
	serverWSMap = make(map[string]*websocket.Conn)
	http.HandleFunc("/server", serverEndpoint)
	http.HandleFunc("/connect", connectEndpoint)

	log.Fatal(http.ListenAndServe(":9000", nil))
}
