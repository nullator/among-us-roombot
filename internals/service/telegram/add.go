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

func (b *Telegram) handleAdd(message *tgbotapi.Message) error {
	const path = "service.telegram.add"

	exist_room, err := b.rep.GetUserStatus(message.Chat.ID, "room")
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о созданной пользователем комнате")
		return fmt.Errorf("%s: %w", path, err)
	}
	if exist_room != "" {
		msg_text := fmt.Sprintf("Вы уже создали комнату %s.\n"+
			"Для удаления существующей комнаты введите команду /del\n"+
			"Для редактирования существующей комнаты введите команду /edit", exist_room)
		msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
		_, err := b.bot.Send(msg)
		if err != nil {
			slog.Error("error send message to user")
			return fmt.Errorf("%s: %w", path, err)
		}
		return nil
	}

	var room *models.Room
	arg := message.CommandArguments()
	slog.Debug("Получены аргументы команды add: %s", arg)

	if arg == "" { // Если аргументы не переданы, запускаем пошаговый цикл создания комнаты

		// TODO Пошаговый цикл создания комнаты

	} else { // Если аргументы переданы, то разбиваем их на отдельные значения
		values := strings.Split(arg, " ")
		room, err = b.validateValues(values)
		if err != nil {
			switch err {
			case models.ErrInvalidNumberArgument:
				msg_text := "Неверное количество аргументов.\n" +
					"Комнату можно создать 2 способами:\n" +
					"1. Пошагово, введя команду /add и следуя инструкциям бота\n" +
					"2. Введя команду /add и через пробел параметры, например:\n" +
					"\"/add ABCDEF никнейм карта описание\""
				msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
					return fmt.Errorf("%s: %w", path, err)
				}
				return nil

			case models.ErrInvalidCode:
				msg_text := "Неверный код комнаты.\n" +
					"Код комнаты должен состоять из 6 латинских букв"
				msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
					return fmt.Errorf("%s: %w", path, err)
				}
				return nil

			case models.ErrRoomAlreadyExist:
				msg_text := "Комната с таким кодом уже существует"
				msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
					return fmt.Errorf("%s: %w", path, err)
				}
				return nil

			case models.ErrInvalidName:
				msg_text := "Слишком длинный никнейм.\n" +
					"Никнейм должен состоять не более чем из 10 символов"
				msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
					return fmt.Errorf("%s: %w", path, err)
				}
				return nil

			case models.ErrInvalidMap:
				msg_text := "Слишком длинное название карты.\n" +
					"Название карты должно состоять не более чем из 10 символов"
				msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
					return fmt.Errorf("%s: %w", path, err)
				}
				return nil

			case models.ErrInvalidMode:
				msg_text := "Слишком длинное описание режима игры.\n" +
					"Описание режима игры должно состоять не более чем из 10 символов"
				msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
					return fmt.Errorf("%s: %w", path, err)
				}
				return nil

			default:
				msg_text := "При выполнении команды произошла неожиданная ошибка"
				msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
					return fmt.Errorf("%s: %w", path, err)
				}
				return nil
			}

		}

		slog.Debug("Получены валидные аргументы команды add: %s", room)
	}

	if room == nil {
		slog.Error("Сформирована пустая модель комнаты")
		return fmt.Errorf("%s: %w", path, fmt.Errorf("room is nil"))
	}

	err = b.rep.AddRoom(room)
	if err != nil {
		slog.Error("Ошибка добавления комнаты в БД")
		return fmt.Errorf("%s: %w", path, err)
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, "Комната добавлена")
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	err = b.rep.SaveUserStatus(message.Chat.ID, "room", room.Code)
	if err != nil {
		slog.Error("Ошибка сохранения в БД данных о созданной пользователем комнате")
		return fmt.Errorf("%s: %w", path, err)
	}

	slog.Info("Выполнена команда add")
	return nil
}

func (b *Telegram) validateValues(values []string) (*models.Room, error) {
	path := "service.telegram.validateValues"

	// Проверка на количество аргументов
	if len(values) != 4 {
		return nil, models.ErrInvalidNumberArgument
	}

	// Проверка на валидность кода комнаты
	code := values[0]
	match, _ := regexp.MatchString("^[a-zA-Z]{6}$", code)
	if !match {
		return nil, models.ErrInvalidCode
	}
	code = strings.ToUpper(code)

	// Проверка на уникальность кода комнаты
	var rooms models.RoomList
	rooms, err := b.rep.GetRoomList()
	if err != nil {
		slog.Error("Ошибка получения списка комнат из БД")
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	for _, room := range rooms {
		if room.Code == code {
			return nil, models.ErrRoomAlreadyExist
		}
	}

	// Проверка на длину ника
	name := values[1]
	length := utf8.RuneCountInString(name)
	if length > 10 {
		return nil, models.ErrInvalidName
	}

	// Проверка на длину названия карты
	mapa := values[2]
	length = utf8.RuneCountInString(mapa)
	if length > 10 {
		return nil, models.ErrInvalidMap
	}

	// Проверка на длину описания режима игры
	mode := values[3]
	length = utf8.RuneCountInString(mode)
	if length > 10 {
		return nil, models.ErrInvalidMode
	}

	// Формирование и возврат модели комнаты
	room := models.Room{
		Code:       code,
		Mode:       mode,
		Hoster:     name,
		Map:        mapa,
		Descrition: "",
		Time:       time.Now(),
	}

	return &room, nil
}
