package telegram

import (
	"among-us-roombot/internals/models"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"unicode/utf8"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Telegram) handleEdit(message *tgbotapi.Message) error {
	const path = "service.telegram.edit.handleEdit"

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
		slog.Info("Попытка изменить комнату пользователем у котого нет румы",
			slog.String("user", message.From.String()),
			slog.Int64("id", message.Chat.ID))
		return nil
	}

	var kb = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Код", "change_code"),
			tgbotapi.NewInlineKeyboardButtonData("Карту", "change_map"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Ник хоста", "change_hoster"),
			tgbotapi.NewInlineKeyboardButtonData("Описание", "change_description"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Отмена", "cancel"),
		),
	)
	msg_text := "Что ты хочешь изменить в своей комнате?"
	msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
	msg.ReplyMarkup = kb
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	slog.Info("Запущено редактирование комнаты, жду подтверждение")
	return nil
}

func (b *Telegram) changeCode(message *tgbotapi.Message) error {
	const path = "service.telegram.edit.changeCode"

	code := message.Text
	// Проверка корректности нового кода комнаты
	match, _ := regexp.MatchString("^[a-zA-Z]{6}$", code)
	if !match {
		return models.ErrInvalidCode
	}
	code = strings.ToUpper(code)

	// Проверка на уникальность кода комнаты
	var rooms models.RoomList
	rooms, err := b.rep.GetRoomList()
	if err != nil {
		slog.Error("Ошибка получения списка комнат из БД")
		return fmt.Errorf("%s: %w", path, err)
	}
	for _, room := range rooms {
		if room.Code == code {
			return models.ErrRoomAlreadyExist
		}
	}

	// Загрузить старую комнату из базы данных
	old_room_code, err := b.rep.GetUserStatus(message.Chat.ID, "room")
	if err != nil {
		slog.Error("Ошибка чтения из БД кода существующей комнаты")
		return fmt.Errorf("%s: %w", path, err)
	}

	var old_room *models.Room
	old_room, err = b.rep.GetRoom(old_room_code)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о существующей комнате")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Скорректировать код
	old_room.Code = code

	// Удалить старую комнату из базы данных
	err = b.rep.DeleteRoom(old_room_code)
	if err != nil {
		slog.Error("Ошибка удаления комнаты из БД")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Сохранить скорректированную комнату в базу данных
	err = b.rep.AddRoom(old_room)
	if err != nil {
		slog.Error("Ошибка добавления комнаты в БД")
		return fmt.Errorf("%s: %w", path, err)
	}

	err = b.rep.SaveUserStatus(message.Chat.ID, "room", code)
	if err != nil {
		slog.Error("Ошибка сохранения в БД данных о новом коде комнаты")
		return fmt.Errorf("%s: %w", path, err)
	}

	return nil
}

func (b *Telegram) changeMap(message *tgbotapi.Message) error {
	const path = "service.telegram.edit.changeMap"

	mapa := message.Text
	length := utf8.RuneCountInString(mapa)
	if length > 10 {
		return models.ErrInvalidMap
	}

	// Загрузить старую комнату из базы данных
	old_room_code, err := b.rep.GetUserStatus(message.Chat.ID, "room")
	if err != nil {
		slog.Error("Ошибка чтения из БД кода существующей комнаты")
		return fmt.Errorf("%s: %w", path, err)
	}

	var old_room *models.Room
	old_room, err = b.rep.GetRoom(old_room_code)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о существующей комнате")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Скорректировать код
	old_room.Map = mapa

	// Удалить старую комнату из базы данных
	err = b.rep.DeleteRoom(old_room_code)
	if err != nil {
		slog.Error("Ошибка удаления комнаты из БД")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Сохранить скорректированную комнату в базу данных
	err = b.rep.AddRoom(old_room)
	if err != nil {
		slog.Error("Ошибка добавления комнаты в БД")
		return fmt.Errorf("%s: %w", path, err)
	}

	return nil
}

func (b *Telegram) changeHoster(message *tgbotapi.Message) error {
	const path = "service.telegram.edit.changeHoster"

	hoster := message.Text
	length := utf8.RuneCountInString(hoster)
	if length > 10 {
		return models.ErrInvalidName
	}

	// Загрузить старую комнату из базы данных
	old_room_code, err := b.rep.GetUserStatus(message.Chat.ID, "room")
	if err != nil {
		slog.Error("Ошибка чтения из БД кода существующей комнаты")
		return fmt.Errorf("%s: %w", path, err)
	}

	var old_room *models.Room
	old_room, err = b.rep.GetRoom(old_room_code)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о существующей комнате")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Скорректировать код
	old_room.Hoster = hoster

	// Удалить старую комнату из базы данных
	err = b.rep.DeleteRoom(old_room_code)
	if err != nil {
		slog.Error("Ошибка удаления комнаты из БД")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Сохранить скорректированную комнату в базу данных
	err = b.rep.AddRoom(old_room)
	if err != nil {
		slog.Error("Ошибка добавления комнаты в БД")
		return fmt.Errorf("%s: %w", path, err)
	}

	return nil
}

func (b *Telegram) changeDescription(message *tgbotapi.Message) error {
	const path = "service.telegram.edit.changeDescription"

	mode := message.Text
	length := utf8.RuneCountInString(mode)
	if length > 10 {
		return models.ErrInvalidName
	}

	// Загрузить старую комнату из базы данных
	old_room_code, err := b.rep.GetUserStatus(message.Chat.ID, "room")
	if err != nil {
		slog.Error("Ошибка чтения из БД кода существующей комнаты")
		return fmt.Errorf("%s: %w", path, err)
	}

	var old_room *models.Room
	old_room, err = b.rep.GetRoom(old_room_code)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о существующей комнате")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Скорректировать код
	old_room.Mode = mode

	// Удалить старую комнату из базы данных
	err = b.rep.DeleteRoom(old_room_code)
	if err != nil {
		slog.Error("Ошибка удаления комнаты из БД")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Сохранить скорректированную комнату в базу данных
	err = b.rep.AddRoom(old_room)
	if err != nil {
		slog.Error("Ошибка добавления комнаты в БД")
		return fmt.Errorf("%s: %w", path, err)
	}

	return nil
}
