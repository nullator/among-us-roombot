package telegram

import (
	"fmt"
	"log/slog"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Функция выполняется если в handleUpdates зафиксировано нажатие на кнопку
// button - название кнопки
func (b *Telegram) handleButton(update *tgbotapi.Update, button string, id int64) {
	switch button {
	// Обработка кнопки удаления комнаты
	case "delete":
		err := b.delete(update.CallbackQuery.Message)
		if err != nil {
			slog.Error("Ошибка удаления комнаты",
				slog.String("error", err.Error()))
		}

	// Обработка кнопки продления срока жизни комнаты
	case "add_time":
		err := b.addTime(update.CallbackQuery.Message)
		if err != nil {
			slog.Error("Ошибка добавления времени",
				slog.String("error", err.Error()))
		}

	// Обработка кнопки при создании комнаты в случае если используется ранее сохраненный ник
	case "save_old_name":
		slog.Info("Зафиксировано нажатие на кнопку использования старого ника хостера",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Ранее используемый ник хранится в БД в статусе host_name
		old_name, err := b.rep.GetUserStatus(id, "host_name")
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

		// Вызов функции добавления ника хостера в черновик комнаты, выполнение кода
		// завершается в функции addHostName в файле add.go
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
			// Если функция addHostName выполнилась без ошибок, то следующим шагом
			// пользователю предлагается выбрать карту, для чего создается клавиатура
			// и обновляется статус на ожидание названия карты
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

			// Сохраняем статус пользователя на "wait_mapname"
			b.rep.SaveUserStatus(id, "status", "wait_mapname")

		}

	// Обработка кнопки с названием карты при создании комнаты
	case "skeld":
		slog.Info("Зафиксировано нажатие на кнопку выбора карты Skeld",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Запускается функция обработки названия карты, в функцию передается название карты
		b.handleMap(update, id, "Skeld")

	// Обработка кнопки с названием карты при создании комнаты
	case "polus":
		slog.Info("Зафиксировано нажатие на кнопку выбора карты Polus",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Запускается функция обработки названия карты, в функцию передается название карты
		b.handleMap(update, id, "Polus")

	// Обработка кнопки с названием карты при создании комнаты
	case "fungle":
		slog.Info("Зафиксировано нажатие на кнопку выбора карты Fungle",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Запускается функция обработки названия карты, в функцию передается название карты
		b.handleMap(update, id, "Fungle")

	// Обработка кнопки с названием карты при создании комнаты
	case "airship":
		slog.Info("Зафиксировано нажатие на кнопку выбора карты Airship",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Запускается функция обработки названия карты, в функцию передается название карты
		b.handleMap(update, id, "Airship")

	// Обработка кнопки с названием карты при создании комнаты
	case "mira":
		slog.Info("Зафиксировано нажатие на кнопку выбора карты Mira HQ",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Запускается функция обработки названия карты, в функцию передается название карты
		b.handleMap(update, id, "Mira HQ")

	// Обработка кнопка с названием режима игры при создании комнаты
	case "classic":
		slog.Info("Зафиксировано нажатие на кнопку выбора режима Classic",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Запускается функция обработки названия режима, в функцию передается название режима
		b.handleMode(update, id, "Классика")

	// Обработка кнопка с названием режима игры при создании комнаты
	case "hide":
		slog.Info("Зафиксировано нажатие на кнопку выбора режима Hide and Seek",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Запускается функция обработки названия режима, в функцию передается название режима
		b.handleMode(update, id, "Прятки")

	// Обработка кнопка с названием режима игры при создании комнаты
	case "mods":
		slog.Info("Зафиксировано нажатие на кнопку выбора режима Mods",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Запускается функция обработки названия режима, в функцию передается название режима
		b.handleMode(update, id, "Моды")

	// Обработка кнопки изменения кода комнаты
	case "change_code":
		slog.Info("Зафиксировано нажатие на кнопку изменения кода комнаты",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Сохраняем статус пользователя на "change_code", в следующем сообщении
		// от пользователя ожидается новый код комнаты
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

	// Обработка кнопки изменения названия карты
	case "change_map":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия карты",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Сохраняем статус пользователя на "change_map", в следующем сообщении
		// от пользователя ожидается новое название карты
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

		// Выводится клавиатура со стандартным названием карт
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

	// Обработка кнопки изменения навания карты
	case "change_skeld":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия карты на Skeld",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Запускается функция обработки названия карты, в функцию передается название карты
		b.handleNewMap(update, id, "Skeld")

	// Обработка кнопки изменения навания карты
	case "change_polus":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия карты на Polus",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Запускается функция обработки названия карты, в функцию передается название карты
		b.handleNewMap(update, id, "Polus")

	// Обработка кнопки изменения навания карты
	case "change_fungle":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия карты на Fungle",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Запускается функция обработки названия карты, в функцию передается название карты
		b.handleNewMap(update, id, "Fungle")

	// Обработка кнопки изменения навания карты
	case "change_airship":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия карты на Airship",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Запускается функция обработки названия карты, в функцию передается название карты
		b.handleNewMap(update, id, "Airship")

	// Обработка кнопки изменения навания карты
	case "change_mira":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия карты на Mira",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Запускается функция обработки названия карты, в функцию передается название карты
		b.handleNewMap(update, id, "Mira HQ")

	// Обработка кнопки изменения ника хостера
	case "change_hoster":
		slog.Info("Зафиксировано нажатие на кнопку изменения ника хостера",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Сохраняем статус пользователя на "change_hoster", в следующем сообщении
		// от пользователя ожидается новый ник хостера
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

	// Обработка кнопки изменения режима игры
	case "change_description":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия режима",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Сохраняем статус пользователя на "change_description", в следующем сообщении
		// от пользователя ожидается новое название режима
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

		// Выводится клавиатура со стандартными названиями режимов
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

	// Обработка кнопки изменения режима игры
	case "change_classic":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия режима на Classic",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Запускается функция обработки названия режима, в функцию передается название режима
		b.handleNewMode(update, id, "Классика")

	// Обработка кнопки изменения режима игры
	case "change_hide":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия режима на Hide and Seek",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Запускается функция обработки названия режима, в функцию передается название режима
		b.handleNewMode(update, id, "Прятки")

	// Обработка кнопки изменения режима игры
	case "change_mods":
		slog.Info("Зафиксировано нажатие на кнопку изменения названия режима на Mods",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Запускается функция обработки названия режима, в функцию передается название режима
		b.handleNewMode(update, id, "Моды")

	// Обработка кнопки отмены команды
	// Кнопка сбрасывает статус пользователя на "null" (т.е. прекращается выполнение команд),
	// выводится клавиатура с командой /list (т.е. удаляются другие клавиатуры)
	// TODO: подумать над удалением черновика комнаты из БД и других временных данных
	case "cancel":
		slog.Info("Зафиксировано нажатие на кнопку отмены команды",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Сбрасываем статус пользователя на "null"
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

	// Обработка кнопки вывода списка комнат
	case "roomlist":
		slog.Info("Зафиксировано нажатие на кнопку вывода списка комнат",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID))

		// Вызов функции вывода списка комнат
		err := b.handleList(update.CallbackQuery.Message)
		if err != nil {
			slog.Error("Ошибка вывода списка комнат",
				slog.String("error", err.Error()))
		}

	// Обработка кнопки рассылки типового сообщения от хостера в адрес его подписчиков
	case "send_template":
		// Для этого берётся код комнаты хостера из БД
		room_code, err := b.rep.GetUserStatus(id, "room")

		if err != nil {
			slog.Error("Ошибка чтения из БД данных о созданной пользователем комнате",
				slog.String("error", err.Error()))
			msg := tgbotapi.NewMessage(id, "Произошла ошибка при выполнении команды")
			msg.ReplyMarkup = list_kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}
			err = b.rep.SaveUserStatus(id, "status", "null")
			if err != nil {
				slog.Error("Ошибка сохранения в БД данных о статусе пользователя",
					slog.String("error", err.Error()))
			}
			break
		}

		// По коду комнаты получаем данные о комнате из БД
		room, err := b.rep.GetRoom(room_code)
		if err != nil {
			slog.Error("Ошибка чтения из БД данных о созданной пользователем комнате",
				slog.String("error", err.Error()))
			msg := tgbotapi.NewMessage(id, "Произошла ошибка при выполнении команды")
			msg.ReplyMarkup = list_kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}
			err = b.rep.SaveUserStatus(id, "status", "null")
			if err != nil {
				slog.Error("Ошибка сохранения в БД данных о статусе пользователя",
					slog.String("error", err.Error()))
			}
			break
		}

		// Формируем типовое сообщение для рассылки
		// TODO: подумать над тем, чтобы вывести формат этого и других сообщений в отдельный
		// файл, чтобы можно было его менять без перекомпиляции и (или) без правок
		// в разных кусках кода
		post := fmt.Sprintf("_Привет!\nЗаходи ко мне поиграть, "+
			"я играю на карте %s, режим %s, код:_\n\n`%s`", room.Map, room.Mode, room.Code)

		// Вызов функции рассылки сообщения
		err = b.sendPost(update.CallbackQuery.Message, post)
		if err != nil {
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID,
				"Произошла ошибка при отправке рассылки")
			msg.ReplyMarkup = list_kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}
		}
		b.rep.SaveUserStatus(update.CallbackQuery.Message.Chat.ID, "status", "null")

	// В иных случаях нажата кнопка с параметрами, которы закодированы в data кнопки
	default:
		// Первых 3 символа кодируют обрабатываему команду, остальные - параметры
		cmd := string([]rune(button)[0:3])
		slog.Info("Нажатие на кнопку с параметрами",
			slog.String("user", update.CallbackQuery.Message.Chat.UserName),
			slog.Int64("id", update.CallbackQuery.Message.Chat.ID),
			slog.String("cmd", cmd))
		switch cmd {
		// Команда подписки на хостера, параметр - ID хостера
		case "sub":
			userID := update.CallbackQuery.Message.Chat.ID
			hostID_str := string([]rune(button)[3:])
			hostID, err := strconv.ParseInt(hostID_str, 10, 64)
			if err != nil {
				slog.Error("Ошибка парсинга ID хоста",
					slog.String("error", err.Error()))
				return
			}

			// Вызов функции подписки на хостера
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

		// Are you shure? - подтверждение отписки от хостера
		case "ays":
			userID := update.CallbackQuery.Message.Chat.ID
			hostID_str := string([]rune(button)[3:])
			hostID, err := strconv.ParseInt(hostID_str, 10, 64)
			if err != nil {
				slog.Error("Ошибка парсинга ID хоста",
					slog.String("error", err.Error()))
				return
			}

			// Выполняется функция подтверждения отписки от хостера, которая
			// отправит пользователю кнопки с подтверждением или отменой отписки
			err = b.areYouShure(userID, hostID)
			if err != nil {
				slog.Error("Ошибка подтверждения удаления комнаты",
					slog.String("error", err.Error()))
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID,
					"Произошла ошибка при попытке отписаться от хостера")
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("Ошибка отправки сообщения",
						slog.String("error", err.Error()))
				}
			}

		// Отмена отписки от хостера
		case "uns":
			userID := update.CallbackQuery.Message.Chat.ID
			hostID_str := string([]rune(button)[3:])
			hostID, err := strconv.ParseInt(hostID_str, 10, 64)
			if err != nil {
				slog.Error("Ошибка парсинга ID хоста",
					slog.String("error", err.Error()))
				return
			}

			// Выполняется функция отписки от хостера
			// Учитывая что в параметрах команды указан ID хостера, то для вывода
			// сообщения об успешной отписке от хостера с указанием ника хостера
			// команда вернет ник хостера, или сообщение об ошибке которое будет
			// выведено пользователю
			txt, err := b.unsubscribe(userID, hostID)
			if err != nil {
				slog.Error("Ошибка отписки",
					slog.String("error", err.Error()))
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, txt)
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("Ошибка отправки сообщения",
						slog.String("error", err.Error()))
				}
			} else {
				msg_text := fmt.Sprintf("Успешно выполнена отписка от %s\n\n"+
					"Если хочешь подписаться на других хостов жми /subscribe\n"+
					"Для отписки жми /unsubscribe", txt)
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, msg_text)
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

// Обрабатывает полученное название карты и вызывает функцию добавления названия карт
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
		// Если название карты успешно добавлено, то следующим шагом является
		// добавление режима игры, поэтому выводится клавиатура с режимами игры
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

		// Сохраняем статус пользователя на "wait_gamemode"
		b.rep.SaveUserStatus(id, "status", "wait_gamemode")
	}
}

// Обрабатывает полученное название режима игры и вызывает функцию добавления названия режима игры
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
		// Добавление режима было последним шагом, поэтосу выводится напоминание
		// и необходимости удаления неактивных комнат, статус пользователя обнуляется
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

// Обрабатывает новое название карты и вызывает функцию изменения карты
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

// Обрабатывает новое название режима игры и вызывает функцию изменения режима игры
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
