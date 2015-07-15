package main

const (
	HubCmdShutdown = 1 << iota
)

type conn chan []byte

type hub struct {
	conns      map[conn]bool
	broadcast  chan []byte
	register   chan conn
	unregister chan conn
}

var workerHub = hub{
	conns:      make(map[conn]bool),
	broadcast:  make(chan []byte),
	register:   make(chan conn),
	unregister: make(chan conn),
}

func (h *hub) run() {
	go func() {
		for {
			select {
			case c := <-h.register:
				h.conns[c] = true
			case c := <-h.unregister:
				if _, ok := h.conns[c]; ok {
					delete(h.conns, c)
				}
			case m := <-h.broadcast: // broadcast to workers
				for c := range h.conns {
					select {
					case c <- m:
					default:
						delete(h.conns, c)
					}
				}
			}
		}
	}()
}

func (h *hub) Broadcast(cmd []byte) {
	h.broadcast <- cmd
}

func (h *hub) Register(c conn) {
	h.register <- c
}

func (h *hub) Unregister(c conn) {
	h.unregister <- c
}
