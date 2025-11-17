package server

import (
	"time"

	"github.com/goccy/go-json"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

const (
	_defaultAddr            = ":80"
	_defaultReadTimeout     = 40 * time.Second
	_defaultWriteTimeout    = 40 * time.Second
	_defaultShutdownTimeout = 40 * time.Second
)

type Server struct {
	App             *fiber.App
	notify          chan error
	address         string
	readTimeout     time.Duration
	writeTimeout    time.Duration
	shutdownTimeout time.Duration
}

func New(opts ...Option) *Server {
	s := &Server{
		App:             nil,
		notify:          make(chan error, 1),
		address:         _defaultAddr,
		readTimeout:     _defaultReadTimeout,
		writeTimeout:    _defaultWriteTimeout,
		shutdownTimeout: _defaultShutdownTimeout,
	}

	for _, opt := range opts {
		opt(s)
	}

	engine := html.New("./templates", ".html")

	app := fiber.New(fiber.Config{
		Views:        engine,
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
	})

	s.App = app
	return s
}

func (s *Server) Start() {
	go func() {
		s.notify <- s.App.Listen(s.address)
		close(s.notify)
	}()
}

func (s *Server) Notify() <-chan error {
	return s.notify
}

func (s *Server) Shutdown() error {
	return s.App.ShutdownWithTimeout(s.shutdownTimeout)
}
