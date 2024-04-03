package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

type TargetAddr struct {
	Name string // fully-qualified domain name
	IP   net.IP
	Port int
}

func handleConnection(clientConn net.Conn) {
	// socks5握手
	handshake, addr, err := Handshake(clientConn)
	if err != nil {
		fmt.Printf("Error Handshake: %s\n", err)
		return
	}

	var targetStr string
	if addr.IP != nil {
		targetStr = fmt.Sprintf("%s:%d", addr.IP, addr.Port)
	} else {
		targetStr = fmt.Sprintf("%s:%d", addr.Name, addr.Port)
	}

	log.Println("targetStr=", targetStr)
	// 连接目标服务器
	targetConn, err := net.Dial("tcp", targetStr)
	if err != nil {
		fmt.Printf("Error connecting to target: %s\n", err)
		return
	}
	defer targetConn.Close()

	// 启动 goroutine 将从客户端读取的数据转发给目标服务器
	go func() {
		if _, err := io.Copy(targetConn, handshake); err != nil {
			fmt.Printf("Error copying from client to target: %s\n", err)
		}
	}()

	// 将从目标服务器读取的数据转发给客户端
	if _, err := io.Copy(handshake, targetConn); err != nil {
		fmt.Printf("Error copying from target to client: %s\n", err)
	}
}

func main() {
	// 代理监听地址
	listenAddr := "0.0.0.0:8080"

	// 监听客户端连接
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Printf("Error listening: %s\n", err)
		return
	}
	defer listener.Close()

	fmt.Printf("Proxy listening on %s\n", listenAddr)

	for {
		// 接受客户端连接
		clientConn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %s\n", err)
			continue
		}
		fmt.Printf("Accepted connection from %s\n", clientConn.RemoteAddr())

		// 处理连接
		go handleConnection(clientConn)
	}
}
