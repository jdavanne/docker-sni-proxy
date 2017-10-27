package main

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
)

type Session struct {
	id int64
	wg sync.WaitGroup
	c1 net.Conn
	c2 net.Conn
}

var globalSessionID int64

func NewSession() *Session {
	var session Session
	g := atomic.AddInt64(&globalSessionID, 1)
	session.id = g //FIXME: race?
	return &session
}

func (session *Session) forwardHalf(c1 net.Conn, c2 net.Conn) {
	defer c1.Close()
	defer c2.Close()
	defer session.wg.Done()
	io.Copy(c1, c2)
}

func (session *Session) Close() {
	if session.c1 != nil {
		session.c1.Close()
	}
	if session.c2 != nil {
		session.c2.Close()
	}
}

func (session *Session) streamToConn(c1, c2 net.Conn) {
	log.Println("Session", session.id, "- Streaming...", session.c1.RemoteAddr(), session.c2.RemoteAddr().String())
	session.wg.Add(2)
	go session.forwardHalf(c1, c2)
	go session.forwardHalf(c2, c1)
	session.wg.Wait()
	log.Println("Session", session.id, "- Closed", c1.RemoteAddr().String())
}

func (session *Session) ServerDispatch(conn net.Conn, tls bool, mode string) {
	b2 := make([]byte, 17000)
	//time.Sleep(100 * time.Millisecond) //FIXME: trick to accumulate data...
	var err error
	for {
		n, err := conn.Read(b2) //FIXME: should read the full TLS first packet and the full first HTTP packet...
		if err != nil {
			log.Errorln("Session", session.id, "- Read Error", err)
			conn.Close()
			return
		}
		b2 = b2[:n]
		if tls {
			if b2[0] != 22 {
				log.Errorln("Session", session.id, "- Not TLS Handshake")
				conn.Close()
				return
			}
			if n < 5 {
				log.Errorln("Session", session.id, "- No header (No luck?)")
				conn.Close()
				return
			}
			version := fmt.Sprintf("%d.%d", b2[2], b2[1])
			var size int
			size = 256*int(b2[3]) + int(b2[4])
			log.Println("Session", session.id, "- Hello TLS Packet, version:", version, "size:", size+5, n)
			break
		} else {
			if strings.Contains(string(b2[:n]), "\r\n\r\n") {
				break
			}
			log.Errorln("Session", session.id, "- Missing HTTP header trailer", n)
			conn.Close()
			return
		}
	}

	var hostname string
	if tls {
		hostname, err = GetHostname(b2)
		if err != nil {
			log.Errorln("Session", session.id, "- Not SSL Hello :", err, string(b2))
			conn.Close()
			return
		}
	} else {
		hostname, err = GetHostnameHTTP(string(b2))
		if err != nil {
			log.Errorln("Session", session.id, "- Not HTTP1.1 :", err, string(b2))
			conn.Close()
			return
		}
	}

	parts := strings.Split(hostname, ".")
	if (mode == "stack" && len(parts) < 4) || (mode == "service" && len(parts) < 3) {
		log.Errorln("Session", session.id, "- SNI too short ", hostname)
		conn.Close()
		return
	}

	var host string
	if mode == "stack" {
		host = parts[1] + "_" + parts[0]
	} else if mode == "service" {
		host = parts[0]
	} else {
		log.Fatal("mode not supported")
	}

	var port string
	if tls {
		if p, ok := TlsPorts[host]; ok {
			port = p
		} else {
			port = "443"
		}
	} else {
		if p, ok := HttpPorts[host]; ok {
			port = p
		} else {
			port = "80"
		}
	}

	addr := host + ":" + port

	log.Println("Session", session.id, "- Dialing...", addr, hostname)
	client, err := net.Dial("tcp", addr)
	if err != nil {
		log.Errorln("Session", session.id, "- Dial failed :", addr, err)
		session.c1.Close()
		return
	}
	session.c2 = client

	_, err = client.Write(b2)
	if err != nil {
		log.Errorln("Error putting back data in tunnel", err)
		session.c1.Close()
		session.c2.Close()
		return
	}
	session.streamToConn(session.c1, session.c2)
	log.Println("Session", session.id, "- Done")
}
