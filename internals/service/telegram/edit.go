package telegram

import (
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Telegram) handleEdit(message *tgbotapi.Message) error {
	const path = "service.telegram.edit"

	room := message.Text
	err := b.rep.DeleteRoom(room)
	if err != nil {
		slog.Error("Ошибка изменеия комнаты в БД")
		return fmt.Errorf("%s: %w", path, err)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Комната изменена")
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	slog.Info("Выполнена команда edit")
	return nil
}
