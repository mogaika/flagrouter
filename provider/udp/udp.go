package udp

import (
	"bufio"
	"fmt"
	"log"
	"net"

	"github.com/mogaika/flagrouter/provider"
	"github.com/mogaika/flagrouter/router"
)

type UdpProvider struct {
	r          *router.Router
	ListenerHi *net.UDPConn
	ListenerLo *net.UDPConn
}

func serverConnectionThread(rout *router.Router, conn *net.UDPConn, priority byte) {
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

func (p *UdpProvider) newUdpPrioritedProvider(priority byte, addr string) (*net.UDPConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	go serverConnectionThread(p.r, udpConn, priority)

	return udpConn, nil
}

func (p *UdpProvider) Init(r *router.Router) (err error) {
	p.r = r
	p.ListenerHi, err = p.newUdpPrioritedProvider(router.PRIORITY_HIGH, ":5555")
	if err != nil {
		return fmt.Errorf("Error listen hipri: %v", err)
	}

	p.ListenerLo, err = p.newUdpPrioritedProvider(router.PRIORITY_LOW, ":4444")
	if err != nil {
		return fmt.Errorf("Error listen lopri: %v", err)
	}
	return nil
}

func init() {
	provider.RegisterProvider("udp", &UdpProvider{})
}
