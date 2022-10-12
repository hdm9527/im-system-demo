package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	//online user list
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//message broadcast channel
	Message chan string
}

// NewServer create a server
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// MessageListener The goroutine that listens to the Message broadcast message channel,
// and once there is a message, it is sent to all online users
func (s *Server) MessageListener() {
	for {
		msg := <-s.Message

		//Send msg to all online User
		s.mapLock.Lock()
		for _, cli := range s.OnlineMap {
			cli.C <- msg
		}
		s.mapLock.Unlock()
	}
}

// BroadCast broadcast message
func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	//send message to the channel of server message
	s.Message <- sendMsg
}

// Handler handle business
func (s *Server) Handler(conn net.Conn) {
	//fmt.Println("Connection is established")

	user := NewUser(conn, s)
	user.Online()

	//The channel that monitors whether the user is active
	isLive := make(chan bool)

	//Receive messages sent by clients
	go func() {
		buf := make([]byte, 4096)
		for {
			read, err := conn.Read(buf)
			if read == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			//extract the user's message
			msg := string(buf[:read-1])

			//broadcast the received message
			user.handleMessage(msg)

			//The user sends an arbitrary message, indicating that the current user is an active user
			isLive <- true
		}
	}()

	//The current handler is blocked
	for {
		select {
		case <-isLive:
			//The current user is active, reset the timer
			//Do nothing, in order to activate the select, update the timer below
		case <-time.After(time.Second * 10):
			//timeout to close the client's connection

			//resources to destroy
			close(user.C)

			//close connect
			err := conn.Close()
			if err != nil {
				fmt.Println("connect close err:", err)
				return
			}

			//Exit the current Handler
			return //runtime.Goexit()
		}
	}
}

// Start startup a tcp server
func (s *Server) Start() {
	//socket listen
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	//close listen socket
	defer func(listen net.Listener) {
		err := listen.Close()
		if err != nil {
			fmt.Println("close listen err:", err)
		}
	}(listen)

	//Start a goroutine that listens for Messages
	go s.MessageListener()

	//handle connection
	for {
		//accept
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("listen accept err:", err)
			continue
		}
		//do handler
		go s.Handler(conn)
	}
}
