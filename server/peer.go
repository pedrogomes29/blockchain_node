package server

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
)

type peer struct{
	conn     net.Conn
	commands chan<- command
}

func (p *peer) GetAddress() string{
	if addr, ok := p.conn.RemoteAddr().(*net.TCPAddr); ok {
		return addr.IP.String();
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



func (p * peer) ReadInput() {
	for {
		msg, err := bufio.NewReader(p.conn).ReadBytes('\n')
		if err != nil {
			return
		}

		args := bytes.Split(msg, []byte(" "))
		cmd := string(args[0])

		switch cmd {
			case "GET_ADDR":
				p.commands <- command{
					id:     GET_ADDR,
					peer:  p,
					args: args[1:],
				}
			case "ADDR":
				p.commands <- command{
					id:     VERSION,
					peer:  p,
					args: args[1:],
				}
			case "VERSION":
				p.commands <- command{
					id:     VERSION,
					peer:  p,
					args: args[1:],
				}
			case "VERSION_ACK":
				p.commands <- command{
					id:     VERSION_ACK,
					peer:  p,
					args: args[1:],
			}
			default:
				p.send([]byte(fmt.Errorf("unknown command: %s", cmd).Error()))
			}
		}
}


func (c *peer) send(msg []byte) {
	c.conn.Write(msg)
}