package hub

import (
	"errors"
	"log"
	"sync"
)

type Hub struct {
	players          map[string]*player
	remotes          map[string][]*remote
	playerBroadcast  chan broadcastMessage
	remotesBroadcast chan broadcastMessage
	lock             *sync.Mutex
}

type broadcastMessage struct {
	target  string
	payload []byte
}

func NewHub() *Hub {
	return &Hub{
		players:          make(map[string]*player),
		playerBroadcast:  make(chan broadcastMessage),
		remotesBroadcast: make(chan broadcastMessage),
		lock:             &sync.Mutex{},
	}
}

func (h *Hub) SendToPlayer(token string, update []byte) error {
	message := broadcastMessage{
		target:  token,
		payload: update,
	}
	h.playerBroadcast <- message
	return nil
}

func (h *Hub) SendToRemotes(token string, update []byte) error {
	message := broadcastMessage{
		target:  token,
		payload: update,
	}
	h.remotesBroadcast <- message
	return nil
}

func (h *Hub) NewPlayer(token string) (<-chan []byte, chan<- bool, error) {
	if h.PlayerExists(token) {
		return nil, nil, errors.New("Player with that token already exists.")
	}

	h.lock.Lock()
	uc := make(chan []byte)
	qc := make(chan bool)
	player := &player{
		updateChannel: uc,
		quitChannel:   qc,
		token:         token,
	}
	h.players[token] = player
	h.lock.Unlock()

	return uc, qc, nil
}

func (h *Hub) NewRemote(token string) (<-chan []byte, chan<- bool, error) {
	h.lock.Lock()
	uc := make(chan []byte)
	qc := make(chan bool)
	remote := &remote{
		updateChannel: uc,
		quitChannel:   qc,
		token:         token,
	}
	h.remotes[token] = append(h.remotes[token], remote)
	h.lock.Unlock()

	return uc, qc, nil
}

func (h *Hub) DestroyPlayer(token string) error {
	if !h.PlayerExists(token) {
		return errors.New("Player with that token does not exist.")
	}

	h.lock.Lock()
	delete(h.players, token)
	h.lock.Unlock()
	return nil
}

func (h *Hub) PlayerExists(token string) bool {
	h.lock.Lock()
	_, exists := h.players[token]
	h.lock.Unlock()
	return exists
}

func (h *Hub) Run() {
	for {
		select {
		case msg := <-h.playerBroadcast:
			//Send the message over to a single place (the player)
			if !h.PlayerExists(msg.target) {
				log.Printf("Failed sending to non-existant player with token %s", msg.target)
			} else {
				h.lock.Lock()
				h.players[msg.target].updateChannel <- msg.payload
				h.lock.Unlock()
			}
		case msg := <-h.remotesBroadcast:
			//Send the message over to multiple places (the remotes of a single player)
			h.lock.Lock()
			if remoteGroup, ok := h.remotes[msg.target]; ok {
				for _, remote := range remoteGroup {
					remote.updateChannel <- msg.payload
				}
			}
			h.lock.Unlock()

		default:
		}
		h.lock.Lock()
		for token, player := range h.players {
			select {
			case <-player.quitChannel:
				//Handle a player disconnect
				close(player.updateChannel)
				delete(h.players, token)
			default:
			}
		}
		for token, remoteGroup := range h.remotes {
			//Find any remotes that have disconnected
			var toRemove []int
			for i, remote := range remoteGroup {
				select {
				case <-remote.quitChannel:
					//Handle a remote disconnect
					toRemove = append(toRemove, i)
				default:
				}
			}
			for i := range toRemove {
				close(h.remotes[token][i].updateChannel)
				h.remotes[token] = append(h.remotes[token][:i], h.remotes[token][i+1:]...)
			}
		}
		h.lock.Unlock()
	}
}
