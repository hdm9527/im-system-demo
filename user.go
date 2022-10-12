package main

import (
	"net"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// NewUser create a user
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	//启动监听当前user channel消息的goroutine
	go user.ListenMessage()

	return user
}

// Online user online
func (u *User) Online() {
	//When the user goes online, add the user to the onlineMap
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	//Broadcast user online message
	u.server.BroadCast(u, "已上线")
}

// Offline user offline
func (u *User) Offline() {
	//When the user goes offline, remove from onlineMap
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	//Broadcast user online message
	u.server.BroadCast(u, "已下线")
}

// SendMsg Send a message to the client corresponding to the current User
func (u *User) SendMsg(msg string) {
	u.conn.Write([]byte(msg))
}

// handleMessage handle user message
func (u *User) handleMessage(msg string) {
	if msg == "who" {
		//query online users
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
			u.SendMsg(onlineMsg)
		}
		u.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//message format： rename|张三
		newName := msg[7:]
		//newName := strings.Split(msg, "|")[1]

		//判断name是否存在
		_, ok := u.server.OnlineMap[newName]
		if ok {
			u.SendMsg("当前用户名已被使用")
		} else {
			u.server.mapLock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.server.OnlineMap[newName] = u
			u.server.mapLock.Unlock()

			u.Name = newName
			u.SendMsg("您已经更新用户名：" + u.Name + "\n")
		}
	} else {
		u.server.BroadCast(u, msg)
	}
}

// ListenMessage listen for messages
func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}
