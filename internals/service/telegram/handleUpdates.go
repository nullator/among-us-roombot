package telegram

import (
	"fmt"
	"os"
	"strconv"

	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
					slog.String("sender", sender),
					slog.String("cmd", update.Message.Text))

				if err := b.handleCommand(update.Message); err != nil {
					slog.Error("При обработке команды произошла ошибка",
						slog.String("cmd", update.Message.Command()),
						slog.String("error", err.Error()),
					)
				}
				continue
			}

			slog.Info("Получено сообщение",
				slog.String("sender", sender),
				slog.String("message", update.Message.Text))

			// Checking to receive feedback
			user_status, err := b.rep.GetUserStatus(update.Message.Chat.ID)
			if err != nil {
				slog.Error("Ошибка чтения из БД данных о статусе пользователя",
					slog.String("error", err.Error()))
			}
			if user_status == "wait_feedback" {
				admin_id, err := strconv.ParseInt(os.Getenv("TG_adminID"), 10, 64)
				if err != nil {
					slog.Error("при выполнении авторизации не удалось распарсить ID в TelegramId",
						slog.String("error", err.Error()))
				}

				msg_text := fmt.Sprintf("Получена обратная связь от %s содержания: %s",
					update.Message.From.String(), update.Message.Text)
				msg := tgbotapi.NewMessage(admin_id, msg_text)
				_, err = b.bot.Send(msg)
				if err != nil {
					slog.Error("Не удалось отправить обратную связь",
						slog.String("error", err.Error()))
				}

				forvard_msg := tgbotapi.NewForward(admin_id,
					update.Message.Chat.ID,
					update.Message.MessageID)
				_, err = b.bot.Send(forvard_msg)
				if err != nil {
					slog.Error("Не удалось переслать сообщение",
						slog.String("error", err.Error()))
				}

				msg_text = "Спасибо, сообщение отправлено разработчику! " +
					"При необходимости можно повторно ввести команду /feedback " +
					"и отправить ещё одно сообщение, в том числе можно отправить файлы, скриншоты и т.п."
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, msg_text)
				_, err = b.bot.Send(msg)
				if err != nil {
					slog.Error("Не удалось отправить обратную связь",
						slog.String("error", err.Error()))
				}

				err = b.rep.SaveUserStatus(update.Message.Chat.ID, "null")
				if err != nil {
					slog.Error("Ошибка сохранения в БД данных о статусе пользователя",
						slog.String("error", err.Error()))

				}
				slog.Debug("Успешно изменён статус в БД")
				// break - нужен если будет код для обработки прочих сообщений
			}

			// Proceed messages if needed

		}
	}
}
