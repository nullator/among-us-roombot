package telegram

import (
	"among-us-roombot/internals/models"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Обработка команды /subscribe
func (b *Telegram) handleSubscribe(message *tgbotapi.Message) error {
	const path = "service.telegram.subscribe.handleSubscribe"

	// Из БД загружается список активных комнат
	var rooms models.RoomList
	rooms, err := b.rep.GetRoomList()
	if err != nil {
		slog.Error("Ошибка чтения из БД списка комнат")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Если список комнат пуст, то пользователь не может подписаться на хостера
	if len(rooms) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Сейчас нет активных хостеров, "+
			"не на кого подписываться")
		msg.ReplyMarkup = list_kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("error send message to user")
			return fmt.Errorf("%s: %w", path, err)
		}
		return nil
	}

	// Создание клавиатуры с кнопками для подписки
	kb := make_subscribe_kb(b, message.Chat.ID, rooms)
	msg := tgbotapi.NewMessage(message.Chat.ID, "Нажми на кнопку "+
		"с ником хостера, на которого хочешь подписаться:")
	msg.ReplyMarkup = kb
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	return nil
}

// Подписка на хостера
// userID - ID пользователя который выполняет подписку
// hostID - ID хостера
func (b *Telegram) subscribe(
	callback *tgbotapi.CallbackQuery, userID int64, hostID int64) error {
	const path = "service.telegram.subscribe.subscribe"

	// Загрузка модели подписчика из БД
	user, err := b.rep.GetUser(userID)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о пользователе")
		return fmt.Errorf("%s: %w", path, err)
	}
	// Если подписчик не найден в БД, создается новый
	if user == nil {
		slog.Info("Подписчик не найден в БД, создаю нового",
			slog.String("user", callback.From.String()),
			slog.Int64("id", callback.Message.Chat.ID))
		user = &models.Follower{
			ID:      userID,
			Hosters: []models.User{},
		}
	}

	// Загрузка модели хостера из БД
	host, err := b.rep.GetHoster(hostID)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о хостере")
		return fmt.Errorf("%s: %w", path, err)
	}
	if host == nil {
		slog.Error("Хостер не найден в БД")
		return fmt.Errorf("%s: %s", path, "хостер не найден в БД")
	}

	// Создание модели на основе модели хостера
	newHost := models.User{
		ID:   hostID,
		Name: host.Name,
	}

	// Проверка на то, что пользователь уже подписан на хостера
	var userList models.UserList
	userList.Users = user.Hosters
	index := userList.FindUserIndexByID(userList.Users, hostID)
	if index != -1 {
		// TODO: скорректировать сообщение и уведомлять о том что в БД будет изменён ник хостера
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID,
			"Ты уже подписан на этого хостера")
		msg.ReplyMarkup = list_kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("error send message to user")
			return fmt.Errorf("%s: %w", path, err)
		}

		// Если пользователь уже подписан на хостера, в модели подписчика меняется ник хостера
		user.Hosters[index].Name = host.Name
		err = b.rep.SaveUser(user)
		if err != nil {
			slog.Error("Ошибка сохранения пользователя в БД после обновления ника хостера")
			return fmt.Errorf("%s: %w", path, err)
		}
		return nil
	}

	// Добавление хостера в список подписок пользователя
	user.Hosters = append(user.Hosters, newHost)
	err = b.rep.SaveUser(user)
	if err != nil {
		slog.Error("Ошибка сохранения пользователя в БД после добавления подписки")
		return fmt.Errorf("%s: %w", path, err)
	}

	slog.Info("Пользователь подписался на хостера",
		slog.String("user", callback.From.String()),
		slog.Int64("id", callback.Message.Chat.ID),
		slog.String("hoster", host.Name))

	// Создание модели на основе модели подписчика
	newFollower := models.User{
		ID:   userID,
		Name: "",
	}

	// Добавление подписчика в список подписчиков хостера
	host.Followers = append(host.Followers, newFollower)
	err = b.rep.SaveHoster(host)
	if err != nil {
		slog.Error("Ошибка сохранения хостера в БД после добавления подписчика")
		return fmt.Errorf("%s: %w", path, err)
	}

	slog.Info("Хостер получил нового подписчика",
		slog.String("hoster", host.Name),
		slog.Int64("host_id", host.ID),
		slog.String("follower", callback.From.String()),
		slog.Int64("follower_id", callback.Message.Chat.ID))

	// Вывод сообщения о подписке
	msg_text := fmt.Sprintf("Успешно выполнена подписка на %s\n\n"+
		"Если хочешь подписаться на других хостов жми /subscribe\n"+
		"Для отписки жми /unsubscribe", host.Name)
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, msg_text)
	msg.ReplyMarkup = list_kb
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	return nil
}
