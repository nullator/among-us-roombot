package telegram

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Обработка команды /del
func (b *Telegram) handleDel(message *tgbotapi.Message) error {
	const path = "service.telegram.delete.handleDel"

	// У моманды 2 режима работы: чтение кода из аргументов сообщения или удаление через диалог
	arg := message.CommandArguments()

	// Если аргументы есть, то пытаемся удалить комнату, проверяя корректность аргумента
	// Удалять комнату с аргументами может только администратор,
	// поэтому проверяем наличие администратора в списке администраторов
	if arg != "" {
		match, _ := regexp.MatchString("^[a-zA-Z]{6}$", arg)
		if !match {
			return nil
		}
		code := strings.ToUpper(arg)

		// Загрузка администраторов из переменной окружения и проверка наличия прав администратора
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

	// Если аргументов нет, то пытаемся удалить комнату через диалог
	// Проверяем наличие созданной комнаты у пользователя
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
		slog.Info("Пользователь у которого нет созданной комнаты пытается удалить комнату",
			slog.String("user", message.From.String()),
			slog.Int64("id", message.Chat.ID))
		return nil
	}

	// Если комната есть, то запускаем диалог удаления
	// Создается кнопка с кодом удаляемой комнаты и кнопка отмены
	// В дальнейшем при нажалии на кнопку "delete" запускается удаление комнаты
	caption := fmt.Sprintf("Удалить %s", exist_room)
	var kb = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(caption, "delete"),
			tgbotapi.NewInlineKeyboardButtonData("Отмена", "cancel"),
		),
	)

	msg_text := fmt.Sprintf("Уверен что хочешь удалить комнату %s?", exist_room)
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

// Функция удаления комнаты после нажатия на кнопку "delete"
func (b *Telegram) delete(message *tgbotapi.Message) error {
	const path = "service.telegram.delete.delete"

	// Получить из базы код существующей комнаты пользователя
	exist_room, err := b.rep.GetUserStatus(message.Chat.ID, "room")
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о созданной пользователем комнате")
		return fmt.Errorf("%s: %w", path, err)
	}

	// Удалить комнату из базы
	err = b.rep.DeleteRoom(exist_room)
	if err != nil {
		slog.Error("Ошибка удаления комнаты из БД")
		return fmt.Errorf("%s: %w", path, err)
	}
	slog.Info("Комната удалена из БД",
		slog.String("room", exist_room))

	// Оннулить статус пользователя о наличии существующей комнаты
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
		slog.String("user", message.Chat.UserName),
		slog.Int64("id", message.Chat.ID))

	return nil
}

// Функция проверки наличия пользователя в списке администраторов
// id - telegram id пользователя
func (b *Telegram) isAdmin(list []string, id string) bool {
	for _, v := range list {
		if v == id {
			return true
		}
	}
	return false
}
