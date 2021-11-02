package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"log"
	"net"
	"time"

	"github.com/pion/dtls/v2"
)

func main() {
	log.SetFlags(log.Lshortfile)

	// Certificate setup
	cer, err := tls.LoadX509KeyPair("certs/server.pub.pem", "certs/server.pem")
	if err != nil {
		log.Println(err)
		return
	}
	certPool := x509.NewCertPool()
	cert, err := x509.ParseCertificate(cer.Certificate[0])
	certPool.AddCert(cert)

	// DTLS setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dtlsConfig := &dtls.Config{
		Certificates:         []tls.Certificate{cer},
		ExtendedMasterSecret: dtls.RequireExtendedMasterSecret,
		ClientCAs:            certPool,
		// Create timeout context for accepted connection.
		ConnectContextMaker: func() (context.Context, func()) {
			return context.WithTimeout(ctx, time.Second)
		},
	}
	dtlsListener, err := dtls.Listen("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 4440}, dtlsConfig)
	defer func() {
		dtlsListener.Close()
	}()

	// TLS setup
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cer},
	}
	tlsListener, err := tls.Listen("tcp", ":4444", tlsConfig)
	if err != nil {
		log.Println(err)
		return
	}
	defer tlsListener.Close()

	// Handle connections
	go func() {
		for {
			tlsConn, err := tlsListener.Accept()
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println("tls connection: ", tlsConn.RemoteAddr())
			go handleConnectionServer(tlsConn, "tls")
		}
	}()
	for {
		dtlsConn, err := dtlsListener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("dtls connection: ", dtlsConn.RemoteAddr())
		go handleConnectionServer(dtlsConn, "dtls")
	}
}

func handleConnectionServer(conn net.Conn, logmsg string) {
	defer conn.Close()
	r := bufio.NewReader(conn)

	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// Handle disconnect on tls
				// dtls connection is connectionless, so a manual timeout (using
				// select with a go timer, pinging the addr etc) is needed.
				// dtls connection could also be closed when tls connection
				// is closed, but a way to pair the connections together would
				// need to be created and that is out of scope for this example
				log.Println(logmsg+" disconnection: ", conn.RemoteAddr())
			} else {
				log.Println(err)
			}

			return
		}

		log.Print(conn.RemoteAddr().String() + " -> " + msg)

		n, err := conn.Write([]byte(logmsg + " response " + msg))
		if err != nil {
			log.Println(n, err)
			return
		}
	}
}
