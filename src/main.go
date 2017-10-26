package main

import (
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/namsral/flag"
	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var Version string
var Build string
var Date string

func serve(l net.Listener, tls bool, mode string) {
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
		go session.ServerDispatch(conn, tls, mode)
	}
}

func listen(host string, port int, why string) net.Listener {
	addr := host + ":" + strconv.Itoa(port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	log.Println("Listening " + why + " on " + addr)
	return l
}

func main() {
	formatter := new(prefixed.TextFormatter)
	formatter.DisableTimestamp = false
	formatter.FullTimestamp = true
	formatter.TimestampFormat = "2006-01-02 15:04:05.000000000"

	log.SetFormatter(formatter)
	log.SetLevel(log.DebugLevel)

	var host string
	var tlsPort int
	var httpPort int
	var mode string
	var publicNetwork string

	flag.IntVar(&tlsPort, "tls-port", 443, "Specify the port to listen to.")
	flag.IntVar(&httpPort, "http-port", 80, "Specify the port to listen to.")
	flag.StringVar(&host, "host", "0.0.0.0", "Specify the interface to listen to.")
	flag.StringVar(&mode, "mode", "stack", "Specify the mode : stack, service")
	flag.StringVar(&publicNetwork, "docker-network", "public", "Specify the public network for docker")
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Starting docker-sni-proxy - version:", Version, " build:", Build, " date:", Date)

	// Listen for incoming connections.

	// Close the listener when the application closes.
	l := listen(host, tlsPort, "TLS")
	defer l.Close()
	go serve(l, true, mode)

	http := listen(host, httpPort, "HTTP")
	defer http.Close()
	go serve(http, false, mode)

	if publicNetwork != "" {
		go DockerInit(publicNetwork)
	}

	sig := <-sigs
	log.Println()
	log.Println(sig)
}
