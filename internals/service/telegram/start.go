package telegram

import (
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Telegram) handleStart(message *tgbotapi.Message) error {
	const path = "service.telegram.start"

	msg := tgbotapi.NewMessage(message.Chat.ID,
		"Привет! Я запасной бот для создания комнат в Among Us.\n"+
			"Пожалуйста, пользуйся только основным ботом @among_room_bot\n"+
			"Если он не работает, то ты можешь создать комнату со мной.\n")
	msg.ReplyMarkup = list_kb
	_, err := b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	slog.Info("Пользователь начал работу с ботом",
		slog.String("user", message.From.String()),
		slog.Int64("id", message.Chat.ID))
	return nil
}
