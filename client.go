package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int //client mode
}

func NewClient(serverIp string, serverPort int) *Client {
	//create a client
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	//link to server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}

	client.conn = conn

	//return object
	return client
}

// DealResponse Process the message responded by the server and display it directly to the standard output
func (c *Client) DealResponse() {
	//Once c.conn has data, copy it directly to stdout standard output, permanently blocking listening
	io.Copy(os.Stdout, c.conn)
}

func (c *Client) UpdateName() bool {
	fmt.Println(">>>>>请输入用户名:")
	fmt.Scanln(&c.Name)

	sendMsg := "rename|" + c.Name + "\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

func (c *Client) menu() bool {
	var flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更改用户名")
	fmt.Println("0.退出")

	_, err := fmt.Scanln(&flag)
	if err != nil {
		fmt.Println("输入有误！")
		return false
	}
	if flag >= 0 && flag <= 3 {
		c.flag = flag
		return true
	} else {
		fmt.Println(">>>>>请输入合法范围内的数字<<<<<")
		return false
	}
}

func (c *Client) PublicChat() {
	//Prompt the user to send a message
	var chatMsg string

	fmt.Println(">>>>>请输入聊天内容，exit退出.")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		//send to server
		sendMsg := chatMsg + "\n"
		_, err := c.conn.Write([]byte(sendMsg))
		if err != nil {
			fmt.Println("conn Write err:", err)
			break
		}

		chatMsg = ""
		fmt.Println(">>>>>请输入聊天内容，exit退出.")
		fmt.Scanln(&chatMsg)
	}
}

func (c *Client) Run() {
	for c.flag != 0 {
		for c.menu() != true {
		}

		switch c.flag {
		case 1:
			c.PublicChat()
		case 2:
			fmt.Println("私聊模式选择...")
		case 3:
			c.UpdateName()
		}
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "host", "127.0.0.1", "设置服务器IP地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(默认是8888)")
}

func main() {
	//command line parsing
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>> 链接服务器失败...")
		return
	}

	//Start a goroutine to process the server's receipt message
	go client.DealResponse()

	fmt.Println(">>>>> 链接服务器成功")

	//Start the client's business
	client.Run()
}
