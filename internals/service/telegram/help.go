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
		"Можно изменить код, карту, ник хостера или описание следуя указаниям бота\n\n"+

		"/subscribe - выполнить подписку на любого хоста, который сейчас хостит.\n"+
		"У хостов есть команда для расылки приглашений поиграть, эти сообщения "+
		"прийдут от бота всем кто подписан ха хоста\n\n"+

		"/unsubscribe - отписаться от рассылки\n\n"+

		"/notify - выполнить рассылку всем подписчикам. "+
		"Команду можно выполнять не чаще чем раз в 6 часов\n\n"+

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
