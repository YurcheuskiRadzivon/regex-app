package main

import (
	"log"
	"os"
	"os/signal"
	"regexp-helper/internal/server"
	"regexp-helper/internal/service"
	"syscall"
)

func main() {
	httpSvr := server.New()

	svc := service.NewRegexService()

	server.NewRouter(
		httpSvr.App,
		svc,
	)

	httpSvr.Start()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		log.Printf("app - Run - signal: %s", s.String())
	case err := <-httpSvr.Notify():
		log.Printf("app - Run - httpSvr.Notify: %v", err)
	}

	err := httpSvr.Shutdown()
	if err != nil {
		log.Printf("app - Run - httpSvr.Shutdown: %v", err)
	}

}
