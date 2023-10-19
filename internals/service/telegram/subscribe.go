package telegram

import (
	"among-us-roombot/internals/models"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Telegram) handleSubscribe(message *tgbotapi.Message) error {
	const path = "internals/service/telegram/subscribe.go"

	var rooms models.RoomList
	rooms, err := b.rep.GetRoomList()
	if err != nil {
		slog.Error("Ошибка чтения из БД списка комнат")
		return fmt.Errorf("%s: %w", path, err)
	}

	// rooms = append(rooms, models.Room{
	// 	Code:   "AAAAAA",
	// 	Hoster: "hoster1",
	// 	Map:    "Skeld",
	// 	Mode:   "Классика",
	// })

	// rooms = append(rooms, models.Room{
	// 	Code:   "BBBBBB",
	// 	Hoster: "hoster2",
	// 	Map:    "Polus",
	// 	Mode:   "Прятки",
	// })

	kb := make_subscribe_kb(b, message.Chat.ID, rooms)
	msg := tgbotapi.NewMessage(message.Chat.ID, "Выбери хостера, на которого хочешь подписаться")
	msg.ReplyMarkup = kb
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	return nil
}

func (b *Telegram) subscribe(
	callback *tgbotapi.CallbackQuery, userID int64, hostID int64) error {
	const path = "internals/service/telegram/subscribe.go"

	// Загрузка модели подписчика из БД
	user, err := b.rep.GetUser(userID)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о пользователе")
		return fmt.Errorf("%s: %w", path, err)
	}
	slog.Debug("Пользователь успешно загружен из БД",
		slog.Any("user", user))
	// Если подписчик не найден в БД, создается новый
	if user == nil {
		slog.Info("Пользователь не найден в БД, создаю нового",
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
	slog.Debug("Хостер успешно загружен из БД",
		slog.Any("host", host))

	// Создание модели на основе модели хостера
	newHost := models.User{
		ID:   hostID,
		Name: host.Name,
	}

	// Добавление хостера в список подписок пользователя
	user.Hosters = append(user.Hosters, newHost)
	err = b.rep.SaveUser(user)
	if err != nil {
		slog.Error("Ошибка сохранения пользователя в БД после добавления подписчика")
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
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID,
		"Успешно выполнена подписка на "+host.Name)
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	return nil
}
