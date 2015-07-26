package main

// Defined some commands to manage workers.
const (
	WorkerCmdShutdown = 1 << iota
)

// Define inbox type of workers.
type inbox chan int

// A hub collects inbox of all workers.
type hub struct {
	// a map of inbox of all workers
	conns map[inbox]bool

	// a channel use to receive command, and then broadcast the command to all inbox
	broadcast chan int

	// a channel use to register inbox of worker
	register chan inbox

	// a channel use to unregister inbox of worker
	unregister chan inbox
}

var workerHub = hub{
	conns:      make(map[inbox]bool),
	broadcast:  make(chan int),
	register:   make(chan inbox),
	unregister: make(chan inbox),
}

// Start the hub server,
// create a new goroutine to handle requests
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

func (h *hub) Register(c inbox) {
	h.register <- c
}

func (h *hub) Unregister(c inbox) {
	h.unregister <- c
}
