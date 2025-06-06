package main

type WsServer struct {
	clients    map[*Client]bool
	register   chan *Client // new client connections
	unregister chan *Client // disconnect the client
	broadcast  chan []byte  // broadcast message to all client
}

// NewWebsocketServer creates a new WsServer type
func NewWebsocketServer() *WsServer {
	return &WsServer{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}
}

// Run our websocket server, accepting various requests
// listen register, unregister and broadcast events
func (server *WsServer) Run() {
	for {
		select {

		case client := <-server.register:
			server.registerClient(client)

		case client := <-server.unregister:
			server.unregisterClient(client)

		case message := <-server.broadcast:
			server.broadcastToClients(message)
		}
	}
}

// send the mssage to all clients to channel send
func (server *WsServer) broadcastToClients(message []byte) {
	for client := range server.clients {
		client.send <- message
	}
}

func (server *WsServer) registerClient(client *Client) {
	server.clients[client] = true
}

func (server *WsServer) unregisterClient(client *Client) {
	delete(server.clients, client)
	// if _, ok := server.clients[client]; ok {
	// 	delete(server.clients, client)
	// }
}
