package main

import (
	"fmt"
	"net"
)

type Server struct {
	Ip   string
	Port int
}

// NewServer create a server
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,
	}
	return server
}

// Handler handle business
func (s *Server) Handler(conn net.Conn) {
	fmt.Println("Connection is established")
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
