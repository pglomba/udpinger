package server

import (
	"log/slog"
	"net"
)

type Server struct {
	localAddress *net.UDPAddr
	bufferSize   int
}

func NewServer(address string, bufferSize int) (*Server, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	return &Server{
		localAddress: udpAddr,
		bufferSize:   bufferSize,
	}, nil
}

func (s *Server) Start() error {
	conn, err := net.ListenUDP("udp", s.localAddress)
	if err != nil {
		return err
	} else {
		slog.Info("Starting server on " + s.localAddress.String())
	}

	defer conn.Close()

	for {
		buffer := make([]byte, s.bufferSize)
		_, clientAddress, err := conn.ReadFromUDP(buffer)
		if err != nil {
			slog.Error(err.Error())
		}

		go responseHandler(conn, clientAddress, buffer)
	}
}

func responseHandler(conn net.PacketConn, address net.Addr, payload []byte) {
	_, err := conn.WriteTo(payload, address)
	if err != nil {
		slog.Error(err.Error())
	}
}
