package telegram

import (
	"fmt"
	"strings"

	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Типовая клавиатура с одной кнопкой "Список рум" (команда /list)
var list_kb = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Список рум", "roomlist"),
	),
)

// Типовая клавиатура с одной кнопкой "Отменить" (команда /cancel)
var cancel_kb = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Отменить", "cancel"),
	),
)

// Обработка обновлений
func (b *Telegram) handleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		switch {
		// Обработка сообщений
		case update.Message != nil:
			// Получение имени пользователя для последующего логирования
			// TODO: переделать так, чтобы сохранялось что-то одно, имя или если оно
			// пустое то telegramID
			sender := fmt.Sprintf("%s (%s)",
				update.Message.From.UserName,
				update.Message.From.String())

			// Проверка не является ли сообщение командой
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

			// Отдельно проверяется сообщение "продлить" для простоты продления времени
			// жизни комнаты, потому что это сообщение может быть отправлено в любой момент
			// с любым статусом пользователя и не является командой
			// В том числе это требуется на случай, если хостер случайно удалит клавиатуру
			// с кнопкой продления времени, чтобы не пришлось создавать новую комнату
			if strings.ToLower(update.Message.Text) == "продлить" {
				err := b.addTime(update.Message)
				if err != nil {
					slog.Error("Ошибка продления времени",
						slog.String("error", err.Error()))
				}
			}

			// Проверка статуса пользователя и если он не пустой то обработка
			// сообщения соответствующей функцией в handleStatus.go
			user_status, err := b.rep.GetUserStatus(update.Message.Chat.ID, "status")
			if err != nil {
				slog.Error("Ошибка чтения из БД данных о статусе пользователя",
					slog.String("error", err.Error()))
			}
			if user_status != "null" {
				b.handleUserStatus(&update, user_status)
			}

		// Если нажата кнопка то она обрабатывается в handleButton.go функцией handleButton
		case update.CallbackQuery != nil:
			q := update.CallbackQuery.Data
			id := update.CallbackQuery.Message.Chat.ID
			b.handleButton(&update, q, id)

			// После нажатия на кнопку старая клавиатура удаляется
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
