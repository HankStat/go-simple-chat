package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

const (
	// Max wait time when writing message to peer
	writeWait = 10 * time.Second

	// Max time till next pong from peer
	pongWait = 60 * time.Second

	// Send ping interval, must be less then pong wait time
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 10000
)

var upgrader websocket.Upgrader

func init() {
	_ = godotenv.Load() // Load .env file

	origin := os.Getenv("WS_ALLOWED_ORIGIN")

	upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		// frontend and backend run on different origin
		CheckOrigin: func(r *http.Request) bool {
			return r.Header.Get("Origin") == origin
		},
	}
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// define Client that will be client at the server
type Client struct {
	// The actual websocket connection.
	conn     *websocket.Conn
	wsServer *WsServer
	send     chan []byte
}

// define the function for creating new client
func newClient(conn *websocket.Conn, wsServer *WsServer) *Client {
	return &Client{
		conn:     conn,
		wsServer: wsServer,
		send:     make(chan []byte, 256),
	}
}

// listen the messages from client
func (client *Client) readPump() {
	defer func() {
		// disconnect when exit the function
		client.disconnect()
	}()

	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	// refresh the timeout
	client.conn.SetPongHandler(func(string) error { client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// Start endless read loop, waiting for messages from client
	for {
		_, jsonMessage, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected close error: %v", err)
			}
			break
		}
		// send the message to broadcast channel
		client.wsServer.broadcast <- jsonMessage
	}
}

// pump the messasges from the send channel to websocket connection
func (client *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()
	for {
		select {
		case message, ok := <-client.send: // read from send channel
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The WsServer closed the channel.
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			// write for new message
			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Attach queued chat messages to the current websocket message.
			n := len(client.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-client.send)
			}
			// Close the writer and flush the message to websocket connection
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			// send ping
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (client *Client) disconnect() {
	// remove the client from the server
	// close send channel to stop writePump goroutine
	// close the connection
	client.wsServer.unregister <- client
	close(client.send)
	client.conn.Close()
}

// ServeWs handles websocket requests from clients requests.
// public
func ServeWs(wsServer *WsServer, w http.ResponseWriter, r *http.Request) {
	// upgrade HTTP to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := newClient(conn, wsServer)
	// run two goroutines
	go client.writePump()
	go client.readPump()
	// register the client
	wsServer.register <- client

	// fmt.Println("New Client joined the hub!")
	// fmt.Println(client)
}
