package main

import (
	"fmt"
	"net"
)

type Client struct {
	ServerIP   string
	serverPort string
	Name       string
	conn       net.Conn
}

func NewClient(serverIP string, serverPort string) *Client {
	newClient := &Client{
		ServerIP:   serverIP,
		serverPort: serverPort,
	}
	return newClient
}

func main() {
	client := NewClient("127.0.0.1", "8080")
	if client == nil {
		fmt.Println("client is nil")
		return
	}

	fmt.Println("链接成功")

	select {}
}
