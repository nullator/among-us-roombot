package telegram

import (
	"log/slog"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Telegram) handleButton(update *tgbotapi.Update, button string, id int64) {
	switch button {
	case "delete":
		err := b.delete(update.CallbackQuery.Message)
		if err != nil {
			slog.Error("Ошибка удаления комнаты",
				slog.String("error", err.Error()))
		}

	case "add_time":
		err := b.addTime(update.CallbackQuery.Message)
		if err != nil {
			slog.Error("Ошибка добавления времени",
				slog.String("error", err.Error()))
		}

	case "save_old_name":
		slog.Info("Зафиксировано нажатие на кнопку использования старого ника хостера",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))
		old_name, err := b.rep.GetUserStatus(id, "host_name")
		slog.Debug("Получен старый ник хостера из БД", slog.String("host_name", old_name))
		if err != nil {
			slog.Error("Ошибка получения из БД данных о старом нике хостера",
				slog.String("error", err.Error()))
			msg := tgbotapi.NewMessage(id, "Произошла ошибка при выполнении команды")
			msg.ReplyMarkup = list_kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}
		}
		err = b.addHostName(update.CallbackQuery.Message, old_name)
		if err != nil {
			slog.Error("Ошибка добавления ника хостера",
				slog.String("error", err.Error()))
			msg := tgbotapi.NewMessage(id, "Произошла ошибка при выполнении команды")
			msg.ReplyMarkup = list_kb
			_, err = b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}

		} else {
			kb := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("🚀 Skeld", "skeld"),
					tgbotapi.NewInlineKeyboardButtonData("⛄ Polus", "polus"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("🍄 Fungle", "fungle"),
					tgbotapi.NewInlineKeyboardButtonData("🛩️ Airship", "airship"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("🏢 Mira HQ", "mira"),
					tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "cancel"),
				),
			)
			msg := tgbotapi.NewMessage(id,
				"Выбери название карты или введи свой вариант "+
					"(не более 10 символов):\n")
			msg.ReplyMarkup = kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}
			b.rep.SaveUserStatus(id, "status", "wait_mapname")

		}

	case "skeld":
		slog.Info("Зафиксировано нажатие на кнопку выбора карты Skeld",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		b.handleMap(update, id, "Skeld")

	case "polus":
		slog.Info("Зафиксировано нажатие на кнопку выбора карты Polus",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		b.handleMap(update, id, "Polus")

	case "fungle":
		slog.Info("Зафиксировано нажатие на кнопку выбора карты Fungle",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		b.handleMap(update, id, "Fungle")

	case "airship":
		slog.Info("Зафиксировано нажатие на кнопку выбора карты Airship",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		b.handleMap(update, id, "Airship")

	case "mira":
		slog.Info("Зафиксировано нажатие на кнопку выбора карты Mira HQ",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		b.handleMap(update, id, "Mira HQ")

	case "classic":
		slog.Info("Зафиксировано нажатие на кнопку выбора режима Classic",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		b.handleMode(update, id, "Классика")

	case "hide":
		slog.Info("Зафиксировано нажатие на кнопку выбора режима Hide and Seek",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		b.handleMode(update, id, "Прятки")

	case "mods":
		slog.Info("Зафиксировано нажатие на кнопку выбора режима Mods",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		b.handleMode(update, id, "Моды")

	case "change_code":
		slog.Info("Зафиксировано нажатие на кнопку изменения кода комнаты",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))
		err := b.rep.SaveUserStatus(id, "status", "edit_code")
		if err != nil {
			slog.Error("Ошибка сохранения в БД данных о статусе пользователя",
				slog.String("error", err.Error()))
			msg := tgbotapi.NewMessage(id, "Произошла ошибка при выполнении команды")
			msg.ReplyMarkup = list_kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}
		}
		msg := tgbotapi.NewMessage(id, "Отправь мне новый код комнаты:")
		msg.ReplyMarkup = cancel_kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("Ошибка отправки сообщения",
				slog.String("error", err.Error()))
		}

	case "change_map":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия карты",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))
		err := b.rep.SaveUserStatus(id, "status", "change_map")
		if err != nil {
			slog.Error("Ошибка сохранения в БД данных о статусе пользователя",
				slog.String("error", err.Error()))
			msg := tgbotapi.NewMessage(id, "Произошла ошибка при выполнении команды")
			msg.ReplyMarkup = list_kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}
		}
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🚀 Skeld", "change_skeld"),
				tgbotapi.NewInlineKeyboardButtonData("⛄ Polus", "change_polus"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🍄 Fungle", "change_fungle"),
				tgbotapi.NewInlineKeyboardButtonData("🛩️ Airship", "change_airship"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏢 Mira HQ", "change_mira"),
				tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "cancel"),
			),
		)
		msg := tgbotapi.NewMessage(id,
			"Выбери название карты или введи свой вариант "+
				"(не более 10 символов):\n")
		msg.ReplyMarkup = kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("Ошибка отправки сообщения",
				slog.String("error", err.Error()))
		}

	case "change_skeld":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия карты на Skeld",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		b.handleNewMap(update, id, "Skeld")

	case "change_polus":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия карты на Polus",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		b.handleNewMap(update, id, "Polus")

	case "change_fungle":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия карты на Fungle",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		b.handleNewMap(update, id, "Fungle")

	case "change_airship":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия карты на Airship",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		b.handleNewMap(update, id, "Airship")

	case "change_mira":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия карты на Mira",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		b.handleNewMap(update, id, "Mira HQ")

	case "change_hoster":
		slog.Info("Зафиксировано нажатие на кнопку изменения ника хостера",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))
		err := b.rep.SaveUserStatus(id, "status", "change_hoster")
		if err != nil {
			slog.Error("Ошибка сохранения в БД данных о статусе пользователя",
				slog.String("error", err.Error()))
			msg := tgbotapi.NewMessage(id, "Произошла ошибка при выполнении команды")
			msg.ReplyMarkup = list_kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}
		}
		msg := tgbotapi.NewMessage(id, "Отправь мне новый ник хостера:")
		msg.ReplyMarkup = cancel_kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("Ошибка отправки сообщения",
				slog.String("error", err.Error()))
		}

	case "change_description":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия режима",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))
		err := b.rep.SaveUserStatus(id, "status", "change_description")
		if err != nil {
			slog.Error("Ошибка сохранения в БД данных о статусе пользователя",
				slog.String("error", err.Error()))
			msg := tgbotapi.NewMessage(id, "Произошла ошибка при выполнении команды")
			msg.ReplyMarkup = list_kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}
		}
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("👨‍🎓 Классика", "change_classic"),
				tgbotapi.NewInlineKeyboardButtonData("🧌 Прятки", "change_hide"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🛠️ Моды", "change_mods"),
				tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "cancel"),
			),
		)
		msg := tgbotapi.NewMessage(id,
			"Выбери режим игры или введи свой вариант "+
				"(не более 10 символов):\n")
		msg.ReplyMarkup = kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("Ошибка отправки сообщения",
				slog.String("error", err.Error()))
		}

	case "change_classic":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия режима на Classic",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		b.handleNewMode(update, id, "Классика")

	case "change_hide":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия режима на Hide and Seek",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		b.handleNewMode(update, id, "Прятки")

	case "change_mods":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия режима на Mods",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		b.handleNewMode(update, id, "Моды")

	case "cancel":
		slog.Info("Зафиксировано нажатие на кнопку отмены команды",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))
		err := b.rep.SaveUserStatus(id, "status", "null")
		if err != nil {
			slog.Error("Ошибка сохранения в БД данных о статусе пользователя",
				slog.String("error", err.Error()))
		}
		msg := tgbotapi.NewMessage(id, "Выполнение команды отменено")
		msg.ReplyMarkup = list_kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("Ошибка отправки сообщения",
				slog.String("error", err.Error()))
		}

	case "roomlist":
		slog.Info("Зафиксировано нажатие на кнопку вывода списка комнат",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))
		err := b.handleList(update.CallbackQuery.Message)
		if err != nil {
			slog.Error("Ошибка вывода списка комнат",
				slog.String("error", err.Error()))
		}

	default:
		cmd := string([]rune(button)[0:3])
		slog.Debug("Получена команда", slog.String("cmd", cmd))
		switch cmd {
		case "sub":
			userID := update.CallbackQuery.Message.Chat.ID
			hostID_str := string([]rune(button)[3:])
			hostID, err := strconv.ParseInt(hostID_str, 10, 64)

			slog.Debug("Получены аргументы",
				slog.Int64("userID", userID),
				slog.Int64("hostID", hostID))
			if err != nil {
				slog.Error("Ошибка парсинга ID хоста",
					slog.String("error", err.Error()))
				return
			}
			err = b.subscribe(update.CallbackQuery, userID, hostID)
			if err != nil {
				slog.Error("Ошибка подписки",
					slog.String("error", err.Error()))
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID,
					"Произошла ошибка при попытке подписаться на хостера")
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("Ошибка отправки сообщения",
						slog.String("error", err.Error()))
				}
			}
		}
	}
}

func (b *Telegram) handleMap(update *tgbotapi.Update, id int64, mapa string) {
	err := b.addMapName(update.CallbackQuery.Message, mapa)
	if err != nil {
		slog.Error("Ошибка добавления названия карты",
			slog.String("error", err.Error()))
		msg := tgbotapi.NewMessage(id, "Произошла ошибка при выполнении команды")
		msg.ReplyMarkup = list_kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("Ошибка отправки сообщения",
				slog.String("error", err.Error()))
		}
	} else {
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("👨‍🎓 Классика", "classic"),
				tgbotapi.NewInlineKeyboardButtonData("🧌 Прятки", "hide"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🛠️ Моды", "mods"),
				tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "cancel"),
			),
		)
		msg := tgbotapi.NewMessage(id,
			"Выбери режим игры или введи свой вариант "+
				"(не более 10 символов):\n")
		msg.ReplyMarkup = kb
		_, err := b.bot.Send(msg)
		if err != nil {
			slog.Error("Ошибка отправки сообщения",
				slog.String("error", err.Error()))
		}
		b.rep.SaveUserStatus(id, "status", "wait_gamemode")
	}
}

func (b *Telegram) handleMode(update *tgbotapi.Update, id int64, mode string) {
	err := b.addGameMode(update.CallbackQuery.Message, mode)
	if err != nil {
		slog.Error("Ошибка добавления режима игры",
			slog.String("error", err.Error()))
		msg := tgbotapi.NewMessage(id, "Произошла ошибка при выполнении команды")
		msg.ReplyMarkup = list_kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("Ошибка отправки сообщения",
				slog.String("error", err.Error()))
		}
	} else {
		msg := tgbotapi.NewMessage(id, "*Комната успешно добавлена*\n\n"+
			"Для того чтобы не засорять бота неактивными комнатами, "+
			"не забудь её удалить когда закончишь играть")
		msg.ReplyMarkup = list_kb
		msg.ParseMode = "markdownV2"
		_, err := b.bot.Send(msg)
		if err != nil {
			slog.Error("Ошибка отправки сообщения",
				slog.String("error", err.Error()))
		}
		b.rep.SaveUserStatus(id, "status", "null")
	}
}

func (b *Telegram) handleNewMap(update *tgbotapi.Update, id int64, mapa string) {
	err := b.changeMap(update.CallbackQuery.Message, mapa)
	if err != nil {
		slog.Error("Ошибка изменения названия карты при нажатии на кнопку",
			slog.String("error", err.Error()))
		msg := tgbotapi.NewMessage(id, "Произошла ошибка при выполнении команды")
		msg.ReplyMarkup = list_kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("Ошибка отправки сообщения",
				slog.String("error", err.Error()))
		}
	} else {
		msg := tgbotapi.NewMessage(id,
			"Название карты успешно изменено")
		msg.ReplyMarkup = list_kb
		_, err := b.bot.Send(msg)
		if err != nil {
			slog.Error("Ошибка отправки сообщения",
				slog.String("error", err.Error()))
		}
		b.rep.SaveUserStatus(id, "status", "null")
	}
}

func (b *Telegram) handleNewMode(update *tgbotapi.Update, id int64, mode string) {
	err := b.changeDescription(update.CallbackQuery.Message, mode)
	if err != nil {
		slog.Error("Ошибка изменения режима игры при нажатии на кнопку",
			slog.String("error", err.Error()))
		msg := tgbotapi.NewMessage(id, "Произошла ошибка при выполнении команды")
		msg.ReplyMarkup = list_kb
		_, err = b.bot.Send(msg)
		if err != nil {
			slog.Error("Ошибка отправки сообщения",
				slog.String("error", err.Error()))
		}
	} else {
		msg := tgbotapi.NewMessage(id,
			"Режим игры успешно изменён")
		msg.ReplyMarkup = list_kb
		_, err := b.bot.Send(msg)
		if err != nil {
			slog.Error("Ошибка отправки сообщения",
				slog.String("error", err.Error()))
		}
		b.rep.SaveUserStatus(id, "status", "null")
	}
}
