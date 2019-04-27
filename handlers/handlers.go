package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/bbutkovic/deezer-remote/hub"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const maxTokenAttempts = 5

var playerActions = []string{"play", "pause", "next", "prev", "setVolume", "setPosition", "setQueue"}

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewTokenHandler(hub *hub.Hub, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	for i := 0; i < maxTokenAttempts; i++ {
		token := generateToken(8, charset)
		if hub.PlayerExists(token) == false {
			fmt.Fprintf(w, "{\"token\":\"%s\"}", token)
			return
		}
	}
	http.Error(w, "Failed to generate token.", 500)
	log.Println("Failed to generate new random token.")
}

func PlayerWSHandler(hub *hub.Hub, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	uc, qc, err := hub.NewPlayer(vars["token"])
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	//We managed to create a new player, proceed to upgrade to WS
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error while creating a connection.", 500)
		qc <- true
		return
	}

	writePump(conn, uc, qc)
}

func writePump(conn *websocket.Conn, uc <-chan []byte, qc chan<- bool) {
	ticker := time.NewTicker(1 * time.Second)
	defer func() {
		qc <- true
		ticker.Stop()
		conn.Close()
	}()

	clientDisconnect := make(chan bool)

	go func() {
		validCloseCodes := []int{
			websocket.CloseGoingAway,
			websocket.CloseNormalClosure,
		}

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, validCloseCodes...) {
					log.Printf("Unexpected socket error: %v", err)
				}
				//Connection is done, send a client disconnect
				clientDisconnect <- true
				return
			}
		}
	}()

	for {
		select {
		case update, ok := <-uc:
			conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			//Got a new update, send it over to the WS
			err := conn.WriteMessage(websocket.TextMessage, update)
			if err != nil {
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-clientDisconnect:
			return
		}
	}
}

func RemoteWSHandler(hub *hub.Hub, w http.ResponseWriter, r *http.Request) {

}

type PlayerCommand struct {
	Action string `json:"action"`
	Value  string `json:"value,omitempty"`
}

func SendPlayerCommand(hub *hub.Hub, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if !hub.PlayerExists(vars["token"]) {
		http.Error(w, "Player with that token does not exist.", 400)
		return
	}

	var command PlayerCommand
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&command)
	if err != nil {
		http.Error(w, "Request payload incorrect.", 400)
		return
	}
	if !checkCorrectAction(command.Action, playerActions) {
		http.Error(w, "Action incorrect.", 400)
		return
	}

	payload, _ := json.Marshal(&command)
	hub.SendToPlayer(vars["token"], payload)
}

func checkCorrectAction(action string, correctActions []string) bool {
	for _, correctAction := range correctActions {
		if correctAction == action {
			return true
		}
	}
	return false
}

//Generates an alphanumeric token with the given length
func generateToken(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
