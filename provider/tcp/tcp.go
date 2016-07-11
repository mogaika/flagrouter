package tcp

import (
	"bufio"
	"fmt"
	"log"
	"net"

	"github.com/mogaika/flagrouter/provider"
	"github.com/mogaika/flagrouter/router"
)

type TcpProvider struct {
	r          *router.Router
	ListenerHi *net.TCPListener
	ListenerLo *net.TCPListener
}

func serverConnectionThread(rout *router.Router, conn net.Conn, priority byte) {
	r := bufio.NewReader(conn)
	log.Printf("Connection listen [%s] on %s", conn.LocalAddr().Network(), conn.LocalAddr().String())
	for {
		str, err := r.ReadString('\n')
		if err != nil {
			log.Printf("Closing connection [%s] from %s on %s: %v", conn.LocalAddr().Network(), conn.RemoteAddr().String(), conn.LocalAddr().String(), err)
			conn.Close()
			return
		}

		if err := rout.AddToQueue(priority, str); err != nil {
			log.Printf("Error when queuing flag '%s': %v", str, err)
		}

		log.Printf("info from socket: %v", str)
	}
}

func serverProviderThread(rout *router.Router, l *net.TCPListener, priority byte) {
	log.Printf("Created listener [%s] on %s", l.Addr().Network(), l.Addr().String())
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Error when accepting [%s] connection from %s on %s: %v", l.Addr().Network(), conn.RemoteAddr().String(), l.Addr().String(), err)
			return
		}

		go serverConnectionThread(rout, conn, priority)
	}
}

func (p *TcpProvider) newTcpPrioritedProvider(priority byte, addr string) (*net.TCPListener, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	tcpListen, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}

	go serverProviderThread(p.r, tcpListen, priority)

	return tcpListen, nil
}

func (p *TcpProvider) Init(r *router.Router) (err error) {
	p.r = r
	p.ListenerHi, err = p.newTcpPrioritedProvider(router.PRIORITY_HIGH, ":5555")
	if err != nil {
		return fmt.Errorf("Error listen hipri: %v", err)
	}

	p.ListenerLo, err = p.newTcpPrioritedProvider(router.PRIORITY_LOW, ":4444")
	if err != nil {
		return fmt.Errorf("Error listen lopri: %v", err)
	}
	return nil
}

func init() {
	provider.RegisterProvider("tcp", &TcpProvider{})
}
