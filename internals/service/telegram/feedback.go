package telegram

import (
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Telegram) handleFeedback(message *tgbotapi.Message) error {
	const path = "service.telegram.feedback"

	err := b.rep.SaveUserStatus(message.Chat.ID, "status", "wait_feedback")
	if err != nil {
		slog.Error("Ошибка сохранения в БД данных о статусе пользователя",
			slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", path, err)
	}
	msg_text := fmt.Sprintf("Введи сообщение, которое будет доставлено разработчику бота.\n" +
		"Если хочешь получить обратную связь, не забудь указать как с тобой связаться.\n" +
		"К сообщению можно приложить файлы, скриншоты, видео и т.п.:\n")
	msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
	msg.ReplyMarkup = cancel_kb
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user",
			slog.String("msg_text", msg_text),
			slog.String("error", err.Error()),
		)
	}
	slog.Info("Жду обратную связь от пользователя",
		slog.String("user", message.From.String()),
		slog.Int64("id", message.Chat.ID))
	return nil
}
