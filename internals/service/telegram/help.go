package telegram

import (
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Telegram) handleHelp(message *tgbotapi.Message) error {
	const path = "service.telegram.help"

	msg := tgbotapi.NewMessage(message.Chat.ID, "Список команд:\n"+
		"/add - добавить комнату\n"+
		"Возможно ввести паматеры в формате \"/add ABCDEF никнейм карта описание\"\n"+
		"или ввести команду без параметров и указать данные пошагово\n\n"+

		"/edit - редактировать комнату\n"+
		"Изменить можно код, карту, ник хостера или описание следуюя указания бота\n\n"+

		"/del - удалить комнату\n\n"+

		"/feedback - обратная связь с разработчиком\n"+
		"Можно отправить сообщение разработчику, к сообщению можно приложить файлы")
	msg.ReplyMarkup = list_kb
	_, err := b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	slog.Info("Пользователю отправлена помощь по командам",
		slog.String("user", message.From.String()),
		slog.Int64("id", message.Chat.ID))
	return nil
}
