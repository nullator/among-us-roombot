package telegram

import (
	"among-us-roombot/internals/models"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Функция обработки команды /edit
func (b *Telegram) handleEdit(message *tgbotapi.Message) error {
	const path = "service.telegram.edit.handleEdit"

	// Проверка наличия у пользователя активной комнаты
	exist_room, err := b.rep.GetUserStatus(message.Chat.ID, "room")
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о созданной пользователем комнате")
		return fmt.Errorf("%s: %w", path, err)
	}
	if exist_room == "" {
		msg_text := "У тебя нет активной румы.\n" +
			"Для создания введи команду /add"
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

	// Если комната есть, то запускаем диалог изменения
	// Создается клавиатура с кнопками для выбора изменяемого параметра
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
	msg_text := "Что ты хочешь изменить в своей руме?"
	msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
	msg.ReplyMarkup = kb
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	slog.Info("Запущено редактирование комнаты",
		slog.String("user", message.From.String()),
		slog.Int64("id", message.Chat.ID),
		slog.String("room", exist_room))
	return nil
}

// Функция изменения кода комнаты, новый код в сообщении message.Text
func (b *Telegram) changeCode(message *tgbotapi.Message) error {
	const path = "service.telegram.edit.changeCode"

	code := message.Text
	// Проверка корректности нового кода комнаты
	match, _ := regexp.MatchString("^[a-zA-Z]{5}[fgqFGQ]$", code)
	if !match {
		slog.Info("Попытка изменить код комнаты на некорректный",
			slog.String("user", message.From.String()),
			slog.Int64("id", message.Chat.ID),
			slog.String("code", code))
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
			slog.Info("Попытка изменить код комнаты на уже существующий",
				slog.String("user", message.From.String()),
				slog.Int64("id", message.Chat.ID),
				slog.String("code", code))
			return models.ErrRoomAlreadyExist
		}
	}

	// Получить код существующей комнаты пользователя
	old_room_code, err := b.rep.GetUserStatus(message.Chat.ID, "room")
	if err != nil {
		slog.Error("Ошибка чтения из БД кода существующей комнаты")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Загрузить существующую комнату из базы данных
	var old_room *models.Room
	old_room, err = b.rep.GetRoom(old_room_code)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о существующей комнате")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Скорректировать код, а также сбросить время создания на текущее время
	// и сбросить флаг предупреждения о том что комната создана слишком давно
	old_room.Code = code
	old_room.Time = time.Now()
	old_room.Warning = false

	// Удалить старую комнату из базы данных
	// Скорее всего это избыточно, можно просто сохранить скорректированную комнату
	// TODO: проверить, что будет, если не удалять старую комнату
	err = b.rep.DeleteRoom(old_room_code)
	if err != nil {
		slog.Error("Ошибка удаления комнаты из БД")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Сохранить скорректированную комнату в базу данных
	err = b.rep.SaveRoom(old_room)
	if err != nil {
		slog.Error("Ошибка добавления комнаты в БД")
		return fmt.Errorf("%s: %w", path, err)
	}
	slog.Info("Комната изменена",
		slog.String("user", message.From.String()),
		slog.Int64("id", message.Chat.ID),
		slog.String("new_code", old_room_code))

	// Скорректированный код требуется сохранить в статус пользователя
	err = b.rep.SaveUserStatus(message.Chat.ID, "room", code)
	if err != nil {
		slog.Error("Ошибка сохранения в БД данных о новом коде комнаты")
		return fmt.Errorf("%s: %w", path, err)
	}

	return nil
}

// Функция изменения карты комнаты, новая карта в сообщении message.Text
func (b *Telegram) changeMap(message *tgbotapi.Message, mapa string) error {
	const path = "service.telegram.edit.changeMap"

	length := utf8.RuneCountInString(mapa)
	if length > 10 {
		slog.Info("Попытка изменить навазине карты на слишком длинное",
			slog.String("user", message.From.String()),
			slog.Int64("id", message.Chat.ID),
			slog.String("map", mapa))
		return models.ErrInvalidMap
	}

	// Загрузить код существующей комнаты пользователя
	old_room_code, err := b.rep.GetUserStatus(message.Chat.ID, "room")
	if err != nil {
		slog.Error("Ошибка чтения из БД кода существующей комнаты")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Загрузить существующую комнату из базы данных
	var old_room *models.Room
	old_room, err = b.rep.GetRoom(old_room_code)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о существующей комнате")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Скорректировать карту, а также сбросить время создания на текущее время
	// и сбросить флаг предупреждения о том что комната создана слишком давно
	old_room.Map = mapa
	old_room.Time = time.Now()
	old_room.Warning = false

	// Сохранить скорректированную комнату в базу данных
	err = b.rep.SaveRoom(old_room)
	if err != nil {
		slog.Error("Ошибка добавления комнаты в БД")
		return fmt.Errorf("%s: %w", path, err)
	}
	slog.Info("Название карты комнаты изменено",
		slog.String("user", message.From.String()),
		slog.Int64("id", message.Chat.ID),
		slog.String("code", old_room.Code),
		slog.String("new_map", old_room_code))

	return nil
}

// Функция изменения ника хоста комнаты, новый ник в сообщении message.Text
func (b *Telegram) changeHoster(message *tgbotapi.Message) error {
	const path = "service.telegram.edit.changeHoster"

	hoster := message.Text
	length := utf8.RuneCountInString(hoster)
	if length > 10 {
		slog.Info("Попытка изменить ник хоста на слишком длинный",
			slog.String("user", message.From.String()),
			slog.Int64("id", message.Chat.ID),
			slog.String("hostname", hoster))
		return models.ErrInvalidName
	}

	// Загрузить код старой комнаты пользователя
	old_room_code, err := b.rep.GetUserStatus(message.Chat.ID, "room")
	if err != nil {
		slog.Error("Ошибка чтения из БД кода существующей комнаты")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Загрузить существующую комнату из базы данных
	var old_room *models.Room
	old_room, err = b.rep.GetRoom(old_room_code)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о существующей комнате")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Скорректировать ник хоста, а также сбросить время создания на текущее время
	// и сбросить флаг предупреждения о том что комната создана слишком давно
	old_room.Hoster = hoster
	old_room.Time = time.Now()
	old_room.Warning = false

	// Сохранить скорректированный ник хоста в статус пользователя
	err = b.rep.SaveUserStatus(message.Chat.ID, "host_name", hoster)
	if err != nil {
		slog.Warn("Ошибка сохранения в БД данных о новом нике хоста")
	}

	// Сохранить скорректированную комнату в базу данных
	err = b.rep.SaveRoom(old_room)
	if err != nil {
		slog.Error("Ошибка добавления комнаты в БД")
		return fmt.Errorf("%s: %w", path, err)
	}
	slog.Info("Ник хоста комнаты изменен",
		slog.String("user", message.From.String()),
		slog.Int64("id", message.Chat.ID),
		slog.String("code", old_room.Code),
		slog.String("new_hostname", hoster))

	// Обновить модель хостера
	host, err := b.rep.GetHoster(message.Chat.ID)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о хостере")
		return fmt.Errorf("%s: %w", path, err)
	}
	// Если модель хостера не найдена, то создать новую
	if host == nil {
		host = &models.Hoster{
			ID:        message.Chat.ID,
			Name:      hoster,
			Followers: []models.User{},
			LastSend:  time.Now().Add(-12 * time.Hour),
		}
		slog.Warn("При обновлении румы не найдена модель хостера, создана новая",
			slog.String("user", message.Chat.UserName),
			slog.Int64("id", message.Chat.ID),
			slog.String("host", host.Name))
	} else {
		host.Name = hoster
	}
	err = b.rep.SaveHoster(host)
	if err != nil {
		slog.Error("Ошибка сохранения в БД данных о хостере")
		return fmt.Errorf("%s: %w", path, err)
	}
	slog.Info("В модели хостера успешно обновлен ник",
		slog.Int64("id", message.Chat.ID),
		slog.String("host", host.Name))

	return nil
}

// Функция изменения описания комнаты, новое описание в сообщении message.Text
func (b *Telegram) changeDescription(message *tgbotapi.Message, mode string) error {
	const path = "service.telegram.edit.changeDescription"

	length := utf8.RuneCountInString(mode)
	if length > 10 {
		slog.Info("Попытка изменить описание на слишком длинное",
			slog.String("user", message.From.String()),
			slog.Int64("id", message.Chat.ID),
			slog.String("description", mode))
		return models.ErrInvalidName
	}

	// Загрузить код существующей комнаты пользователя
	old_room_code, err := b.rep.GetUserStatus(message.Chat.ID, "room")
	if err != nil {
		slog.Error("Ошибка чтения из БД кода существующей комнаты")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Загрузить существующую комнату из базы данных
	var old_room *models.Room
	old_room, err = b.rep.GetRoom(old_room_code)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о существующей комнате")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Скорректировать описание, а также сбросить время создания на текущее время
	// и сбросить флаг предупреждения о том что комната создана слишком давно
	old_room.Mode = mode
	old_room.Time = time.Now()
	old_room.Warning = false

	// Сохранить скорректированную комнату в базу данных
	err = b.rep.SaveRoom(old_room)
	if err != nil {
		slog.Error("Ошибка добавления комнаты в БД")
		return fmt.Errorf("%s: %w", path, err)
	}
	slog.Info("Описание комнаты изменено",
		slog.String("user", message.From.String()),
		slog.Int64("id", message.Chat.ID),
		slog.String("code", old_room.Code),
		slog.String("new_description", mode))

	return nil
}
