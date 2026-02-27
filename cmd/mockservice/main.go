package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {
	port := flag.Int("port", 22, "port to listen on (should match server.target_port)")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("failed to listen", "addr", addr, "error", err)
		os.Exit(1)
	}
	defer ln.Close()

	slog.Info("mockservice listening", "addr", addr)
	fmt.Println("---")
	fmt.Println("Waiting for connections from the proxy...")
	fmt.Println("Every incoming TCP connection will be printed below.")
	fmt.Println("---")

	var connID atomic.Uint64

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		slog.Info("shutting down mockservice")
		ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		id := connID.Add(1)
		go handleConn(conn, id)
	}
}

func handleConn(conn net.Conn, id uint64) {
	defer conn.Close()
	remote := conn.RemoteAddr().String()
	ts := time.Now().Format("15:04:05")

	fmt.Printf("\n[#%d] %s  New connection from %s\n", id, ts, remote)

	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			data := buf[:n]
			ts = time.Now().Format("15:04:05")
			fmt.Printf("[#%d] %s  Received %d bytes:\n", id, ts, n)

			if isPrintable(data) {
				fmt.Printf("%s\n", data)
			} else {
				fmt.Print(hex.Dump(data))
			}
		}
		if err != nil {
			if err != io.EOF {
				fmt.Printf("[#%d] %s  Read error: %v\n", id, time.Now().Format("15:04:05"), err)
			}
			fmt.Printf("[#%d] %s  Connection closed\n", id, time.Now().Format("15:04:05"))
			return
		}
	}
}

func isPrintable(data []byte) bool {
	for _, b := range data {
		if b < 0x20 && b != '\n' && b != '\r' && b != '\t' {
			return false
		}
	}
	return true
}
