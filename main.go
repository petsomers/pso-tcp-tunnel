package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

type Tunnel struct {
	listenOn  int
	forwardTo string
}

func main() {
	args := os.Args
	var tunnelCount = len(args) - 1
	tunnels := make([]Tunnel, tunnelCount)
	for i, a := range args {
		fmt.Printf("arg %d: %s\r\n", i, a)
		if i == 0 {
			continue
		}
		s := strings.Split(a, ":")
		if len(s) != 3 {
			fmt.Printf("Tunnel %d\r\n", i)
			fmt.Printf("Invalid parameter %d: %s", i, a)
			fmt.Println("Usage port:destinaionhost:destinationport")
			return
		}
		listenOnPort, err := strconv.Atoi(s[0])
		if err != nil {
			fmt.Printf("Tunnel %d\r\n", i)
			fmt.Printf("Invalid port number %s", s[0])
			return
		}
		destinationPort, err := strconv.Atoi(s[2])
		if err != nil {
			fmt.Printf("Tunnel %d\r\n", i)
			fmt.Printf("Invalid destination port number %s", s[2])
			return
		}
		tunnel := Tunnel{
			listenOn:  listenOnPort,
			forwardTo: fmt.Sprintf("%s:%d", s[1], destinationPort),
		}
		tunnels[i-1] = tunnel
	}
	for _, tunnel := range tunnels {
		go tunnel.startTunnel()
	}
	select {}
}

func (t Tunnel) startTunnel() {
	listenOn := fmt.Sprintf("0.0.0.0:%d", t.listenOn)
	listener, err := net.Listen("tcp4", listenOn)
	fmt.Printf("Server running on %s", listenOn)
	if err != nil {
		fmt.Println(err)
	}
	defer listener.Close()
	for {
		incomingConnection, err := listener.Accept()
		fmt.Printf("Incoming connection from %s\r\n", incomingConnection.RemoteAddr().String())
		if err != nil {
			fmt.Println(err)
		}
		go func(incomingConnection net.Conn) {
			defer incomingConnection.Close()
			fmt.Printf("Connecting to %s\r\n", t.forwardTo)
			outConnection, err := net.Dial("tcp4", t.forwardTo)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer outConnection.Close()
			go io.Copy(outConnection, incomingConnection)
			io.Copy(incomingConnection, outConnection)
			fmt.Printf("Closing connection for tunnel %d:%s\r\n", t.listenOn, t.forwardTo)

		}(incomingConnection)

	}
}
