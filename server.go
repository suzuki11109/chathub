package chathub

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) writePump() {
	for message := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println(err)
			break
		}
	}
}

func NewServer() *Server {
	return &Server{
		registerCh:   make(chan *Client),
		unregisterCh: make(chan *Client),
		broadcaster:  make(chan []byte),
		clients:      make(map[*Client]bool),
	}
}

type Server struct {
	registerCh   chan *Client
	unregisterCh chan *Client
	broadcaster  chan []byte
	clients      map[*Client]bool
}

func (s *Server) Run() {
	for {
		select {
		case client := <-s.registerCh:
			s.clients[client] = true
			log.Println("connected")
		case client := <-s.unregisterCh:
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
			}
			log.Println("disconnected")

		case message := <-s.broadcaster:
			for client := range s.clients {
				select {
				case client.send <- message:
				default:
					delete(s.clients, client)
					close(client.send)
				}
			}

		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) handleClient(c *Client) {
	s.registerCh <- c

	defer func() {
		s.unregisterCh <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		s.broadcaster <- message
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ws upgrade failed:", err)
		return
	}

	c := &Client{conn: conn, send: make(chan []byte)}

	go c.writePump()
	go s.handleClient(c)
}
