package main

import (
	// "fmt"
	// "io"

	"log"
	"net"
	"net/http"
	"os"
	"time"
)

func helloServer(w http.ResponseWriter, req *http.Request) {
	log.Println("Hello: ", os.Getenv("NAME"))
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("=" + os.Getenv("NAME") + "=\n"))
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

func listenAndServe(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Println("Listening on", addr)
	srv := &http.Server{Addr: addr, Handler: nil}
	return srv.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)})
}

func main() {
	http.HandleFunc("/", helloServer)
	port := os.Getenv("PORT")
	err := listenAndServe(":" + port)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
