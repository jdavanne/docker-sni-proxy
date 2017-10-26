package main

import (
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

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
	b2 := make([]byte, 1000)
	time.Sleep(100 * time.Millisecond)
	n, err := conn.Read(b2)
	if err != nil {
		conn.Close()
		return
	}
	b2 = b2[:n]
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
