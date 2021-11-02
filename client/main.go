package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	"github.com/pion/dtls/v2"
)

func main() {
	log.SetFlags(log.Lshortfile)

	// Certificate setup
	cert, err := ioutil.ReadFile("certs/server.pub.pem")
	if err != nil {
		log.Fatal("Couldn't load file", err)
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(cert)

	conf := &tls.Config{
		RootCAs: certPool,
	}

	// DTLS setup
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dtlsConfig := &dtls.Config{
		ExtendedMasterSecret: dtls.RequireExtendedMasterSecret,
		RootCAs:              certPool,
	}
	dtlsConn, err := dtls.DialWithContext(ctx, "udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 4440}, dtlsConfig)
	defer func() {
		dtlsConn.Close()
	}()

	// TLS setup
	tlsConn, err := tls.Dial("tcp", "127.0.0.1:4444", conf)
	if err != nil {
		log.Println(err)
		return
	}
	defer tlsConn.Close()

	tlsChan := make(chan string, 16)
	dtlsChan := make(chan string, 16)
	go handleConnectionClient(tlsConn, "tls", tlsChan)
	go handleConnectionClient(dtlsConn, "dtls", dtlsChan)

	for {
		// Send stdin on the connection
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		tlsChan <- input
		dtlsChan <- input
	}
}

func handleConnectionClient(conn net.Conn, logmsg string, inputChan chan string) {
	for {
		select {
		case input := <-inputChan:
			n, err := conn.Write([]byte(logmsg + " " + input))
			if err != nil {
				log.Println(n, err)
				return
			}

			buf := make([]byte, 100)
			n, err = conn.Read(buf)
			if err != nil {
				log.Println(n, err)
				return
			}

			log.Print(string(buf[:n]))
		}
	}
}
