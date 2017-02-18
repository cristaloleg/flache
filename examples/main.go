package main

import (
	"fmt"
	"github.com/cristaloleg/flache"
	"net"
	"net/http"
)

var tcpCache flache.Cacher

func main() {
	go tcpListen()
	go httpListen()
}

func httpListen() {
	http.HandleFunc("/", handleHTTPConnection)
	http.ListenAndServe(":8080", nil)
}

func handleHTTPConnection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
	case "PUT":
	case "HEAD":
	}
	//fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func tcpListen() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Something bad happend: %s", err)
		}
		go handleTCPConnection(conn)
	}
}

func handleTCPConnection(conn net.Conn) {
	var buffer []byte
	n, err := conn.Read(buffer)
	if err != nil || n <= 0 {
		return
	}
	switch buffer[0] {
	case 1:
		// add
	case 2:
		// put
	case 3:
		// check
	}

}
