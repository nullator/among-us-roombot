package telegram

import (
	"among-us-roombot/internals/models"
	"fmt"

	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Выполняется запрос к БД, получается список хостеров на которых подписан пользователь,
// генерируются кнопки с никами. При нажатии на кнопку - отписка от хостера:
// из БД загружается модель юзера в которой есть поле Hosters []User.
// Из этого поля удаляется хостер. После этого модель юзера сохраняется в БД.
// Из БД загружается модель хостера, в которой есть поле Followers []User.
// Из этого поля удаляется подписчик. После этого модель хостера сохраняется в БД.
// Выводится сообщение об отписке.
func (b *Telegram) handleUnsubscribe(message *tgbotapi.Message) error {
	const path = "service.telegram.unsubscribe.handleUnsubscribe"

	var user *models.Follower
	user, err := b.rep.GetUser(message.Chat.ID)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о пользователе",
			slog.String("user", message.From.String()),
			slog.Int64("id", message.Chat.ID),
			slog.String("path", path))
	}

	if len(user.Hosters) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Ты ни на кого не подписан")
		msg.ReplyMarkup = list_kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("error send message to user")
			return fmt.Errorf("%s: %w", path, err)
		}
		return nil
	}

	kb := make_unsubscribe_kb(b, user.Hosters)
	msg := tgbotapi.NewMessage(message.Chat.ID, "Нажми на кнопку "+
		"с ником хостера, от которого хочешь отписаться:")
	msg.ReplyMarkup = kb
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	return nil
}
