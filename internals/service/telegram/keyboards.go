package telegram

import (
	"among-us-roombot/internals/models"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Функция создает клавиатуру с кнопками для подписки на хостеров. Проверяется четность количества хостеров для корректного формирования кнопок
// TODO: зачем передавать в функцию id_chat? Наверное раньше использовалось, но сейчас не актуально
func make_subscribe_kb(
	b *Telegram, id_chat int64,
	rooms []models.Room) tgbotapi.InlineKeyboardMarkup {

	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup()
	i := 0
	n := len(rooms) / 2
	for n > 0 {
		numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(rooms[i].Hoster, fmt.Sprintf("sub%d", rooms[i].ID)),
			tgbotapi.NewInlineKeyboardButtonData(rooms[i+1].Hoster, fmt.Sprintf("sub%d", rooms[i+1].ID)),
		),
		)
		i += 2
		n -= 1
	}

	if len(rooms)%2 == 1 {
		numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(rooms[len(rooms)-1].Hoster, fmt.Sprintf("sub%d", rooms[len(rooms)-1].ID)),
		),
		)
	}

	numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Отменить", "cancel"),
	),
	)

	return numericKeyboard
}

// Функция создает клавиатуру с кнопками для отписки от хостеров. Проверяется четность количества хостеров для корректного формирования кнопок
func make_unsubscribe_kb(b *Telegram, hosters []models.User) tgbotapi.InlineKeyboardMarkup {

	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup()
	i := 0
	n := len(hosters) / 2
	for n > 0 {
		numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(hosters[i].Name, fmt.Sprintf("ays%d", hosters[i].ID)),
			tgbotapi.NewInlineKeyboardButtonData(hosters[i+1].Name, fmt.Sprintf("ays%d", hosters[i+1].ID)),
		),
		)
		i += 2
		n -= 1
	}

	if len(hosters)%2 == 1 {
		numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(hosters[len(hosters)-1].Name, fmt.Sprintf("ays%d", hosters[len(hosters)-1].ID)),
		),
		)
	}

	numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Отменить", "cancel"),
	),
	)

	return numericKeyboard

}
