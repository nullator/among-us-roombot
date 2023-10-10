package service

import (
	"among-us-roombot/internals/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Service struct {
	Bot
}

type Bot interface {
	Start()
}

func NewService(bot *tgbotapi.BotAPI, rep *repository.Repository) *Service {
	return &Service{
		Bot: NewBotService(bot, rep),
	}
}
