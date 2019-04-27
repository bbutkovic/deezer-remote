package hub

type remote struct {
	quitChannel   chan bool
	updateChannel chan []byte
	token         string
}
