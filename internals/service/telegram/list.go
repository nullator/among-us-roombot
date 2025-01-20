package telegram

import (
	"among-us-roombot/internals/models"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Обработчик команды /list
func (b *Telegram) handleList(message *tgbotapi.Message) error {
	const path = "service.telegram.list"
	var rooms models.RoomList

	// Получаем список комнат из БД
	rooms, err := b.rep.GetRoomList()
	if err != nil {
		slog.Error("Ошибка получения списка комнат из БД")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Сортируем комнаты по времени создания
	sort.Sort(rooms)

	msgText := "*Румы, где ты можешь поиграть:*\n\n"

	if len(rooms) == 0 {
		last, err := b.rep.GetAndUpdateUserRequestTimestamp(message.Chat.ID)
		if err != nil {
			slog.Error("Ошибка получения времени последнего запроса пользователя")
			last = time.Now().Add(-24 * time.Hour)
		}
		// Если прошло менее 20 секунд
		if time.Since(last) < 20*time.Second {
			err = sendImage(b, message.Chat.ID, "sad.png")
			if err != nil {
				slog.Warn("Не удалось отправить картинку")
				return fmt.Errorf("%s: %w", path, err)
			}
			slog.Info("Отправлен хомяк")
		}

		msgText = "Пока нет ни одной румы 😔\nСоздай свою командой /add"
		msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
		msg.ParseMode = "MarkdownV2"
		msg.ReplyMarkup = list_kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("error send message to user")
			return fmt.Errorf("%s: %w", path, err)
		}
		slog.Info("Пользователю отправлен пустой список комнат",
			slog.String("user", message.Chat.UserName),
			slog.Int64("id", message.Chat.ID))

		return nil
	}

	i := 1
	indent := ""
	var emoji_map, emoji_mode string
	for _, room := range rooms {

		// Для каждой комнаты определяем эмодзи для карты и режима
		switch room.Map {
		case "Skeld":
			emoji_map = "🚀 "
		case "Polus":
			emoji_map = "⛄ "
		case "Airship":
			emoji_map = "🛩️ "
		case "Mira HQ":
			emoji_map = "🏢 "
		case "Fungle":
			emoji_map = "🍄 "
		default:
			emoji_map = "🚀 "
		}

		switch room.Mode {
		case "Классика":
			emoji_mode = "👨‍🎓 "
		case "Прятки":
			emoji_mode = "🧌 "
		case "Моды":
			emoji_mode = "🛠️ "
		default:
			emoji_mode = "🎲 "
		}

		indent = strings.Repeat(" ", 9)
		msgText += fmt.Sprintf("`%s`    ╭  %s %-10s\n", indent, emoji_map, room.Map)
		msgText += fmt.Sprintf("%d. `%-6s`       -   👑   *%-10s*\n", i, room.Code, room.Hoster)
		msgText += fmt.Sprintf("`%s`    ╰  %s %-10s\n\n", indent, emoji_mode, room.Mode)
		i++
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
	msg.ReplyMarkup = list_kb
	msg.ParseMode = "Markdown"

	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	slog.Info("Пользователю отправлен список комнат",
		slog.String("user", message.Chat.UserName),
		slog.Int64("id", message.Chat.ID))
	return nil
}
