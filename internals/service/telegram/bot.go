package telegram

import (
	"among-us-roombot/internals/repository"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Команды бота
const (
	cmdStart       = "start"
	cmdList        = "list"
	cmdAdd         = "add"
	cmdDel         = "del"
	cmdEdit        = "edit"
	cmdRules       = "rules"
	cmdHelp        = "help"
	cmdAbout       = "about"
	cmdFeedback    = "feedback"
	cmdSubscribe   = "subscribe"
	cmdUnsubscribe = "unsubscribe"
	cmdNotify      = "notify"
)

type Telegram struct {
	bot *tgbotapi.BotAPI
	rep *repository.Repository
}

func NewTelegram(bot *tgbotapi.BotAPI, rep *repository.Repository) *Telegram {
	return &Telegram{bot, rep}
}

func (t *Telegram) Start() {
	slog.Info("Authorized", slog.String("account", t.bot.Self.UserName))
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := t.bot.GetUpdatesChan(u)
	t.handleUpdates(updates)
}
