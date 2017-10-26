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

	// fmt.Fprintf(w, "This is an example server.\n")
	// io.WriteString(w, "This is an example server.\n")
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

func listenAndServeTLS(addr, certFile, keyFile string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log.Println("Listening on", addr)

	srv := &http.Server{Addr: addr, Handler: nil}
	return srv.ServeTLS(tcpKeepAliveListener{ln.(*net.TCPListener)}, certFile, keyFile)
}

func main() {
	http.HandleFunc("/", helloServer)
	port := os.Getenv("PORT")
	err := listenAndServeTLS(":"+port, "server.crt", "server.key")
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
