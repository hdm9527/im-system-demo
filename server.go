package main

import (
	"fmt"
	"io"
	"net"
	"sync"
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

	user := NewUser(conn)

	//When the user goes online, add the user to the onlineMap
	s.mapLock.Lock()
	s.OnlineMap[user.Name] = user
	s.mapLock.Unlock()

	//Broadcast user online message
	s.BroadCast(user, "已上线")

	//Receive messages sent by clients
	go func() {
		buf := make([]byte, 4096)
		for {
			read, err := conn.Read(buf)
			if read == 0 {
				s.BroadCast(user, "下线")
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			//extract the user's message
			msg := string(buf[:read-1])

			//broadcast the received message
			s.BroadCast(user, msg)
		}
	}()

	//The current handler is blocked
	select {}
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
