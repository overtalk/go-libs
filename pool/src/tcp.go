package src

import (
	"net"
	"time"
)

// tcpSender defines the tcp sender
type tcpSender struct {
	addr string
	pool Pool
}

func (u *tcpSender) factory() (interface{}, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", u.addr)
	if err != nil {
		return nil, err
	}
	return net.DialTCP("tcp", nil, tcpAddr)
}

func close(v interface{}) error {
	return v.(*net.TCPConn).Close()
}

func ping(v interface{}) error {
	return nil
}

// TCPSender defines the constructor of tcp sender
func TCPSender(addr string) (*tcpSender, error) {
	s := &tcpSender{
		addr: addr,
	}

	p, err := NewChannelPool(&Config{
		InitialCap: 5,
		MaxCap:     30,
		Factory:    s.factory,
		Close:      close,
		//Ping:       ping,
		//连接最大空闲时间，超过该时间的连接 将会关闭，可避免空闲时连接EOF，自动失效的问题
		IdleTimeout: 15 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	s.pool = p
	return s, nil
}

// Send defines the method to push data to log server
func (u *tcpSender) Send(data []byte) error {
	c, err := u.pool.Get()
	if err != nil {
		return err
	}

	defer u.pool.Put(c)

	_, err = c.(*net.TCPConn).Write(data)
	return err
}
