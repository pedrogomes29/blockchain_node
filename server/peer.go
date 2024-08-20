package server

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
)

type peer struct {
	conn     net.Conn
	commands chan<- command
}

func (p *peer) GetAddress() string {
	if addr, ok := p.conn.RemoteAddr().(*net.TCPAddr); ok {
		return addr.IP.String()
	}
	return ""
}

func (s *Server) NewPeer(conn net.Conn) *peer {
	log.Printf("new peer has connected: %s\n", conn.RemoteAddr().String())

	return &peer{
		conn:     conn,
		commands: s.commands,
	}
}

func (p *peer) ReadInput() {
	for {
		msg, err := bufio.NewReader(p.conn).ReadBytes('\n')
		if err != nil {
			return
		}
		msg = msg[:len(msg)-1] //removes \n
		//fmt.Printf("received: %s\n",string(msg))
		args := bytes.Split(msg, []byte(" "))
		cmd := string(args[0])

		switch cmd {
		case "GET_ADDR":
			p.commands <- command{
				id:   GET_ADDR,
				peer: p,
				args: args[1:],
			}
		case "ADDR":
			p.commands <- command{
				id:   ADDR,
				peer: p,
				args: args[1:],
			}
		case "VERSION":
			p.commands <- command{
				id:   VERSION,
				peer: p,
				args: args[1:],
			}
		case "VERSION_ACK":
			p.commands <- command{
				id:   VERSION_ACK,
				peer: p,
				args: args[1:],
			}
		default:
			p.sendString(fmt.Errorf("unknown command: %s", cmd).Error())
		}
	}
}

func (c *peer) sendBytes(msg []byte) {
	n, err := c.conn.Write(msg)
	if n != len(msg) {
		log.Fatal("Failed to send message: ", msg)
	}
	if err != nil {
		log.Fatal("Error sending data: ", err)
	}
}

func (c *peer) sendString(msg string) {
	//fmt.Printf("sent: %s\n", msg)
	n, err := c.conn.Write([]byte(string(msg) + "\n"))
	if n != len(msg)+1 {
		log.Fatal("Failed to send message: ", msg)
	}
	if err != nil {
		log.Fatal("Error sending data: ", err)
	}

}
