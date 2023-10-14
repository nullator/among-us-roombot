package telegram

import (
	"fmt"
	"os"
	"strconv"

	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var list_kb = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Список рум", "roomlist"),
	),
)

var cancel_kb = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Отменить", "cancel"),
	),
)

func (b *Telegram) handleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		switch {
		// Processing chat messages
		case update.Message != nil:
			sender := fmt.Sprintf("%s (%s)",
				update.Message.From.UserName,
				update.Message.From.String())

			// Checking for command input
			if update.Message.IsCommand() {
				slog.Info("Зафиксирована команда",
					slog.String("cmd", update.Message.Text),
					slog.String("user", sender),
					slog.Int64("id", update.Message.Chat.ID),
				)

				if err := b.handleCommand(update.Message); err != nil {
					slog.Error("При обработке команды произошла ошибка",
						slog.String("cmd", update.Message.Command()),
						slog.String("error", err.Error()),
						slog.String("user", sender),
						slog.Int64("id", update.Message.Chat.ID),
					)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID,
						"Произошла неожиданная ошибка при обработке команды")
					_, err := b.bot.Send(msg)
					if err != nil {
						slog.Error("error send message to user")
					}

				}
				continue
			}

			slog.Info("Получено сообщение",
				slog.String("user", sender),
				slog.String("message", update.Message.Text),
				slog.Int64("id", update.Message.Chat.ID))

			// Checking status
			user_status, err := b.rep.GetUserStatus(update.Message.Chat.ID, "status")
			if err != nil {
				slog.Error("Ошибка чтения из БД данных о статусе пользователя",
					slog.String("error", err.Error()))
			}
			if user_status != "null" {
				b.handleUserStatus(&update, user_status)
			}

		// Обработка нажатий на кнопки
		case update.CallbackQuery != nil:
			q := update.CallbackQuery.Data
			id := update.CallbackQuery.Message.Chat.ID
			switch q {
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

			msg := tgbotapi.NewEditMessageReplyMarkup(
				id, update.CallbackQuery.Message.MessageID,
				tgbotapi.InlineKeyboardMarkup{
					InlineKeyboard: make([][]tgbotapi.InlineKeyboardButton, 0)})
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка удаления клавиатуры",
					slog.String("error", err.Error()))
			}

		}
	}
}

func (b *Telegram) feedback(update *tgbotapi.Update) error {
	admin_id, err := strconv.ParseInt(os.Getenv("TG_adminID"), 10, 64)
	if err != nil {
		slog.Error("при выполнении авторизации не удалось распарсить ID в TelegramId",
			slog.String("error", err.Error()))
		return err
	}

	msg_text := fmt.Sprintf("Получена обратная связь от %s содержания: %s",
		update.Message.From.String(), update.Message.Text)
	msg := tgbotapi.NewMessage(admin_id, msg_text)
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("Не удалось отправить обратную связь",
			slog.String("error", err.Error()))
		return err
	}

	forvard_msg := tgbotapi.NewForward(admin_id,
		update.Message.Chat.ID,
		update.Message.MessageID)
	_, err = b.bot.Send(forvard_msg)
	if err != nil {
		slog.Error("Не удалось переслать сообщение",
			slog.String("error", err.Error()))
		return err
	}

	msg_text = "Спасибо, сообщение отправлено разработчику! " +
		"При необходимости можно повторно ввести команду /feedback " +
		"и отправить ещё одно сообщение, в том числе можно отправить файлы, скриншоты и т.п."
	msg = tgbotapi.NewMessage(update.Message.Chat.ID, msg_text)
	msg.ReplyMarkup = list_kb
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("Не удалось отправить обратную связь",
			slog.String("error", err.Error()))
		return err
	}

	err = b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")
	if err != nil {
		slog.Error("Ошибка сохранения в БД данных о статусе пользователя",
			slog.String("error", err.Error()))
		return err
	}
	slog.Debug("Успешно изменён статус в БД")

	return nil
}
