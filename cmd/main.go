package main

import (
	"context"
	"github.com/fl64/gist-updater-bot/internal/app"
	"github.com/fl64/gist-updater-bot/internal/cfg"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		s := <-sigChan
		log.Println("Catch signal: ", s)
		cancel()
	}()

	config, err := cfg.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	botApp := app.NewBotApp(config)
	err = botApp.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Bot stopped")
}
