package api

import (
	"github.com/googollee/go-socket.io"
)

type EventHandler struct {
	*socketio.Server
}

func NewEventHandler() *EventHandler {
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Errorf("new event handler failed, error: %s", err.Error())
		return nil
	}

	server.On("connection", func(so socketio.Socket) {
		so.Join("fabric")
		so.On("message", func(msg string) {
			log.Debug("emit:", msg, so.Emit("message", msg))
			so.Emit("event", map[string]string{"hello": "gogoog"})
		})
		so.On("disconnection", func() {
			log.Debug("on disconnect")
		})
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Debug("error:", err)
	})

	// go func() {
	// 	for {
	// 		select {
	// 		case msg := <-chMsg:
	// 			if server.Count() > 0 {
	// 				server.BroadcastTo("fabric", "event", msg)
	// 				log.Debug("sended", msg)
	// 			}
	// 		}
	// 	}
	// }()

	// go func() {
	// 	time.Sleep(time.Second)
	// 	for {
	// 		select {
	// 		case <-time.Tick(2 * time.Second):
	// 			chMsg <- map[string]string{"he": "ooo" + time.Now().String()}
	// 		}
	// 	}
	// }()

	return &EventHandler{server}
}

// func (e *EventHandler) Send(msg interface{}) {

// }

func (e *EventHandler) Broadcast(msg interface{}) {
	e.BroadcastTo("fabric", "event", msg)
}
