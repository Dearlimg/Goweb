package main

import (
	"net"
	"strings"
)

type user struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

func (user *user) Online() {
	user.server.maplock.Lock()
	user.server.onlineMap[user.Name] = user
	user.server.maplock.Unlock()

	user.server.Broadcast(user, "已上线")
}

func (user *user) Offline() {
	user.server.maplock.Lock()
	delete(user.server.onlineMap, user.Name)
	user.server.maplock.Unlock()

	user.server.Broadcast(user, "已下线")
}

func (user *user) SendMsg(msg string) {
	user.conn.Write([]byte(msg))
}

func (user *user) DoMessage(msg string) {
	if msg == "who" {
		user.server.maplock.Lock()
		for _, user := range user.server.onlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线"
			user.SendMsg(onlineMsg)
		}
		user.server.maplock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := msg[7:]
		_, ok := user.server.onlineMap[user.Name]
		if ok {
			user.SendMsg("当前用户名已经被使用\n")
		} else {
			user.server.maplock.Lock()
			delete(user.server.onlineMap, user.Name)
			user.server.onlineMap[newName] = user
			user.server.maplock.Unlock()

			user.Name = newName
			user.SendMsg("用户名已经更新:" + newName + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			user.SendMsg("没有输入名字")
			return
		}

		remoteUser, ok := user.server.onlineMap[remoteName]
		if !ok {
			user.SendMsg("用户不存在")
			return
		}

		content := strings.Split(msg, "|")[2]
		if content == "" {
			user.SendMsg("内容为空")
			return
		}
		remoteUser.SendMsg(user.Name + "对您说:" + content + "\n")
	} else {
		user.SendMsg(msg)
	}

}

func newUser(conn net.Conn, server *Server) *user {
	userAddr := conn.RemoteAddr().String()
	user := &user{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	go user.ListenMessage()
	return user
}

func (user *user) ListenMessage() {
	for {
		msg := <-user.C
		user.conn.Write([]byte(msg + "\n"))
	}
}
