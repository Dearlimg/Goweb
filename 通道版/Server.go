package main

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	ip        string
	port      int
	onlineMap map[string]*user
	maplock   sync.RWMutex
	Message   chan string
}

func NewServer(ip string, port int) *Server {
	return &Server{
		ip:        ip,
		port:      port,
		onlineMap: make(map[string]*user),
		Message:   make(chan string),
	}
}

func (server *Server) Broadcast(user *user, msg string) {
	sendmessage := "[" + user.Addr + "]" + user.Name + ":" + msg
	server.Message <- sendmessage
}

func (server *Server) ListenMessage() {
	for {
		msg := <-server.Message
		server.maplock.Lock()
		for _, u := range server.onlineMap {
			u.C <- msg
		}
		server.maplock.Unlock()
	}
}

func (server *Server) Handler(conn net.Conn) {
	user := newUser(conn, server)

	user.Online()

	isLive := make(chan bool)

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				fmt.Println(err)
			}
			if n == 0 {
				user.Offline()
			}

			msg := string(buf[:n-1])

			user.DoMessage(msg)
			isLive <- true
		}

	}()

	for {
		select {
		case <-isLive:

		case <-time.After(time.Second * 10):
			user.SendMsg("你被踢了")
			close(user.C)
			conn.Close()
			return
		}
	}
}

// start interface
func (server *Server) Start() {
	listener, err := net.Listen("tcp", server.ip+":"+strconv.Itoa(server.port))
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	go server.ListenMessage()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go server.Handler(conn)
	}
}
