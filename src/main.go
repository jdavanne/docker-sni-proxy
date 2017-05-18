package main

import (
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/namsral/flag"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var Version string
var Build string
var Date string

var host string
var port int

func serve(l net.Listener) {
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine
		session := NewSession()
		session.c1 = conn
		log.Println("Session", session.id, "- Incoming connection from ", conn.RemoteAddr().String())
		go session.ServerDispatch(conn)
	}
}

func main() {
	formatter := new(prefixed.TextFormatter)
	formatter.DisableTimestamp = false
	formatter.FullTimestamp = true
	formatter.TimestampFormat = "2006-01-02 15:04:05.000000000"
	log.SetFormatter(formatter)
	log.SetLevel(log.DebugLevel)

	flag.IntVar(&port, "port", 443, "Specify the port to listen to.")
	flag.StringVar(&host, "host", "0.0.0.0", "Specify the interface to listen to.")

	flag.Parse()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Starting docker-sni-proxy - version:", Version, " build:", Build, " date:", Date)

	// Listen for incoming connections.
	addr := host + ":" + strconv.Itoa(port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	log.Println("Listening on " + addr)

	go serve(l)

	sig := <-sigs
	log.Println()
	log.Println(sig)
}
