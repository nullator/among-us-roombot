package telegram

import (
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Функция обрабатывает полученную команду бота и вызывает соответствующую функцию
func (b *Telegram) handleCommand(message *tgbotapi.Message) error {
	switch message.Command() {
	case cmdStart:
		return b.handleStart(message)
	case cmdList:
		return b.handleList(message)
	case cmdAdd:
		return b.handleAdd(message)
	case cmdDel:
		return b.handleDel(message)
	case cmdEdit:
		return b.handleEdit(message)
	case cmdRules:
		return b.handleRules(message)
	case cmdHelp:
		return b.handleHelp(message)
	case cmdAbout:
		return b.handleStart(message)
	case cmdFeedback:
		return b.handleFeedback(message)
	case cmdSubscribe:
		return b.handleSubscribe(message)
	case cmdUnsubscribe:
		return b.handleUnsubscribe(message)
	case cmdNotify:
		return b.handleNotify(message)
	default:
		b.handleUnknown(message)
		return nil
	}
}

func (b *Telegram) handleUnknown(message *tgbotapi.Message) {
	slog.Info("Получена неизвестная команда",
		slog.String("command", message.Command()),
		slog.String("text", message.Text),
	)
}
