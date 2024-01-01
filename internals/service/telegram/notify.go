package telegram

import (
	"among-us-roombot/internals/models"
	"fmt"
	"log/slog"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Функция рассылки сообщений подписчикам
func (b *Telegram) handleNotify(message *tgbotapi.Message) error {
	const path = "service.telegram.notufy.handleNotify"

	// Загрузка модели хостера из БД
	var host *models.Hoster
	host, err := b.rep.GetHoster(message.Chat.ID)
	if err != nil {
		slog.Error("Ошибка чтения из БД модели хостера")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Проверка наличия модели хостера в БД, если её нет, то пользователь никогда не хостил
	if host == nil {
		msg_text := "Не удлось выполнить команду рассылки.\n\n" +
			"Скорее всего ты никогда не хостил руму, поэтому у тебя не может " +
			"быть подписчиков"
		msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
		msg.ReplyMarkup = list_kb
		_, err := b.bot.Send(msg)
		if err != nil {
			slog.Error("error send message to user")
			return fmt.Errorf("%s: %w", path, err)
		}
		slog.Info("Попытка отправить рассылку пользователем у которого нет модели хостера",
			slog.String("user", message.From.String()),
			slog.Int64("id", message.Chat.ID))
		return nil
	}

	// Проверка наличия подписчиков у хостера, если их нет, то рассылка не имеет смысла
	if len(host.Followers) == 0 {
		err = sendImage(b, message.Chat.ID, "notify.png")
		if err != nil {
			slog.Warn("Не удалось отправить картинку")
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, "У тебя нет подписчиков 😢")
		msg.ReplyMarkup = list_kb
		_, err := b.bot.Send(msg)
		if err != nil {
			slog.Error("error send message to user")
			return fmt.Errorf("%s: %w", path, err)
		}

		slog.Info("Попытка отправить рассылку пользователем у которого нет подписчиков",
			slog.String("user", message.From.String()),
			slog.Int64("id", message.Chat.ID))
		return nil
	}

	// Проверка времени предыдущей рассылки, запрещено направлять рассылку чаще чем раз в 6 часов
	delta := time.Now().Unix() - host.LastSend.Unix()
	if delta < (60 * 60 * 6) {
		t := time.Unix((60*60*6)-delta, 0)
		t_str := fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
		msg_text := fmt.Sprintf("Рассылку можно отправлять не чаще чем раз в 6 часов, "+
			"следующая рассылка возможна через %s", t_str)
		msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
		msg.ReplyMarkup = list_kb
		_, err := b.bot.Send(msg)
		if err != nil {
			slog.Error("error send message to user")
			return fmt.Errorf("%s: %w", path, err)
		}
		slog.Info("Попытка отправить рассылку пользователем, который уже отправлял рассылку",
			slog.String("user", message.From.String()),
			slog.Int64("id", message.Chat.ID))

		return nil
	}

	// Проверка наличия активной комнаты у хостера
	room_code, err := b.rep.GetUserStatus(message.Chat.ID, "room")
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о созданной пользователем комнате")
		return fmt.Errorf("%s: %w", path, err)
	}
	// Если активной румы нет, то просим направить текст рассылки в свободной форме
	if room_code == "" {
		msg_text := "Пришли мне текст рассылки (файлы и фото отправлять не умею), " +
			"который будет направлен твоим подписчикам:\n"
		msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
		msg.ReplyMarkup = cancel_kb
		_, err := b.bot.Send(msg)
		if err != nil {
			slog.Error("error send message to user")
			return fmt.Errorf("%s: %w", path, err)
		}
		b.rep.SaveUserStatus(message.Chat.ID, "status", "wait_post")
		slog.Info("Хостер без румы запустил команду рассылки, жду текст",
			slog.String("user", message.From.String()),
			slog.Int64("id", message.Chat.ID))
		return nil

	} else {
		// Если активная рума есть, то формируется типовой текст рассылки
		room, err := b.rep.GetRoom(room_code)
		if err != nil {
			slog.Error("Ошибка чтения из БД данных о созданной пользователем комнате")
			return fmt.Errorf("%s: %w", path, err)
		}
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Отправить шаблон", "send_template"),
				tgbotapi.NewInlineKeyboardButtonData("Отменить рассылку", "cancel"),
			),
		)
		draft_post := fmt.Sprintf("_Привет!\nЗаходи ко мне поиграть, "+
			"я играю на карте %s, режим %s, код:_\n\n`%s`", room.Map, room.Mode, room.Code)
		msg_text := "Пришли мне текст рассылки (файлы и фото отправлять не умею), " +
			"который будет направлен твоим подписчикам,\n" +
			"*или* нажми на кнопку \"**Отправить шаблон**\", чтобы отправить " +
			"следующее сообщение:\n\n--------------------------------------------------\n" + draft_post
		msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("error send message to user")
			return fmt.Errorf("%s: %w", path, err)
		}
		b.rep.SaveUserStatus(message.Chat.ID, "status", "wait_post")
		slog.Info("Хостер с активной румой запустил команду рассылки, жду текст",
			slog.String("hoster", host.Name),
			slog.Int64("id", message.Chat.ID))
		return nil
	}
}

// Функция выполняет рассылку сообщения подписчикам
// post - текст рассылки
func (b *Telegram) sendPost(message *tgbotapi.Message, post string) error {
	const path = "service.telegram.notify.sendPost"

	// Загрузка модели хостера из БД
	host, err := b.rep.GetHoster(message.Chat.ID)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о хостере")
		return fmt.Errorf("%s: %w", path, err)
	}
	if host == nil {
		slog.Error("Хостер не найден в БД")
		return fmt.Errorf("%s: %s", path, "хостер не найден в БД")
	}

	// Подписчики хостера
	followers := host.Followers
	if len(followers) == 0 {
		slog.Warn("Хостер попытался отправить рассылку, но у него нет подписчиков",
			slog.String("hoster", host.Name),
			slog.Int64("id", message.Chat.ID))
		return nil
	} else {
		post = fmt.Sprintf("*%s* отправил сообщение:\n\n"+
			"--------------------------------------------------\n%s", host.Name, post)

		// В отдельной горутине запускается функция рассылки сообщений
		// чтобы не блокировать работу бота
		go b.notify(followers, host.ID, post)
		if err != nil {
			slog.Error("Ошибка отправки сообщения подписчикам")
			return fmt.Errorf("%s: %w", path, err)
		}

		// Обновление времени последней рассылки
		host.LastSend = time.Now()
		err = b.rep.SaveHoster(host)
		if err != nil {
			slog.Error("Ошибка сохранения модели хостера в БД")
			return fmt.Errorf("%s: %w", path, err)
		}

		msg_text := fmt.Sprintf("Рассылка успешно отправлена *%d* подписчикам", len(followers))
		msg := tgbotapi.NewMessage(message.Chat.ID, msg_text)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = list_kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("error send message to user")
			return fmt.Errorf("%s: %w", path, err)
		}
		slog.Info("Хостер успешно отправил рассылку",
			slog.String("hoster", host.Name),
			slog.Int64("id", message.Chat.ID),
			slog.String("post", post),
			slog.Int("followers", len(followers)))

		return nil
	}

}

// Функция рассылает сообщение подписчикам
// followers - список подписчиков в формате []models.User
// hostID - ID хостера
// post - текст рассылки
func (b *Telegram) notify(followers []models.User, hostID int64, post string) {
	const path = "service.telegram.notify.notify"

	for _, follower := range followers {
		msg := tgbotapi.NewMessage(follower.ID, post)
		msg.ParseMode = "Markdown"
		_, err := b.bot.Send(msg)
		if err != nil {
			switch err.Error() {
			case "Forbidden: bot was blocked by the user":
				slog.Warn("Обнаружена блокировка бота")
				txt, err := b.unsubscribe(follower.ID, hostID)
				if err != nil {
					slog.Warn("Ошибка автоматической отписки от хостера",
						slog.Int64("hostID", hostID),
						slog.Int64("userID", follower.ID),
						slog.String("error", err.Error()),
						slog.String("txt", txt))
				}
			case "Forbidden: user is deactivated":
				slog.Warn("Пользователь TG удалён")
				txt, err := b.unsubscribe(follower.ID, hostID)
				if err != nil {
					slog.Warn("Ошибка автоматической отписки от хостера",
						slog.Int64("hostID", hostID),
						slog.Int64("userID", follower.ID),
						slog.String("error", err.Error()),
						slog.String("txt", txt))
				}
			default:
				slog.Error("Ошибка отправки сообщения подписчику",
					slog.Int64("id", follower.ID),
					slog.String("follower", follower.Name),
					slog.String("error", err.Error()),
					slog.String("path", path))
			}
		}
		// Задержка между отправкой сообщений для избежания блокировки бота
		time.Sleep(time.Millisecond * 300)
	}
}

// Функция отправляет картинку пользователю
// imagePath - путь к картинке
// chatID - ID пользователя
func sendImage(b *Telegram, chatID int64, imagePath string) error {
	// Открываем файл с картинкой
	file, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Создаем объект File для загрузки в Telegram
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	fileBytes := make([]byte, fileInfo.Size())
	_, err = file.Read(fileBytes)
	if err != nil {
		return err
	}
	fileObj := tgbotapi.FileBytes{
		Name:  fileInfo.Name(),
		Bytes: fileBytes,
	}

	// Отправляем картинку пользователю
	msg := tgbotapi.NewPhoto(chatID, fileObj)
	_, err = b.bot.Send(msg)
	if err != nil {
		return err
	}

	return nil
}
