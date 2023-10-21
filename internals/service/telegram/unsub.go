package telegram

import (
	"among-us-roombot/internals/models"
	"fmt"

	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Telegram) handleUnsubscribe(message *tgbotapi.Message) error {
	const path = "service.telegram.unsubscribe.handleUnsubscribe"

	var user *models.Follower
	user, err := b.rep.GetUser(message.Chat.ID)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о пользователе",
			slog.String("user", message.From.String()),
			slog.Int64("id", message.Chat.ID),
			slog.String("path", path))
	}

	if len(user.Hosters) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Ты ни на кого не подписан")
		msg.ReplyMarkup = list_kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("error send message to user")
			return fmt.Errorf("%s: %w", path, err)
		}
		return nil
	}

	kb := make_unsubscribe_kb(b, user.Hosters)
	msg := tgbotapi.NewMessage(message.Chat.ID, "Нажми на кнопку "+
		"с ником хостера, от которого хочешь отписаться:")
	msg.ReplyMarkup = kb
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	return nil
}

func (b *Telegram) areYouShure(userID int64, hostID int64) error {
	const path = "service.telegram.unsubscribe.areYouShure"

	host, err := b.rep.GetHoster(hostID)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о хостере",
			slog.Int64("user_id", userID),
			slog.Int64("host_id", hostID),
			slog.String("path", path))
		return fmt.Errorf("%s: %w", path, err)
	}

	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да", fmt.Sprintf("uns%d", hostID)),
			tgbotapi.NewInlineKeyboardButtonData("Нет", "cancel"),
		),
	)
	msg := tgbotapi.NewMessage(userID, "Ты точно хочешь отписаться от "+host.Name+"?")
	msg.ReplyMarkup = kb
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	return nil
}

func (b *Telegram) unsubscribe(userID int64, hostID int64) (string, error) {
	const path = "service.telegram.subscribe.subscribe"

	// Загрузка модели подписчика из БД
	user, err := b.rep.GetUser(userID)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о пользователе",
			slog.Int64("id", userID),
			slog.String("path", path))
		return "Ошибка удаления подписки", fmt.Errorf("%s: %w", path, err)
	}

	// Загрузка модели хостера из БД
	host, err := b.rep.GetHoster(hostID)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о хостере",
			slog.Int64("id", userID),
			slog.String("path", path))
		return "Ошибка удаления подписки", fmt.Errorf("%s: %w", path, err)
	}

	// Удаление хостера из модели подписчика
	var userList models.UserList
	userList.Users = user.Hosters
	index := userList.FindUserIndexByID(userList.Users, hostID)
	if index == -1 {
		slog.Error("Хостер не найден в модели подписчика",
			slog.Int64("user_id", userID),
			slog.Int64("host_id", hostID),
			slog.String("path", path))
		return "Ошибка удаления подписки (в БД не найдена запись о наличии подписки)",
			fmt.Errorf("%s: %s", path, "хостер не найден в модели подписчика")
	}
	userList.Users = append(userList.Users[:index], userList.Users[index+1:]...)
	user.Hosters = userList.Users
	err = b.rep.SaveUser(user)
	if err != nil {
		slog.Error("Ошибка сохранения модели подписчика в БД",
			slog.Int64("id", userID),
			slog.String("path", path))
		return "Ошибка удаления подписки", fmt.Errorf("%s: %w", path, err)
	}
	slog.Info("Хостер успешно удален из модели подписчика",
		slog.Int64("id", userID),
		slog.String("host", host.Name))

	// Удаление подписчика из модели хостера
	userList.Users = host.Followers
	index = userList.FindUserIndexByID(userList.Users, userID)
	if index == -1 {
		slog.Error("Подписчик не найден в модели хостера",
			slog.Int64("user_id", userID),
			slog.Int64("host_id", hostID),
			slog.String("path", path))
		return "Ошибка удаления подписки (в БД не найдена запись о наличии подписки)", fmt.Errorf("%s: %s", path, "подписчик не найден в модели хостера")
	}
	userList.Users = append(userList.Users[:index], userList.Users[index+1:]...)
	host.Followers = userList.Users
	err = b.rep.SaveHoster(host)
	if err != nil {
		slog.Error("Ошибка сохранения модели хостера в БД",
			slog.Int64("user_id", userID),
			slog.Int64("host_id", hostID),
			slog.String("path", path))
		return "Ошибка удаления подписки", fmt.Errorf("%s: %w", path, err)
	}
	slog.Info("Подписчик успешно удален из модели хостера",
		slog.Int64("user_id", userID),
		slog.Int64("host_id", host.ID),
		slog.String("host", host.Name))

	return host.Name, nil
}
