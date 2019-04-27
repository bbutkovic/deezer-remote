package hub

type playerUpdate struct {
	body []byte
	err  chan error
}

type player struct {
	quitChannel   chan bool
	updateChannel chan []byte
	token         string
}
