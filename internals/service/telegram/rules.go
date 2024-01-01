package telegram

import (
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Обработка команды /rules
func (b *Telegram) handleRules(message *tgbotapi.Message) error {
	const path = "service.telegram.rules"

	msg := tgbotapi.NewMessage(message.Chat.ID, "*Правила:*\n\n"+
		"1\\. Не забывай удалять неактуальные румы\\.\n\n"+
		"2\\. Код, ник, название карты, описание режима и текст рассылки "+
		"не должны содержать оскорблений\\.\n\n"+
		"3\\. Не спамь обратной связью\\. Пожалуйста отправляй только вопросы, замечания "+
		"и предложения по работе бота\\. Не нужно отправлять сообщения вида \"Привет\", "+
		"\"Как дела?\" и т\\.д\\.\n\n"+
		"4\\. Не спамь рассылкой\\. Никого не оскорбляй и не матерись в рассылке\\.\n\n"+
		"За нарушение правил может последовать бан\\.\n\n")
	msg.ReplyMarkup = list_kb
	msg.ParseMode = "MarkdownV2"
	_, err := b.bot.Send(msg)
	if err != nil {
		slog.Error("error send message to user")
		return fmt.Errorf("%s: %w", path, err)
	}

	slog.Info("Пользователю отправлены правила",
		slog.String("user", message.From.String()),
		slog.Int64("id", message.Chat.ID))
	return nil
}
