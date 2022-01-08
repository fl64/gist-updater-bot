package tgbot

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/fl64/gist-updater-bot/internal/cfg"
	"github.com/fl64/gist-updater-bot/internal/models"
	"github.com/fl64/gist-updater-bot/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/go-github/v41/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"strings"
	"time"
)

type MessageProcessor struct {
	storage *storage.Storage
	cfg     *cfg.Cfg
	bot     *tgbotapi.BotAPI
}

func NewMessageProcessor(s *storage.Storage, bot *tgbotapi.BotAPI) *MessageProcessor {
	m := &MessageProcessor{
		storage: s,
		bot:     bot,
	}
	return m
}

func (mp *MessageProcessor) Serve(ctx context.Context) error {

	log.Printf("Authorized on account %s", mp.bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	//u.Timeout = mp.cfg.Timeout

	updates := mp.bot.GetUpdatesChan(u)
	for {
		select {
		case update := <-updates:
			err := mp.checkMsg(&update)
			if err != nil {
				log.Error(err)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (mp *MessageProcessor) sendMsg(chat int64, msg string) error {
	m := tgbotapi.NewMessage(chat, msg)
	if _, err := mp.bot.Send(m); err != nil {
		return err
	}
	return nil
}

func (mp *MessageProcessor) updateGist(msg string, settings *models.Settings) error {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: settings.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	gist, _, err := client.Gists.Get(ctx, settings.GistID)
	if err != nil {
		return err
	}

	fileName := github.GistFilename(settings.GistFile)
	file := gist.GetFiles()[fileName]
	var s string
	if file.Content != nil {
		s = *file.Content + "\n" + msg
	} else {
		s = msg
	}

	gist.Files[fileName] = github.GistFile{
		Content: &s,
	}
	_, _, err = client.Gists.Edit(ctx, settings.GistID, gist)
	if err != nil {
		return err
	}
	return nil
}

func (mp *MessageProcessor) checkMsg(update *tgbotapi.Update) error {
	if update.Message == nil { // ignore any non-Message updates
		log.Println("Non-message update!")
		return nil
	}
	if update.Message.IsCommand() { // ignore any non-command Messages
		log.Println("Command!")
		switch update.Message.Command() {
		case "gist":
			cmd := strings.Fields(update.Message.Text)
			if len(cmd) < 4 {
				err := mp.sendMsg(update.Message.Chat.ID, "How to add gist: /gist <token> <gist_id> <gist_filename>")
				if err != nil {
					return err
				}
				return errors.New("not enough arguments for /gist")
			}
			err := mp.storage.SetUserSettings(update.Message.From.ID, cmd[1], cmd[2], cmd[3])
			if err != nil {
				return err
			}
			return nil
		case "get":
			settings, err := mp.storage.GetSettings(update.Message.From.ID)
			if err == sql.ErrNoRows {
				err := mp.sendMsg(update.Message.Chat.ID, fmt.Sprintf("No settings for user %s", update.Message.From))
				if err != nil {
					return err
				}
			}
			if err != nil && err != sql.ErrNoRows {
				return err
			}
			log.Println(settings)
		default:
			time.Sleep(time.Second)
		}
	}
	if !update.Message.IsCommand() { // ignore any non-command Messages
		log.Println("Not Command!")
		settings, err := mp.storage.GetSettings(update.Message.From.ID)
		if err != nil {
			return err
		}
		err = mp.updateGist(update.Message.Text, settings)
		if err != nil {
			err := mp.sendMsg(update.Message.Chat.ID, fmt.Sprintf("Can't update gist: %v", err))
			if err != nil {
				return err
			}
			return err
		}
		err = mp.sendMsg(update.Message.Chat.ID, "Gist updated")
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
