package server

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
)

type TCPProxy struct {
	store      *IPStore
	targetPort int
}

func NewTCPProxy(store *IPStore, targetPort int) *TCPProxy {
	return &TCPProxy{store: store, targetPort: targetPort}
}

func (p *TCPProxy) HandleConn(clientConn net.Conn) {
	defer clientConn.Close()

	ip, _, ok := p.store.Get()
	if !ok {
		slog.Warn("proxy: no registered ip, rejecting connection", "client", clientConn.RemoteAddr())
		return
	}

	target := fmt.Sprintf("%s:%d", ip, p.targetPort)
	backendConn, err := net.Dial("tcp", target)
	if err != nil {
		slog.Error("proxy: failed to dial target", "target", target, "error", err)
		return
	}
	defer backendConn.Close()

	slog.Info("proxy: connection established", "client", clientConn.RemoteAddr(), "target", target)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(backendConn, clientConn)
		if tc, ok := backendConn.(*net.TCPConn); ok {
			tc.CloseWrite()
		}
	}()

	go func() {
		defer wg.Done()
		io.Copy(clientConn, backendConn)
		if tc, ok := clientConn.(*net.TCPConn); ok {
			tc.CloseWrite()
		}
	}()

	wg.Wait()
	slog.Info("proxy: connection closed", "client", clientConn.RemoteAddr(), "target", target)
}
