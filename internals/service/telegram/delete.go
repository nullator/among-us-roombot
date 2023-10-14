package telegram

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Telegram) handleDel(message *tgbotapi.Message) error {
	const path = "service.telegram.delete.handleDel"

	arg := message.CommandArguments()
	slog.Debug("Получены аргументы команды del: %s", arg)
	if arg != "" {
		match, _ := regexp.MatchString("^[a-zA-Z]{6}$", arg)
		if !match {
			return nil
		}
		code := strings.ToUpper(arg)

		admin_list := os.Getenv("ADMINS")
		admins := strings.Split(admin_list, ",")
		if b.isAdmin(admins, fmt.Sprint(message.Chat.ID)) {
			err := b.rep.DeleteRoom(code)
			if err != nil {
				slog.Error("Ошибка удаления комнаты из БД")
				return fmt.Errorf("%s: %w", path, err)
			}
			msg_text := fmt.Sprintf("Комната %s удалена", code)
			msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
			msg.ReplyMarkup = list_kb
			_, err = b.bot.Send(msg)
			if err != nil {
				slog.Error("error send message to user")
				return fmt.Errorf("%s: %w", path, err)
			}
			slog.Info("Комната удалена администратором",
				slog.String("code", code),
				slog.String("admin", fmt.Sprint(message.From.String())),
				slog.String("user", message.Chat.UserName),
				slog.Int64("id", message.Chat.ID))

			return nil
		} else {
			slog.Info("Попытка удаления комнаты не администратором",
				slog.String("code", code),
				slog.String("user", message.From.String()),
				slog.Int64("id", message.Chat.ID))
		}
	}

	exist_room, err := b.rep.GetUserStatus(message.Chat.ID, "room")
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о созданной пользователем комнате")
		return fmt.Errorf("%s: %w", path, err)
	}
	if exist_room == "" {
		msg_text := "У вас нет созданной комнаты.\n" +
			"Для создания комнаты введите команду /add"
		msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
		msg.ReplyMarkup = list_kb
		_, err := b.bot.Send(msg)
		if err != nil {
			slog.Error("error send message to user")
			return fmt.Errorf("%s: %w", path, err)
		}
		slog.Info("Пользователь у которого нет созданной комнаты пытается удалить комнату",
			slog.String("user", message.From.String()),
			slog.Int64("id", message.Chat.ID))
		return nil
	}

	caption := fmt.Sprintf("Удалить %s", exist_room)
	var kb = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(caption, "delete"),
			tgbotapi.NewInlineKeyboardButtonData("Отмена", "cancel"),
		),
	)

	msg_text := fmt.Sprintf("Вы действительно хотите удалить комнату %s?", exist_room)
	msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
	msg.ReplyMarkup = kb
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	slog.Info("Запущено удаление комнаты",
		slog.String("user", message.From.String()),
		slog.Int64("id", message.Chat.ID),
		slog.String("room", exist_room))
	return nil
}

func (b *Telegram) delete(message *tgbotapi.Message) error {
	const path = "service.telegram.delete.delete"

	// Удалить из базы данных
	exist_room, err := b.rep.GetUserStatus(message.Chat.ID, "room")
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о созданной пользователем комнате")
		return fmt.Errorf("%s: %w", path, err)
	}
	err = b.rep.DeleteRoom(exist_room)
	if err != nil {
		slog.Error("Ошибка удаления комнаты из БД")
		return fmt.Errorf("%s: %w", path, err)
	}
	slog.Info("Комната удалена из БД",
		slog.String("room", exist_room))

	// Обновить статус
	err = b.rep.SaveUserStatus(message.Chat.ID, "room", "")
	if err != nil {
		slog.Error("Ошибка сохранения в БД данных о созданной пользователем комнате")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Отправить сообщение пользователю и удалить кнопку удаления
	msg_text := fmt.Sprintf("Комната %s удалена", exist_room)
	msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
	msg.ReplyMarkup = list_kb
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}
	slog.Info("Комната удалена пользователем",
		slog.String("user", message.From.String()),
		slog.Int64("id", message.Chat.ID))

	return nil
}

func (b *Telegram) isAdmin(list []string, id string) bool {
	for _, v := range list {
		if v == id {
			return true
		}
	}
	return false
}
