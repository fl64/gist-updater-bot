package app

import (
	"context"
	"database/sql"
	"github.com/fl64/gist-updater-bot/internal/cfg"
	"github.com/fl64/gist-updater-bot/internal/storage"
	"github.com/fl64/gist-updater-bot/internal/tgbot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

type BotApp struct {
	cfg *cfg.Cfg
}

func NewBotApp(cfg *cfg.Cfg) *BotApp {
	return &BotApp{
		cfg,
	}
}

func (App *BotApp) Run(ctx context.Context) error {

	db, err := sql.Open("sqlite3", App.cfg.DbFile)
	if err != nil {
		return err
	}
	defer func() {
		err := db.Close()
		if err != nil {
			log.Fatal("Can't close DB", err)
		}
	}()

	s := storage.NewStorage(db)
	err = s.Init()
	if err != nil {
		return err
	}
	botAPI, err := tgbotapi.NewBotAPI(App.cfg.TgToken)
	if err != nil {
		return err
	}

	botAPI.Debug = App.cfg.Debug
	m := tgbot.NewMessageProcessor(s, botAPI)
	err = m.Serve(ctx)
	if err != nil {
		return err
	}
	return nil
}
