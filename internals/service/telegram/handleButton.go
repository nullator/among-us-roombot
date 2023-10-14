package telegram

import (
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Telegram) handleButton(update *tgbotapi.Update, button string, id int64) {
	switch button {
	case "delete":
		err := b.delete(update.CallbackQuery.Message)
		if err != nil {
			slog.Error("Ошибка удаления комнаты",
				slog.String("error", err.Error()))
		}

	case "change_code":
		slog.Info("Зафиксировано нажатие на кнопку изменения кода комнаты",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))
		err := b.rep.SaveUserStatus(id, "status", "edit_code")
		if err != nil {
			slog.Error("Ошибка сохранения в БД данных о статусе пользователя",
				slog.String("error", err.Error()))
			msg := tgbotapi.NewMessage(id, "Произошла ошибка при выполнении команды")
			msg.ReplyMarkup = list_kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}
		}
		msg := tgbotapi.NewMessage(id, "Отправь мне новый код комнаты:")
		msg.ReplyMarkup = cancel_kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("Ошибка отправки сообщения",
				slog.String("error", err.Error()))
		}

	case "change_map":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия карты",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))
		err := b.rep.SaveUserStatus(id, "status", "change_map")
		if err != nil {
			slog.Error("Ошибка сохранения в БД данных о статусе пользователя",
				slog.String("error", err.Error()))
			msg := tgbotapi.NewMessage(id, "Произошла ошибка при выполнении команды")
			msg.ReplyMarkup = list_kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}
		}
		msg := tgbotapi.NewMessage(id, "Отправь мне новое название карты:")
		msg.ReplyMarkup = cancel_kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("Ошибка отправки сообщения",
				slog.String("error", err.Error()))
		}

	case "change_hoster":
		slog.Info("Зафиксировано нажатие на кнопку изменения ника хостера",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))
		err := b.rep.SaveUserStatus(id, "status", "change_hoster")
		if err != nil {
			slog.Error("Ошибка сохранения в БД данных о статусе пользователя",
				slog.String("error", err.Error()))
			msg := tgbotapi.NewMessage(id, "Произошла ошибка при выполнении команды")
			msg.ReplyMarkup = list_kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}
		}
		msg := tgbotapi.NewMessage(id, "Отправь мне новый ник хостера:")
		msg.ReplyMarkup = cancel_kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("Ошибка отправки сообщения",
				slog.String("error", err.Error()))
		}

	case "change_description":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия режима",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))
		err := b.rep.SaveUserStatus(id, "status", "change_description")
		if err != nil {
			slog.Error("Ошибка сохранения в БД данных о статусе пользователя",
				slog.String("error", err.Error()))
			msg := tgbotapi.NewMessage(id, "Произошла ошибка при выполнении команды")
			msg.ReplyMarkup = list_kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}
		}
		msg := tgbotapi.NewMessage(id, "Отправь мне новое название режима:")
		msg.ReplyMarkup = cancel_kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("Ошибка отправки сообщения",
				slog.String("error", err.Error()))
		}

	case "cancel":
		slog.Info("Зафиксировано нажатие на кнопку отмены команды",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))
		err := b.rep.SaveUserStatus(id, "status", "null")
		if err != nil {
			slog.Error("Ошибка сохранения в БД данных о статусе пользователя",
				slog.String("error", err.Error()))
		}
		msg := tgbotapi.NewMessage(id, "Выполнение команды отменено")
		msg.ReplyMarkup = list_kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("Ошибка отправки сообщения",
				slog.String("error", err.Error()))
		}

	case "roomlist":
		slog.Info("Зафиксировано нажатие на кнопку вывода списка комнат",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))
		err := b.handleList(update.CallbackQuery.Message)
		if err != nil {
			slog.Error("Ошибка вывода списка комнат",
				slog.String("error", err.Error()))
		}
	}
}
