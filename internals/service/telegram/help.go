package telegram

import (
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Telegram) handleHelp(message *tgbotapi.Message) error {
	const path = "service.telegram.help"

	msg := tgbotapi.NewMessage(message.Chat.ID, "Список команд:\n"+
		"/start - начать работу с ботом\n"+
		"/add - добавить комнату\n"+
		"/del - удалить комнату\n"+
		"/rules - правила игры\n"+
		"/help - список команд\n")
	_, err := b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	slog.Info("Выполнена команда help")
	return nil
}
