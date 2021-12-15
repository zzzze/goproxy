package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

func startServer() {
	log.Println("start proxy server on port 9999")
	if err := http.ListenAndServe(":9999", &ProxyServer{}); err != nil {
		log.Println(err)
	}
}

func main() {
	go func() {
		time.Sleep(time.Second)
		conn, err := net.Dial("tcp", ":9999")
		if err != nil {
			log.Fatal("network err: ", err)
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		if _, err := io.WriteString(conn, fmt.Sprintf("%s %s HTTP/1.1\n\n", http.MethodConnect, "www.baidu.com:80")); err != nil {
			log.Fatal("connect err: ", err)
		}
		resp, err := http.ReadResponse(reader, nil)
		if err != nil {
			log.Fatal("get err: ", err)
		}
		log.Println("status: ", resp.Status)
		if _, err := io.WriteString(conn, fmt.Sprintf("%s %s HTTP/1.1\n\n", http.MethodGet, "https://www.baidu.com")); err != nil {
			log.Fatal("connect err: ", err)
		}
		resp, err = http.ReadResponse(reader, nil)
		if err != nil {
			log.Fatal("get err: ", err)
		}
		defer resp.Body.Close()
		log.Println("status: ", resp.Status)
		if data, err := io.ReadAll(resp.Body); err != nil {
			log.Println("read body err: ", err)
		} else {
			log.Println("body: ", string(data))
		}
	}()
	startServer()
}
