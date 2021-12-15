package main

import (
	"log"
	"net"
	"net/http"
	"time"
)

type ProxyServer struct{}

func (*ProxyServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodConnect {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Println("[server] ResponseWriter not support Hijacker")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		log.Fatal("Hijack err: ", err)
	}
	serverConn, err := net.DialTimeout("tcp", req.URL.Host, time.Second*3)
	if err != nil {
		log.Printf("[server] dail %q fail: err: %+v", req.URL.Host, err)
		return
	}
	go pipe(clientConn, serverConn)
}

func chanFromConn(conn net.Conn) <-chan []byte {
	ch := make(chan []byte)
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if n > 0 {
				res := make([]byte, n)
				copy(res, buf)
				ch <- res
			}
			if err != nil {
				ch <- nil
				break
			}
		}
	}()
	return ch
}

func pipe(conn1, conn2 net.Conn) {
	defer func() {
		conn1.Close()
		conn2.Close()
	}()
	chan1 := chanFromConn(conn1)
	chan2 := chanFromConn(conn2)
	for {
		select {
		case b1 := <-chan1:
			if b1 == nil {
				return
			} else {
				if _, err := conn2.Write(b1); err != nil {
					log.Println("write to conn2 err", err)
					return
				}
			}
		case b2 := <-chan2:
			if b2 == nil {
				return
			} else {
				if _, err := conn1.Write(b2); err != nil {
					log.Println("write to conn1 err", err)
					return
				}
			}
		}
	}
}

var _ http.Handler = (*ProxyServer)(nil)
