package telegram

import (
	"among-us-roombot/internals/models"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Telegram) handleList(message *tgbotapi.Message) error {
	const path = "service.telegram.list"
	var rooms models.RoomList

	rooms, err := b.rep.GetRoomList()
	if err != nil {
		slog.Error("Ошибка получения списка комнат из БД")
		return fmt.Errorf("%s: %w", path, err)
	}

	sort.Sort(rooms)

	msgText := "*Румы, где ты можешь поиграть:*\n\n"

	if len(rooms) == 0 {
		msgText += "Пока нет ни одной комнаты 😔\nСоздай свою комнату с помощью команды /add"
		msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
		msg.ParseMode = "MarkdownV2"
		msg.ReplyMarkup = list_kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("error send message to user")
			return fmt.Errorf("%s: %w", path, err)
		}
		return nil
	}

	i := 1
	indent := ""
	for _, room := range rooms {
		indent = strings.Repeat(" ", 9)
		msgText += fmt.Sprintf("`%s`    ╭  🚀  %-10s\n", indent, room.Map)
		msgText += fmt.Sprintf("*%d\\. *`%-6s`       \\-   👑   *%-10s*\n", i, room.Code, room.Hoster)
		msgText += fmt.Sprintf("`%s`    ╰  🎲  %-10s\n\n", indent, room.Mode)
		i++
	}

	msgText += "\n"
	msgText += "||Если копировать, то полностью 😊||"

	msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
	msg.ReplyMarkup = list_kb
	msg.ParseMode = "MarkdownV2"

	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	slog.Info("Пользователю отправлен список комнат",
		slog.String("user", message.From.String()),
		slog.Int64("id", message.Chat.ID))
	return nil
}
