package telegram

import (
	"among-us-roombot/internals/models"
	"fmt"
	"log/slog"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Метод бота, который проверяет комнаты на время их создания для целей удаления старых комнат
func (b *Telegram) Iterate() {
	for {
		err := b.checkRooms()
		if err != nil {
			slog.Error("Ошибка проверки комнат", slog.String("error", err.Error()))
		}
		time.Sleep(time.Second * 20)
	}
}

// Проверка комнат на время их создания
func (b *Telegram) checkRooms() error {
	const path = "service.telegram.iterator.checkRooms"
	var rooms models.RoomList

	// Получить список комнат из БД
	rooms, err := b.rep.GetRoomList()
	if err != nil {
		slog.Error("Ошибка получения списка комнат из БД")
		return fmt.Errorf("%s: %w", path, err)
	}

	for _, room := range rooms {
		// Проверка времени создания комнаты
		if time.Now().After(room.Time.Add(time.Minute * 240)) {

			// Если комната создана давно и предупреждение пользователю не отправлено
			// то ему отправляется предупреждение
			if !room.Warning {
				msgText := fmt.Sprintf("*Продлевать будете?*\n\n" +
					"Твоя рума создана более 4 часов назад\\. " +
					"Если код ещё актуален, прошу нажать кнопку \"Продлить\" " +
					"\\(или отправь мне сообщение с текстом \"Продлить\"\\)\\. " +
					"Если рума не актуальна, удали её нажав кнопку \"Удалить\" " +
					"\\(или командой /del\\)\\.")
				msg := tgbotapi.NewMessage(room.ID, msgText)
				var kb = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Продлить", "add_time"),
						tgbotapi.NewInlineKeyboardButtonData("Удалить", "delete"),
					),
				)
				msg.ReplyMarkup = kb
				msg.ParseMode = "MarkdownV2"
				_, err = b.bot.Send(msg)
				if err != nil {
					// Отдельно проверяется не успел ли хостер заблокировать бота
					if err.Error() == "Forbidden: bot was blocked by the user" {
						slog.Warn("Обнаружена блокировка бота")
						err := b.rep.DeleteRoom(room.Code)
						if err != nil {
							slog.Error("Ошибка удаления устаревшей комнаты из БД",
								slog.String("error", err.Error()))
							return fmt.Errorf("%s: %w", path, err)
						}
						err = b.rep.SaveUserStatus(room.ID, "room", "")
						if err != nil {
							slog.Error("Ошибка сохранения в БД данных о комнате",
								slog.String("error", err.Error()))
							return fmt.Errorf("%s: %w", path, err)
						}
						slog.Info("Устаревшая комната автоматически удалена (бот заблокирован)",
							slog.String("code", room.Code),
							slog.String("user", room.Hoster),
							slog.Int64("id", room.ID))
					} else {
						slog.Error("error send message to user")
						return fmt.Errorf("%s: %w", path, err)
					}
				}

				// Флаг ставиться чтобы не направлять пользователю предупреждение повторно
				room.Warning = true
				err = b.rep.SaveRoom(&room)
				slog.Info("Пользователю отправлено предупреждение",
					slog.String("user", room.Hoster),
					slog.Int64("id", room.ID),
					slog.String("room", room.Code))
				if err != nil {
					slog.Error("Ошибка сохранения в БД данных о предупреждении",
						slog.String("error", err.Error()))
				}
			}
		}

		// Удаление комнаты
		if time.Now().After(room.Time.Add(time.Minute * 270)) {
			slog.Debug("Комната устарела, удаляю",
				slog.String("room", room.Code))

			err := b.rep.DeleteRoom(room.Code)
			if err != nil {
				slog.Error("Ошибка удаления устаревшей комнаты из БД",
					slog.String("error", err.Error()))
				return fmt.Errorf("%s: %w", path, err)
			}
			err = b.rep.SaveUserStatus(room.ID, "room", "")
			if err != nil {
				slog.Error("Ошибка сохранения в БД данных о комнате",
					slog.String("error", err.Error()))
				return fmt.Errorf("%s: %w", path, err)
			}
			msgText := fmt.Sprintf("Комната %s автоматически удалена", room.Code)
			msg := tgbotapi.NewMessage(room.ID, msgText)
			msg.ReplyMarkup = list_kb
			_, err = b.bot.Send(msg)
			if err != nil {
				slog.Error("error send message to user")
				return fmt.Errorf("%s: %w", path, err)
			}
			slog.Info("Устаревшая комната автоматически удалена",
				slog.String("code", room.Code),
				slog.String("user", room.Hoster),
				slog.Int64("id", room.ID))
		}
	}

	return nil
}

// Функция продлевает время жизни комнаты обновляя время её создания и сбрасывая флаг предупреждения
func (b *Telegram) addTime(message *tgbotapi.Message) error {
	const path = "service.telegram.iterator.addTime"

	// Получить код комнаты
	exist_room, err := b.rep.GetUserStatus(message.Chat.ID, "room")
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о созданной пользователем комнате")
		return fmt.Errorf("%s: %w", path, err)
	}

	if exist_room == "" {
		slog.Warn("Пользователь пытается продлить несуществующую комнату",
			slog.String("user", message.From.String()),
			slog.Int64("id", message.Chat.ID))
		return nil
	}

	// Получить комнату из БД
	room, err := b.rep.GetRoom(exist_room)
	if err != nil {
		slog.Error("Ошибка получения комнаты из БД",
			slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", path, err)
	}

	// Добавить время
	room.Time = time.Now()
	room.Warning = false

	// Сохранить в БД
	err = b.rep.SaveRoom(room)
	if err != nil {
		slog.Error("Ошибка сохранения в БД данных о комнате",
			slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", path, err)
	}

	// Отправить сообщение пользователю
	msgText := fmt.Sprintf("Комната %s продлена.\n\n"+
		"Пожалуйста, не забудь её удалить когда закончишь играть.\n\n👍", room.Code)
	msg := tgbotapi.NewMessage(room.ID, msgText)
	msg.ReplyMarkup = list_kb
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}
	slog.Info("Комната продлена",
		slog.String("code", room.Code),
		slog.String("user", room.Hoster),
		slog.Int64("id", room.ID))

	return nil

}
