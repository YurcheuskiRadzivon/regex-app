package server

import (
	"net"
	"time"
)

type Option func(*Server)

func Port(port string) Option {
	return func(s *Server) {
		s.address = net.JoinHostPort("", port)
	}
}

func ReadTimeout(readTimeout time.Duration) Option {
	return func(s *Server) {
		s.readTimeout = readTimeout
	}
}

func WriteTimeout(writeTimeout time.Duration) Option {
	return func(s *Server) {
		s.writeTimeout = writeTimeout
	}
}

func ShutdownTimeout(shutdownTimeout time.Duration) Option {
	return func(s *Server) {
		s.shutdownTimeout = shutdownTimeout
	}
}
