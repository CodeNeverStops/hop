package main

const (
	WorkerCmdShutdown = 1 << iota
)

type conn chan int

type hub struct {
	conns      map[conn]bool
	broadcast  chan int
	register   chan conn
	unregister chan conn
}

var workerHub = hub{
	conns:      make(map[conn]bool),
	broadcast:  make(chan int),
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
					close(c)
					delete(h.conns, c)
				}
			case m := <-h.broadcast: // broadcast to workers
				for c := range h.conns {
					select {
					case c <- m:
					default:
						close(c)
						delete(h.conns, c)
					}
				}
			}
		}
	}()
}

func (h *hub) Broadcast(cmd int) {
	h.broadcast <- cmd
}

func (h *hub) Register(c conn) {
	h.register <- c
}

func (h *hub) Unregister(c conn) {
	h.unregister <- c
}
