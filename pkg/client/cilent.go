package client

import (
	"fmt"
	"net"
)

type Client struct {
	Host string
	Port int

	conn *net.UDPConn
}

func NewClient(Host string, Port int) *Client {
	return &Client{}
}

func (c *Client) Connect() error {
	conncetionString := fmt.Sprintf("%s:%d", c.Host, c.Port)
	s, err := net.ResolveUDPAddr("udp4", conncetionString)
	if err != nil {
		return err
	}
	c.conn, err = net.DialUDP("udp", nil, s)
	if err != nil {
		return err
	}
	return nil
}
