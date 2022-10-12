package main

import "net"

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

// handleMessage handle user message
func (u *User) handleMessage(msg string) {
	u.server.BroadCast(u, msg)
}

// ListenMessage listen for messages
func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}
