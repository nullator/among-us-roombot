package telegram

import (
	"among-us-roombot/internals/models"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Функция вызывается при получении от пользователя сообщения с учётом его статуса
func (b *Telegram) handleUserStatus(update *tgbotapi.Update, status string) {
	slog.Debug("Получен статус", slog.String("status", status))

	switch status {
	// Получено сообщение с обратной связью пользователя
	case "wait_feedback":

		err := b.feedback(update)
		if err != nil {
			slog.Error("Ошибка обработки обратной связи",
				slog.String("error", err.Error()))
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				"Произошла ошибка при отправке обратной связи")
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}

			err = b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")
			if err != nil {
				slog.Error("Ошибка сохранения в БД данных о статусе пользователя",
					slog.String("error", err.Error()))
			}
		}

	// Получено сообщение с кодом комнаты
	case "edit_code":
		err := b.changeCode(update.Message)
		if err != nil {
			switch err {
			// Обрабатываются ожидаемые ошибки с указанием пользователю причины
			case models.ErrInvalidCode:
				msg_text := "Неверный код комнаты.\n" +
					"Код комнаты должен состоять из 6 латинских букв, " +
					"последняя буква может быть только F, G, Q, f, g или q.\n" +
					"Попробуй ещё раз: /edit"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, msg_text)
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
				}
			case models.ErrRoomAlreadyExist:
				msg_text := "Комната с таким кодом уже существует"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, msg_text)
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
				}
			default:
				msg_text := "Произошла ошибка при изменении кода комнаты"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, msg_text)
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
				}
			}
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				"Код комнаты успешно изменён")
			msg.ReplyMarkup = list_kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}
		}

		// Обнуляем статус пользователя
		b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")

	// Получено сообщение с названием карты
	case "change_map":
		mapa := update.Message.Text
		err := b.changeMap(update.Message, mapa)
		if err != nil {
			// Обрабатываются ожидаемые ошибки с указанием пользователю причины
			if err == models.ErrInvalidMap {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Слишком длинное название карты.\n"+
						"Название должно состоять не более чем из 10 символов.\n"+
						"Чтобы попробовать ещё раз введи команду /edit")
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
				}
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Произошла ошибка при изменении названия карты")
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
				}
			}
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				"Название карты успешно изменено")
			msg.ReplyMarkup = list_kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}
		}

		// Обнуляем статус пользователя
		b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")

	// Получено сообщение с новым ником хостера
	case "change_hoster":
		err := b.changeHoster(update.Message)
		if err != nil {
			// Обрабатываются ожидаемые ошибки с указанием пользователю причины
			if err == models.ErrInvalidName {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Слишком длинный ник.\n"+
						"Ник должен быть не длинее 10 символов.\n"+
						"Чтобы попробовать ещё раз введи команду /edit")
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
				}
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Произошла ошибка при изменении ника хостера")
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
				}
			}
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				"Ник хостера успешно изменён")
			msg.ReplyMarkup = list_kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}
		}

		// Обнуляем статус пользователя
		b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")

	// Получено сообщение с новым режимом игры
	case "change_description":
		mode := update.Message.Text
		err := b.changeDescription(update.Message, mode)
		if err != nil {
			// Обрабатываются ожидаемые ошибки с указанием пользователю причины
			if err == models.ErrInvalidName {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Слишком длинное название режима.\n"+
						"Название должно быть не длинее 10 символов.\n"+
						"Чтобы попробовать ещё раз введи команду /edit")
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
				}
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Произошла ошибка при изменении названия режима")
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
				}
			}
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				"Описание режима успешно изменено")
			msg.ReplyMarkup = list_kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}
		}

		// Обнуляем статус пользователя
		b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")

	// Получено сообщение кодом создаваемой комнаты
	case "start_add_room":
		err := b.addDraftRoom(update.Message)
		if err != nil {
			switch err {
			// Обрабатываются ожидаемые ошибки с указанием пользователю причины
			case models.ErrInvalidCode:
				msg_text := "Неверный код комнаты.\n" +
					"Код комнаты должен состоять из 6 латинских букв, " +
					"последняя буква может быть только F, G, Q, f, g или q\n" +
					"Попробуй ещё раз: /add"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, msg_text)
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
				}
			case models.ErrRoomAlreadyExist:
				msg_text := "Комната с таким кодом уже существует.\n" +
					"Попробуй ещё раз: /add"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, msg_text)
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
				}
			default:
				msg_text := "Произошла ошибка при создании комнаты" +
					"Попробуй ещё раз: /add"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, msg_text)
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
				}
			}
			// Обнуляем статус пользователя
			b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")
		} else {
			// Если черновик комнаты успешно создан, то следующим шагом является ввод ника хостера
			// Проверяем, есть ли у пользователя старый ник хостера
			old_host_name, err := b.rep.GetUserStatus(update.Message.Chat.ID, "host_name")
			if err != nil {
				slog.Error("Ошибка чтения из БД данных о старом нике хостера",
					slog.String("error", err.Error()))
				old_host_name = ""
			}

			// Если старый ник есть, то предлагаем пользователю создать комнату со старым ником
			if old_host_name != "" {
				msg_text := fmt.Sprintf("Привет, %s!\nЧтобы создать руму со своим "+
					"предыдущим ником, нажми на соответствующую кнопку, "+
					"или придумай новый ник и отправь мне его в чат", old_host_name)
				kb := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData(old_host_name, "save_old_name"),
						tgbotapi.NewInlineKeyboardButtonData("Отменить", "cancel"),
					),
				)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, msg_text)
				msg.ReplyMarkup = kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("Ошибка отправки сообщения",
						slog.String("error", err.Error()))
				}

				// Если старого ника нет, то предлагаем пользователю придумать новый ник
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Успешно создан черновик комнаты\n"+
						"Введи ник хостера (не более 10 символов):\n")
				msg.ReplyMarkup = cancel_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("Ошибка отправки сообщения",
						slog.String("error", err.Error()))
				}
			}

			// Сохраняем статус пользователя на "wait_hostname"
			b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "wait_hostname")
		}

	// Получено сообщение с ником хостера
	case "wait_hostname":
		name := update.Message.Text
		err := b.addHostName(update.Message, name)
		if err != nil {
			// Обрабатываются ожидаемые ошибки с указанием пользователю причины
			if err == models.ErrInvalidName {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Слишком длинный ник.\n"+
						"Ник должен быть не более 10 символов.\n"+
						"Попробуй ещё раз:\n")
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
				}

				// TODO: Проверить, нужно ли обнулять статус пользователя, т.к. он
				//  не меняется и уже был установлен на "wait_hostname"
				b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "wait_hostname")
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Произошла ошибка при сохранении ника в черновик комнаты")
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
				}
				b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")
			}

		} else {
			// Если ник пользователя получен и сохранен в черновик комнаты, то следующим шагом
			// является ввод названия карты. Создается клавиатура с названиями карт
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
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				"Выбери название карты или введи свой вариант "+
					"(не более 10 символов):\n")
			msg.ReplyMarkup = kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}

			// Сохраняем статус пользователя на "wait_mapname"
			b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "wait_mapname")
		}

	// Получено сообщение с названием карты
	case "wait_mapname":
		mapa := update.Message.Text
		err := b.addMapName(update.Message, mapa)
		if err != nil {
			// Обрабатываются ожидаемые ошибки с указанием пользователю причины
			if err == models.ErrInvalidMap {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Слишком длинное название карты\n"+
						"Название карты должно быть не более 10 символов.\n"+
						"Попробуй ещё раз:\n")
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
				}

				// TODO: Проверить, нужно ли обнулять статус пользователя
				b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "wait_mapname")
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Произошла ошибка при сохранении названия карты в черновик комнаты")
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
				}
				b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")
			}

		} else {
			// Если название карты пользователя получено и сохранено в черновик комнаты,
			// то следующим шагом является ввод режима игры. Создается клавиатура с режимами игры
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
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				"Выбери режим игры или введи свой вариант "+
					"(не более 10 символов):\n")
			msg.ReplyMarkup = kb
			msg.ReplyMarkup = cancel_kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}

			// Сохраняем статус пользователя на "wait_gamemode"
			b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "wait_gamemode")
		}

	// Получено сообщение с режимом игры
	case "wait_gamemode":
		mode := update.Message.Text
		err := b.addGameMode(update.Message, mode)
		if err != nil {
			// Обрабатываются ожидаемые ошибки с указанием пользователю причины
			if err == models.ErrInvalidMap {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Слишком длинное название режима\n"+
						"Название режима должно быть не более 10 символов.\n"+
						"Попробуй ещё раз:\n")
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
				}

				// TODO: Проверить, нужно ли обнулять статус пользователя
				b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "wait_gamemode")
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Произошла ошибка при сохранении режима игры в черновик комнаты")
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("error send message to user")
				}
				b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")
			}

		} else {
			// Выбор режима игры был последним шагом. Выводится сообщение об успешном
			// создании комнаты с напоминанием о необходимости удаления неактуальных рум
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "*Комната успешно добавлена*\n\n"+
				"Для того чтобы не засорять бота неактивными комнатами, "+
				"не забудь её удалить когда закончишь играть")
			msg.ReplyMarkup = list_kb
			msg.ParseMode = "markdownV2"
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}
			b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")
		}

	// Получено сообщение, которое хостер хочет отправить своим подписчикам
	case "wait_post":
		post := update.Message.Text
		// Проверка чтобы текст рассылки не был пустым
		if post == "" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				"Текст рассылки не может быть пустым, попробуй ещё раз командой /notify")
			msg.ReplyMarkup = list_kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}
			b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")
			break
		}

		// Проверка чтобы текст рассылки не был слишком длинным
		if len(post) > 1000 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				"Пожалей своих подписчиков и не отправляй им текст Войны и мира, "+
					"попробуй ещё раз с сообщением покороче /notify")
			msg.ReplyMarkup = list_kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}
			b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")
			break
		}

		// Запускается рассылка сообщения хостера в адрес его подписчиков
		err := b.sendPost(update.Message, post)
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				"Произошла ошибка при отправке рассылки")
			msg.ReplyMarkup = list_kb
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка отправки сообщения",
					slog.String("error", err.Error()))
			}
		}
		b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")

	default:
		cmd := string([]rune(status)[0:2])
		slog.Debug("Получена команда", slog.String("cmd", cmd))
		switch cmd {
		case "sub":
			userID := update.Message.Chat.ID
			hostID_str := string([]rune(status)[4:])
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
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Произошла ошибка при попытке подписаться на хостера")
				msg.ReplyMarkup = list_kb
				_, err := b.bot.Send(msg)
				if err != nil {
					slog.Error("Ошибка отправки сообщения",
						slog.String("error", err.Error()))
				}
			}

		case "uns":
			//
		default:
			slog.Warn("Получен неизвестный статус пользователя",
				slog.String("user", update.Message.From.String()),
				slog.Int64("id", update.Message.Chat.ID),
				slog.String("status", status))
		}
	}
}

// Функция перенаправляет полученное от пользователя сообщение разработчику
func (b *Telegram) feedback(update *tgbotapi.Update) error {
	// TelegramId разработчика
	admin_id, err := strconv.ParseInt(os.Getenv("TG_adminID"), 10, 64)
	if err != nil {
		slog.Error("при выполнении авторизации не удалось распарсить ID в TelegramId",
			slog.String("error", err.Error()))
		return err
	}

	// Формируется новое сообщение для разработчика с текстом сообщения пользователя
	msg_text := fmt.Sprintf("Получена обратная связь от %s содержания: %s",
		update.Message.From.String(), update.Message.Text)
	msg := tgbotapi.NewMessage(admin_id, msg_text)
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("Не удалось отправить обратную связь",
			slog.String("error", err.Error()))
		return err
	}

	// Одновременно сообщение пользователя просто перенаправляется разработчику
	forvard_msg := tgbotapi.NewForward(admin_id,
		update.Message.Chat.ID,
		update.Message.MessageID)
	_, err = b.bot.Send(forvard_msg)
	if err != nil {
		slog.Error("Не удалось переслать сообщение",
			slog.String("error", err.Error()))
		return err
	}

	// Отправляется сообщение пользователю о том, что его сообщение успешно отправлено
	msg_text = "Спасибо, сообщение отправлено разработчику! " +
		"При необходимости можно повторно ввести команду /feedback " +
		"и отправить ещё одно сообщение"
	msg = tgbotapi.NewMessage(update.Message.Chat.ID, msg_text)
	msg.ReplyMarkup = list_kb
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("Не удалось отправить обратную связь",
			slog.String("error", err.Error()))
		return err
	}

	// Обнуляется статус пользователя
	err = b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")
	if err != nil {
		slog.Error("Ошибка сохранения в БД данных о статусе пользователя",
			slog.String("error", err.Error()))
		return err
	}
	slog.Debug("Успешно изменён статус в БД")

	return nil
}
