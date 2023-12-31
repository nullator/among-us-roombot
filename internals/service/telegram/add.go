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

// handleAdd обрабатывает команду /add
func (b *Telegram) handleAdd(message *tgbotapi.Message) error {
	const path = "service.telegram.add"

	// Проверка наличия существующей комнаты у пользователя
	exist_room, err := b.rep.GetUserStatus(message.Chat.ID, "room")
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о созданной пользователем комнате")
		return fmt.Errorf("%s: %w", path, err)
	}
	if exist_room != "" {
		msg_text := fmt.Sprintf("У тебя уже есть рума, зачем тебе вторая?.\n\n" +
			"Для удаления существующей введи команду /del\n" +
			"Для редактирования введи команду /edit")
		msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
		msg.ReplyMarkup = list_kb
		_, err := b.bot.Send(msg)
		if err != nil {
			slog.Error("error send message to user")
			return fmt.Errorf("%s: %w", path, err)
		}
		slog.Info("Пользователь попытался создать вторую комнату",
			slog.String("user", message.From.String()),
			slog.Int64("id", message.Chat.ID))

		return nil
	}

	// Если существующей комнаты нет, то создается её модель
	// arg - аргументы команды /add
	var room *models.Room
	arg := message.CommandArguments()

	if arg == "" { // Если аргументы не переданы, запускаем пошаговый цикл создания комнаты
		// Изменить статус пользователя на "start_add_room"
		// Пользователю отправляется сообщение с просьбой ввести код комнаты
		// В дальнейшем сообщение с кодом будет обработано в handleUpdate
		// где будет зафиксирован и обработан в handleStatus статус "start_add_room"
		slog.Info("Пользователь начал пошаговое создание комнаты",
			slog.String("user", message.From.String()),
			slog.Int64("id", message.Chat.ID))
		err = b.rep.SaveUserStatus(message.Chat.ID, "status", "start_add_room")
		if err != nil {
			slog.Error("Ошибка сохранения в БД статуса о старте создания комнаты")
			return fmt.Errorf("%s: %w", path, err)
		}
		msg := tgbotapi.NewMessage(message.Chat.ID, "Введи код комнаты:")
		msg.ReplyMarkup = cancel_kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("error send message to user")
			return fmt.Errorf("%s: %w", path, err)
		}
		return nil

	} else { // Если аргументы переданы, то разбиваем их на отдельные значения
		values := strings.Split(arg, " ")
		// Проверка на валидность аргументов
		room, err = b.validateValues(values)
		slog.Info("Пользователь ввел аргументы команды add",
			slog.String("user", message.From.String()),
			slog.Int64("id", message.Chat.ID),
			slog.String("values", arg))
		if err != nil {
			// validateValues возвращает ошибку, если аргументы невалидны
			// В зависимости от типа ошибки отправляется соответствующее сообщение
			switch err {
			case models.ErrInvalidNumberArgument:
				msg_text := "Неверный формат команды.\n" +
					"Комнату можно создать 2 способами:\n" +
					"1. Пошагово, введя команду /add и следуя инструкциям бота\n" +
					"2. Введя команду /add и через пробел указать параметры, например:\n" +
					"\"/add ABCDEF никнейм карта описание\""
				msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
					return fmt.Errorf("%s: %w", path, err)
				}
				slog.Info("Пользователь ввел неверное количество аргументов команды add",
					slog.String("user", message.From.String()),
					slog.Int64("id", message.Chat.ID))
				return nil

			case models.ErrInvalidCode:
				msg_text := "Неверный код комнаты.\n" +
					"Код комнаты должен состоять из 6 латинских букв, " +
					"последняя буква может быть только F, G, Q, f, g или q"
				msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
					return fmt.Errorf("%s: %w", path, err)
				}
				slog.Info("Пользователь ввел неверный код комнаты",
					slog.String("user", message.From.String()),
					slog.Int64("id", message.Chat.ID),
					slog.String("code", values[0]))
				return nil

			case models.ErrRoomAlreadyExist:
				msg_text := "Комната с таким кодом уже существует"
				msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
					return fmt.Errorf("%s: %w", path, err)
				}
				slog.Info("Пользователь ввел код существующей комнаты",
					slog.String("user", message.From.String()),
					slog.Int64("id", message.Chat.ID),
					slog.String("code", values[0]))
				return nil

			case models.ErrInvalidName:
				msg_text := "Слишком длинный никнейм.\n" +
					"Никнейм должен состоять не более чем из 10 символов"
				msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
					return fmt.Errorf("%s: %w", path, err)
				}
				slog.Info("Пользователь ввел слишком длинный никнейм",
					slog.String("user", message.From.String()),
					slog.Int64("id", message.Chat.ID),
					slog.String("name", values[1]))
				return nil

			case models.ErrInvalidMap:
				msg_text := "Слишком длинное название карты.\n" +
					"Название карты должно состоять не более чем из 10 символов"
				msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
					return fmt.Errorf("%s: %w", path, err)
				}
				slog.Info("Пользователь ввел слишком длинное название карты",
					slog.String("user", message.From.String()),
					slog.Int64("id", message.Chat.ID),
					slog.String("map", values[2]))
				return nil

			case models.ErrInvalidMode:
				msg_text := "Слишком длинное описание режима игры.\n" +
					"Описание должно состоять не более чем из 10 символов"
				msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
					return fmt.Errorf("%s: %w", path, err)
				}
				slog.Info("Пользователь ввел слишком длинное описание режима игры",
					slog.String("user", message.From.String()),
					slog.Int64("id", message.Chat.ID),
					slog.String("mode", values[3]))
				return nil

			default:
				msg_text := "При выполнении команды произошла неожиданная ошибка"
				msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
					return fmt.Errorf("%s: %w", path, err)
				}
				slog.Error("При выполнении команды произошла неожиданная ошибка",
					slog.String("user", message.From.String()),
					slog.Int64("id", message.Chat.ID),
					slog.String("values", arg),
					slog.String("error", err.Error()))
				return nil
			}

		}
		if room != nil {
			room.ID = message.Chat.ID
		}
	}

	// По идее в данном месте room никогда не может быть = nil
	// Но на всякий случай проверяем чтобы не допустить ошибки сохранения в БД
	if room == nil {
		slog.Error("Сформирована пустая модель комнаты")
		return fmt.Errorf("%s: %w", path, fmt.Errorf("room is nil"))
	}

	// Сохранение комнаты в БД и возвращение сообщения пользователю
	err = b.rep.SaveRoom(room)
	if err != nil {
		slog.Error("Ошибка добавления комнаты в БД")
		return fmt.Errorf("%s: %w", path, err)
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, "*Комната успешно добавлена*\n\n"+
		"Для того чтобы не засорять бота неактивными комнатами, не забудь её удалить когда "+
		"закончишь играть")
	msg.ReplyMarkup = list_kb
	msg.ParseMode = "markdownV2"
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}
	slog.Info("Пользователь успешно создал комнату",
		slog.String("user", message.From.String()),
		slog.Int64("id", room.ID),
		slog.String("room", room.Code),
		slog.String("name", room.Hoster),
		slog.String("map", room.Map),
		slog.String("mode", room.Mode))

	// Сохранение статуса "room" пользователя в БД
	// в статусе указан код комнаты
	err = b.rep.SaveUserStatus(message.Chat.ID, "room", room.Code)
	if err != nil {
		slog.Error("Ошибка сохранения в БД данных о созданной пользователем комнате")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Обновление данных в модели хостера или создание новой модели
	host, err := b.rep.GetHoster(message.Chat.ID)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о хостере")
		return fmt.Errorf("%s: %w", path, err)
	}
	// Если модель хостера не найдена, то создается новая
	if host == nil {
		host = &models.Hoster{
			ID:        message.Chat.ID,
			Name:      room.Hoster,
			Followers: []models.User{},
			LastSend:  time.Now().Add(-12 * time.Hour),
		}
		slog.Info("Создана новая модель хостера",
			slog.String("user", message.From.String()),
			slog.Int64("id", room.ID),
			slog.String("host", host.Name))
	} else {
		// Если модель хостера найдена, то обновляется ник на актуальный
		// другие данные сейчас обновлять не требуется
		host.Name = room.Hoster
	}

	// Сохранение модели хостера в БД
	err = b.rep.SaveHoster(host)
	if err != nil {
		slog.Error("Ошибка сохранения в БД данных о хостере")
		return fmt.Errorf("%s: %w", path, err)
	}
	slog.Info("В модели хостера успешно обновлен ник",
		slog.Int64("id", room.ID),
		slog.String("host", host.Name))

	return nil
}

// Валидация аргументов команды /add
func (b *Telegram) validateValues(values []string) (*models.Room, error) {
	path := "service.telegram.validateValues"

	// Проверка на количество аргументов
	if len(values) != 4 {
		return nil, models.ErrInvalidNumberArgument
	}

	// Проверка на валидность кода комнаты
	code := values[0]
	match, _ := regexp.MatchString("^[a-zA-Z]{5}[fgqFGQ]$", code)
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
		ID:         0,
		Warning:    false,
	}

	return &room, nil
}

// Функция обрабатывает сообщения с кодом комнаты при пошаговом создании комнаты
// на первом этапе обрабатывается код комнаты, который должен содержаться в message.Text
func (b *Telegram) addDraftRoom(message *tgbotapi.Message) error {
	const path = "service.telegram.addDraftRoom"

	code := message.Text
	// Проверка корректности нового кода комнаты
	match, _ := regexp.MatchString("^[a-zA-Z]{5}[fgqFGQ]$", code)
	if !match {
		slog.Info("Пользователь ввел неверный код комнаты",
			slog.String("user", message.Chat.UserName),
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
			slog.Info("Пользователь ввел код существующей комнаты",
				slog.String("user", message.Chat.UserName),
				slog.Int64("id", message.Chat.ID),
				slog.String("code", code))
			return models.ErrRoomAlreadyExist
		}
	}

	// Создание черновика комнаты
	// часть полей заполняется пустыми значениями, они будут получены позднее
	room := models.Room{
		Code:       code,
		Mode:       "",
		Hoster:     "",
		Map:        "",
		Descrition: "",
		Time:       time.Now(),
		ID:         message.Chat.ID,
		Warning:    false,
	}

	// Сохранение черновика комнаты в БД
	room.Code = code
	err = b.rep.SaveDraftRoom(&room)
	if err != nil {
		slog.Error("Ошибка добавления комнаты в БД")
		return fmt.Errorf("%s: %w", path, err)
	}

	// По аналогии со статусом "room" сохраняется статус "draft_room"
	err = b.rep.SaveUserStatus(message.Chat.ID, "draft_room", room.Code)

	slog.Info("Пользователь успешно создал черновик комнаты",
		slog.String("user", message.Chat.UserName),
		slog.Int64("id", message.Chat.ID))
	return err
}

// Второй этап пошагового создания комнаты
// В сообщении message.Text должен содержаться ник хостера
func (b *Telegram) addHostName(message *tgbotapi.Message, name string) error {
	const path = "service.telegram.addHostName"

	// Проверка на длину ника
	length := utf8.RuneCountInString(name)
	if length > 10 {
		slog.Info("Пользователь ввел слишком длинный никнейм",
			slog.String("user", message.Chat.UserName),
			slog.Int64("id", message.Chat.ID),
			slog.String("name", name))
		return models.ErrInvalidName
	}

	// Загрузить код черновика комнаты из статуса "draft_room"
	draft_room_code, err := b.rep.GetUserStatus(message.Chat.ID, "draft_room")
	if err != nil {
		slog.Error("Ошибка чтения из БД кода создаваемой комнаты")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Загрузить черновик комнаты из БД
	var room *models.Room
	room, err = b.rep.GetDraftRoom(draft_room_code)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о создаваемой комнате")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Скорректировать в черновике комнаты ник хостера
	room.Hoster = name

	// Сохранить скорректированную комнату в базу данных
	err = b.rep.SaveDraftRoom(room)
	if err != nil {
		slog.Error("Ошибка добавления комнаты в БД")
		return fmt.Errorf("%s: %w", path, err)
	}

	slog.Info("Пользователь успешно добавил никнейм в черновик комнаты",
		slog.String("user", message.Chat.UserName),
		slog.Int64("id", message.Chat.ID),
		slog.String("name", name))
	return nil

}

// Третий этап пошагового создания комнаты
// В сообщении message.Text должно содержаться название карты
func (b *Telegram) addMapName(message *tgbotapi.Message, mapa string) error {
	const path = "service.telegram.addMapName"

	// Проверка на длину названия карты
	length := utf8.RuneCountInString(mapa)
	if length > 10 {
		slog.Info("Пользователь ввел слишком длинное название карты",
			slog.String("user", message.Chat.UserName),
			slog.Int64("id", message.Chat.ID),
			slog.String("map", mapa))
		return models.ErrInvalidMap
	}

	// Загрузить код черновика комнаты из статуса "draft_room"
	draft_room_code, err := b.rep.GetUserStatus(message.Chat.ID, "draft_room")
	if err != nil {
		slog.Error("Ошибка чтения из БД кода создаваемой комнаты")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Загрузить черновик комнаты из БД
	var room *models.Room
	room, err = b.rep.GetDraftRoom(draft_room_code)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о создаваемой комнате")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Скорректировать название карты
	room.Map = mapa

	// Сохранить скорректированную комнату в базу данных
	err = b.rep.SaveDraftRoom(room)
	if err != nil {
		slog.Error("Ошибка добавления комнаты в БД")
		return fmt.Errorf("%s: %w", path, err)
	}

	slog.Info("Пользователь успешно добавил название карты в черновик комнаты",
		slog.String("user", message.Chat.UserName),
		slog.Int64("id", message.Chat.ID),
		slog.String("map", mapa))
	return nil
}

// Четвертый этап пошагового создания комнаты
// В сообщении message.Text должно содержаться описание режима игры
func (b *Telegram) addGameMode(message *tgbotapi.Message, mode string) error {
	const path = "service.telegram.addGameMode"

	// Проверка на длину описания режима игры
	length := utf8.RuneCountInString(mode)
	if length > 10 {
		slog.Info("Пользователь ввел слишком длинное описание режима игры",
			slog.String("user", message.Chat.UserName),
			slog.Int64("id", message.Chat.ID),
			slog.String("mode", mode))
		return models.ErrInvalidMode
	}

	// Загрузить код черновика комнаты из статуса "draft_room"
	draft_room_code, err := b.rep.GetUserStatus(message.Chat.ID, "draft_room")
	if err != nil {
		slog.Error("Ошибка чтения из БД кода создаваемой комнаты")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Загрузить черновик комнаты из БД
	var room *models.Room
	room, err = b.rep.GetDraftRoom(draft_room_code)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о создаваемой комнате")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Скорректировать описание режима игры
	room.Mode = mode
	room.Time = time.Now()

	// Сохранить комнату в базу данных
	// Учитывая, что все параметры черновика комнаты сформированы, комната сохраняется
	// в таблицу rooms, а черновик удаляется
	err = b.rep.SaveRoom(room)
	if err != nil {
		slog.Error("Ошибка добавления комнаты в БД")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Сохранить статус "room" пользователя в БД с кодом комнаты
	err = b.rep.SaveUserStatus(message.Chat.ID, "room", room.Code)
	if err != nil {
		slog.Error("Ошибка сохранения в БД статуса о созданной пользователем комнате")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Обнулить статус "draft_room" пользователя из БД
	err = b.rep.SaveUserStatus(message.Chat.ID, "draft_room", "")
	if err != nil {
		slog.Error("Ошибка удаления из БД статуса о черновике комнаты")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Удалить черновик комнаты из БД
	err = b.rep.DeleteDraftRoom(draft_room_code)
	if err != nil {
		slog.Error("Ошибка удаления комнаты из БД")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Сохранить ник хостера в статусе "host_name"
	err = b.rep.SaveUserStatus(message.Chat.ID, "host_name", room.Hoster)
	if err != nil {
		slog.Error("Ошибка сохранения в БД никнейма хостера")
		return fmt.Errorf("%s: %w", path, err)
	}

	slog.Info("Пользователь успешно пошагово создал комнату",
		slog.String("user", message.Chat.UserName),
		slog.Int64("id", message.Chat.ID),
		slog.String("room", room.Code),
		slog.String("name", room.Hoster),
		slog.String("map", room.Map),
		slog.String("mode", room.Mode))

	// Сохранить данные в модели хостера
	host, err := b.rep.GetHoster(message.Chat.ID)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о хостере")
		return fmt.Errorf("%s: %w", path, err)
	}
	// Если модель хостера не найдена, то создается новая
	if host == nil {
		host = &models.Hoster{
			ID:        message.Chat.ID,
			Name:      room.Hoster,
			Followers: []models.User{},
			LastSend:  time.Now().Add(-12 * time.Hour),
		}
		slog.Info("Создана новая модель хостера",
			slog.String("user", message.Chat.UserName),
			slog.Int64("id", room.ID),
			slog.String("host", host.Name))
	} else {
		host.Name = room.Hoster
	}

	err = b.rep.SaveHoster(host)
	if err != nil {
		slog.Error("Ошибка сохранения в БД данных о хостере")
		return fmt.Errorf("%s: %w", path, err)
	}
	slog.Info("В модели хостера успешно обновлен ник",
		slog.Int64("id", room.ID),
		slog.String("host", host.Name))

	return nil
}
