package service

import (
	"among-us-roombot/internals/repository"
	"among-us-roombot/internals/service/telegram"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotService struct {
	tg  *tgbotapi.BotAPI
	bot *telegram.Telegram
}

func NewBotService(tg *tgbotapi.BotAPI, rep *repository.Repository) *BotService {
	bot := telegram.NewTelegram(tg, rep)
	return &BotService{tg, bot}
}

func (tg *BotService) Start() {
	tg.bot.Start()
}
