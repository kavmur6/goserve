package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/net/websocket"
)

type Server struct {
	conns      map[*websocket.Conn]bool
	clients    []*websocket.Conn
	serverLock sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		conns:   make(map[*websocket.Conn]bool),
		clients: []*websocket.Conn{},
	}

}

func remove(l []*websocket.Conn, ws *websocket.Conn) []*websocket.Conn {
	for i, other := range l {
		if other == ws {
			return append(l[:i], l[i+1:]...)
		}
	}
	return nil
}

func (s *Server) contains(ws *websocket.Conn) bool {
	for _, a := range s.clients {
		if a == ws {
			return true
		}
	}
	return false
}

func (s *Server) handleWs(ws *websocket.Conn) {
	fmt.Println("incoming ", ws.RemoteAddr())
	s.conns[ws] = true
	if !s.contains(ws) {
		s.serverLock.Lock()
		s.clients = append(s.clients, ws)
		s.serverLock.Unlock()
		fmt.Println("Number of clients ", len(s.clients))
		s.readLoop(ws)
	} else {
		fmt.Println(("redundant conn"))
	}
}

func (s *Server) readLoop(ws *websocket.Conn) {
	buff := make([]byte, 1024)
	for {
		n, err := ws.Read(buff)
		if err != nil {
			if err == io.EOF {
				fmt.Println("closed, clients rem:", len(s.clients))
				s.clients = remove(s.clients, ws)
				ws.Close()
				break
			}
			fmt.Println("read err")
		}
		msg := buff[:n]
		fmt.Println(string(msg))
		if string(msg) == "NoId" {
			fmt.Println("closed, clients rem:", len(s.clients))
			s.clients = remove(s.clients, ws)
			ws.Close()
		}
		if !strings.HasPrefix(string(msg), "sess-") {
			for _, c := range s.clients {
				if ws != c {
					c.Write(msg)
				}
			}
		}
	}
}

func main() {
	s := NewServer()
	http.Handle("/ws", websocket.Handler(s.handleWs))
	http.ListenAndServe(":8080", nil)
	fmt.Println("listening")
}
