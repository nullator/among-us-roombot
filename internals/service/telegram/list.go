package telegram

import (
	"fmt"
	"log/slog"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Telegram) handleList(message *tgbotapi.Message) error {
	const path = "service.telegram.list"

	_, err := b.rep.GetRoomList()
	if err != nil {
		slog.Error("Ошибка получения списка комнат из БД")
		return fmt.Errorf("%s: %w", path, err)
	}

	rooms := map[string][]string{
		"AAAAAA": {"Skeld", "Хост 1", "Классика"},
		"BCVQQQ": {"Все", "младше 10", "прятки"},
		"NBVFFF": {"Polus", "старше 60", "душнилово"},
	}

	msgText := "*Румы, где ты можешь поиграть:*\n\n"
	i := 1
	indent := ""
	for code, room := range rooms {
		indent = strings.Repeat(" ", 9)
		msgText += fmt.Sprintf("`%s`    ╭  🚀  %-10s\n", indent, room[0])
		msgText += fmt.Sprintf("*%d\\. *`%-6s`       \\-   👑   *%-10s*\n", i, code, room[1])
		msgText += fmt.Sprintf("`%s`    ╰  🎲  %-10s\n\n", indent, room[2])
		i++
	}

	msgText += "\n"
	msgText += "||Если копировать, то полностью 😊||"

	msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
	msg.ParseMode = "MarkdownV2"

	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	slog.Info("Выполнена команда list")
	return nil
}
