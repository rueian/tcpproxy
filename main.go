package main

import (
	"flag"
	"log"
	"net"
)

func main() {
	bind := flag.String("bind", "", "local bind address, ex: 127.0.0.1:5432")
	dest := flag.String("dest", "", "target tcp address, ex: 192.0.0.1:5432")
	flag.Parse()

	ln, err := net.Listen("tcp", *bind)
	if err != nil {
		log.Fatalf(`fail to listen bind address "%s": %s`, *bind, err.Error())
	}
	defer ln.Close()

	for {
		client, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			defer client.Close()

			server, err := net.Dial("tcp", *dest)
			if err != nil {
				log.Printf(`fail to connect dest address "%s": %s`, *dest, err.Error())
				return
			}
			defer server.Close()

			ch1 := make(chan error)
			ch2 := make(chan error)
			defer close(ch1)
			defer close(ch2)

			go func() { ch1 <- proxy(client, server) }()
			go func() { ch2 <- proxy(server, client) }()

			if err := <-ch1; err != nil {
				log.Println(err.Error())
			}
			if err := <-ch2; err != nil {
				log.Println(err.Error())
			}
		}()
	}
}

func proxy(in net.Conn, out net.Conn) (err error) {
	_, err = out.(*net.TCPConn).ReadFrom(in)
	return err
}