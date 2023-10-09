package telegram

import (
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Telegram) handleRules(message *tgbotapi.Message) error {
	const path = "service.telegram.rules"

	msg := tgbotapi.NewMessage(message.Chat.ID, "Правила:\n"+
		"1. Не материться\n"+
		"2. Не оскорблять других игроков\n"+
		"3. Не использовать читы\n"+
		"4. Не использовать баги\n"+
		"5. Не использовать никнеймы, которые могут оскорбить других игроков\n")
	_, err := b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	slog.Info("Выполнена команда rules")
	return nil
}
