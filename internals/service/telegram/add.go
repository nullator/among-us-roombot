package telegram

import (
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Telegram) handleAdd(message *tgbotapi.Message) error {
	const path = "service.telegram.add"

	room := message.Text
	err := b.rep.AddRoom(room)
	if err != nil {
		slog.Error("Ошибка добавления комнаты в БД")
		return fmt.Errorf("%s: %w", path, err)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Комната добавлена")
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	slog.Info("Выполнена команда add")
	return nil
}
