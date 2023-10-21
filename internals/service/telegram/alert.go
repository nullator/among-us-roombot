package telegram

import (
	"among-us-roombot/internals/models"
	"fmt"
	"log/slog"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Выполняется рассылка сообщения подписчикам. При нажатии на кнопку запрашивается
// текст сообщения, после этого выполняется запрос к БД, получается список подписчиков,
// которые подписаны на хостера. Выполняется рассылка в цикле через горутину.
// После этого в модели хостера обновляется поле с временем последней рассылки.
// После этого модель хостера сохраняется в БД. Хостеру выводится сообщение об
// успешной рассылке с указанием количества подписчиков.
func (b *Telegram) handleAlert(message *tgbotapi.Message) error {
	const path = "service.telegram.alert.handleAlert"

	var host *models.Hoster
	host, err := b.rep.GetHoster(message.Chat.ID)
	if err != nil {
		slog.Error("Ошибка чтения из БД модели хостера")
		return fmt.Errorf("%s: %w", path, err)
	}

	if host == nil {
		msg_text := "Не удлось выполнить команду рассылки.\n\n" +
			"Скорее всего ты никогда не хостил руму, поэтому у тебя не может " +
			"быть подписчиков"
		msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
		msg.ReplyMarkup = list_kb
		_, err := b.bot.Send(msg)
		if err != nil {
			slog.Error("error send message to user")
			return fmt.Errorf("%s: %w", path, err)
		}
		slog.Info("Попытка отправить рассылку пользователем у которого нет модели хостера",
			slog.String("user", message.From.String()),
			slog.Int64("id", message.Chat.ID))
		return nil
	}

	delta := time.Now().Unix() - host.LastSend.Unix()
	if delta < (60 * 60 * 6) {
		t := time.Unix(delta, 0)
		t_str := fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
		msg_text := fmt.Sprintf("Рассылку можно отправлять не чаще чем раз в 6 часов, "+
			"следующая рассылка возможна через %s", t_str)
		msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
		msg.ReplyMarkup = list_kb
		_, err := b.bot.Send(msg)
		if err != nil {
			slog.Error("error send message to user")
			return fmt.Errorf("%s: %w", path, err)
		}
		slog.Info("Попытка отправить рассылку пользователем, который уже отправлял рассылку",
			slog.String("user", message.From.String()),
			slog.Int64("id", message.Chat.ID))

		return nil
	}

	room_code, err := b.rep.GetUserStatus(message.Chat.ID, "room")
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о созданной пользователем комнате")
		return fmt.Errorf("%s: %w", path, err)
	}
	if room_code == "" {
		msg_text := "Введи текст рассылки, который будет направлен твоим подписчикам.\n\n" +
			"Рассылку можно направлять не чаще чам раз в 6 часов, постарайся ей " +
			"не злоупотреблять, чтобы твои подписчики от тебя не отписались"
		msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
		msg.ReplyMarkup = cancel_kb
		_, err := b.bot.Send(msg)
		if err != nil {
			slog.Error("error send message to user")
			return fmt.Errorf("%s: %w", path, err)
		}
		b.rep.SaveUserStatus(message.Chat.ID, "status", "wait_post")
		slog.Info("Хостер без румы запустил команду рассылки, жду текст",
			slog.String("user", message.From.String()),
			slog.Int64("id", message.Chat.ID))
		return nil

	} else {
		room, err := b.rep.GetRoom(room_code)
		if err != nil {
			slog.Error("Ошибка чтения из БД данных о созданной пользователем комнате")
			return fmt.Errorf("%s: %w", path, err)
		}
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Отправить шаблон", "send_template"),
				tgbotapi.NewInlineKeyboardButtonData("Отменить рассылку", "cancel"),
			),
		)
		draft_post := fmt.Sprintf("Привет\\!\n\nЗаходи ко мне поиграть\\, "+
			"я играю на карте %s\\, режим %s", room.Map, room.Mode)
		msg_text := "Пришли мне текст рассылки\\, который будет направлен твоим подписчикам\\, " +
			"или нажми на кнопку \"Отправить шаблон\"\\, чтобы отправить " +
			"следующее типовое сообщение \\(рассылку можно делать " +
			"не чаще чем раз в 6 часов\\)\\:\n\n" + draft_post
		msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
		msg.ParseMode = "MarkdownV2"
		msg.ReplyMarkup = kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("error send message to user")
			return fmt.Errorf("%s: %w", path, err)
		}
		b.rep.SaveUserStatus(message.Chat.ID, "status", "wait_post")
		slog.Info("Хостер с активной румой запустил команду рассылки, жду текст",
			slog.String("hoster", host.Name),
			slog.Int64("id", message.Chat.ID))
		return nil
	}
}
