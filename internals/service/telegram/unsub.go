package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// Выполняется запрос к БД, получается список хостеров на которых подписан пользователь,
// генерируются кнопки с никами. При нажатии на кнопку - отписка от хостера:
// из БД загружается модель юзера в которой есть поле Hosters []User.
// Из этого поля удаляется хостер. После этого модель юзера сохраняется в БД.
// Из БД загружается модель хостера, в которой есть поле Followers []User.
// Из этого поля удаляется подписчик. После этого модель хостера сохраняется в БД.
// Выводится сообщение об отписке.
func (b *Telegram) handleUnsubscribe(message *tgbotapi.Message) error {
	const path = "service.telegram.unsubscribe.handleUnsubscribe"

	return nil
}
