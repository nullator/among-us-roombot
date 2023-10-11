package telegram

import (
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Telegram) handleDel(message *tgbotapi.Message) error {
	const path = "service.telegram.delete.handleDel"

	exist_room, err := b.rep.GetUserStatus(message.Chat.ID, "room")
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о созданной пользователем комнате")
		return fmt.Errorf("%s: %w", path, err)
	}
	if exist_room == "" {
		msg_text := "У вас нет созданной комнаты.\n" +
			"Для создания комнаты введите команду /add"
		msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
		_, err := b.bot.Send(msg)
		if err != nil {
			slog.Error("error send message to user")
			return fmt.Errorf("%s: %w", path, err)
		}
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

	slog.Info("Запущено удаление комнаты, жду подтверждение")
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

	// Обновить статус
	err = b.rep.SaveUserStatus(message.Chat.ID, "room", "")
	if err != nil {
		slog.Error("Ошибка сохранения в БД данных о созданной пользователем комнате")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Отправить сообщение пользователю и удалить кнопку удаления

	msg_text := fmt.Sprintf("Комната %s удалена", exist_room)
	msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	return nil

}
