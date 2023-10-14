package telegram

import (
	"among-us-roombot/internals/models"
	"fmt"
	"os"
	"strconv"

	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var list_kb = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Список рум", "roomlist"),
	),
)

func (b *Telegram) handleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		switch {
		// Processing chat messages
		case update.Message != nil:
			sender := fmt.Sprintf("%s (%s)",
				update.Message.From.UserName,
				update.Message.From.String())

			// Checking for command input
			if update.Message.IsCommand() {
				slog.Info("Зафиксирована команда",
					slog.String("cmd", update.Message.Text),
					slog.String("user", sender),
					slog.Int64("id", update.Message.Chat.ID),
				)

				if err := b.handleCommand(update.Message); err != nil {
					slog.Error("При обработке команды произошла ошибка",
						slog.String("cmd", update.Message.Command()),
						slog.String("error", err.Error()),
						slog.String("user", sender),
						slog.Int64("id", update.Message.Chat.ID),
					)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID,
						"Произошла неожиданная ошибка при обработке команды")
					_, err := b.bot.Send(msg)
					if err != nil {
						slog.Error("error send message to user")
					}

				}
				continue
			}

			slog.Info("Получено сообщение",
				slog.String("user", sender),
				slog.String("message", update.Message.Text),
				slog.Int64("id", update.Message.Chat.ID))

			// Checking to receive feedback
			user_status, err := b.rep.GetUserStatus(update.Message.Chat.ID, "status")
			if err != nil {
				slog.Error("Ошибка чтения из БД данных о статусе пользователя",
					slog.String("error", err.Error()))
			}
			switch user_status {
			case "wait_feedback":
				err := b.feedback(&update)
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
				}

			case "edit_code":
				err := b.changeCode(update.Message)
				if err != nil {
					switch err {
					case models.ErrInvalidCode:
						msg_text := "Неверный код комнаты.\n" +
							"Код комнаты должен состоять из 6 латинских букв.\n" +
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
				b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")

			case "change_map":
				err := b.changeMap(update.Message)
				if err != nil {
					if err == models.ErrInvalidMap {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID,
							"Слишком длинное название карты.\n"+
								"Название карты должно состоять не более чем из 10 символов.\n"+
								"Попробуй ещё раз: /edit")
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
				b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")

			case "change_hoster":
				err := b.changeHoster(update.Message)
				if err != nil {
					if err == models.ErrInvalidName {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID,
							"Слишком длинный ник.\n"+
								"Ник должен быть не длинее 10 символов.\n"+
								"Попробуй ещё раз: /edit")
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

				b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")

			case "change_description":
				err := b.changeDescription(update.Message)
				if err != nil {
					if err == models.ErrInvalidName {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID,
							"Слишком длинное название режима.\n"+
								"Название должно быть не длинее 10 символов.\n"+
								"Попробуй ещё раз: /edit")
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
				b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")

			case "start_add_room":
				err := b.addDraftRoom(update.Message)
				if err != nil {
					switch err {
					case models.ErrInvalidCode:
						msg_text := "Неверный код комнаты.\n" +
							"Код комнаты должен состоять из 6 латинских букв.\n" +
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
						msg_text := "Произошла ошибка при изменении кода комнаты"
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, msg_text)
						msg.ReplyMarkup = list_kb
						_, err := b.bot.Send(msg)
						if err != nil {
							slog.Error("error send message to user")
						}
					}
					b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")
				} else {
					var cancel_kb = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Отменить", "cancel"),
						),
					)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID,
						"Успешно создан черновик комнаты\n"+
							"Введи ник хостера (не более 10 символов):\n")
					msg.ReplyMarkup = cancel_kb
					_, err := b.bot.Send(msg)
					if err != nil {
						slog.Error("Ошибка отправки сообщения",
							slog.String("error", err.Error()))
					}

					b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "wait_hostname")
				}

			case "wait_hostname":
				err := b.addHostName(update.Message)
				if err != nil {
					if err == models.ErrInvalidName {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID,
							"Слишком длинный ник.\n"+
								"Ник должен быть не более 10 символов.\n"+
								"Попробуй ещё раз:\n")
						_, err := b.bot.Send(msg)
						if err != nil {
							slog.Error("error send message to user")
						}
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
					var cancel_kb = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Отменить", "cancel"),
						),
					)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID,
						"Ник хостера успешно получен и сохранён в черновик комнаты.\n"+
							"Введи название карты (не более 10 символов):\n")
					msg.ReplyMarkup = cancel_kb
					_, err := b.bot.Send(msg)
					if err != nil {
						slog.Error("Ошибка отправки сообщения",
							slog.String("error", err.Error()))
					}
					b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "wait_mapname")
				}

			case "wait_mapname":
				err := b.addMapName(update.Message)
				if err != nil {
					if err == models.ErrInvalidMap {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID,
							"Слишком длинное название карты\n"+
								"Название карты должно быть не более 10 символов.\n"+
								"Попробуй ещё раз:\n")
						_, err := b.bot.Send(msg)
						if err != nil {
							slog.Error("error send message to user")
						}
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
					var cancel_kb = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Отменить", "cancel"),
						),
					)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID,
						"Название карты успешно сохранено в черновик комнаты.\n"+
							"Введи режим игры (не более 10 символов):\n")
					msg.ReplyMarkup = cancel_kb
					_, err := b.bot.Send(msg)
					if err != nil {
						slog.Error("Ошибка отправки сообщения",
							slog.String("error", err.Error()))
					}
					b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "wait_gamemode")
				}

			case "wait_gamemode":
				err := b.addGameMode(update.Message)
				if err != nil {
					if err == models.ErrInvalidMap {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID,
							"Слишком длинное название режима\n"+
								"Название режима должно быть не более 10 символов.\n"+
								"Попробуй ещё раз:\n")
						_, err := b.bot.Send(msg)
						if err != nil {
							slog.Error("error send message to user")
						}
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
					msg := tgbotapi.NewMessage(update.Message.Chat.ID,
						"Комната успешно создана")
					msg.ReplyMarkup = list_kb
					_, err := b.bot.Send(msg)
					if err != nil {
						slog.Error("Ошибка отправки сообщения",
							slog.String("error", err.Error()))
					}
					b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")
				}

			default:

			}

		// Обработка нажатий на кнопки
		case update.CallbackQuery != nil:
			q := update.CallbackQuery.Data
			id := update.CallbackQuery.Message.Chat.ID
			switch q {
			case "delete":
				err := b.delete(update.CallbackQuery.Message)
				if err != nil {
					slog.Error("Ошибка удаления комнаты",
						slog.String("error", err.Error()))
				}

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
				var cancel_kb = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Отменить", "cancel"),
					),
				)
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
				var cancel_kb = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Отменить", "cancel"),
					),
				)
				msg := tgbotapi.NewMessage(id, "Отправь мне новое название карты:")
				msg.ReplyMarkup = cancel_kb
				_, err = b.bot.Send(msg)
				if err != nil {
					slog.Error("Ошибка отправки сообщения",
						slog.String("error", err.Error()))
				}

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
				var cancel_kb = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Отменить", "cancel"),
					),
				)
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
				var cancel_kb = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Отменить", "cancel"),
					),
				)
				msg := tgbotapi.NewMessage(id, "Отправь мне новое название режима:")
				msg.ReplyMarkup = cancel_kb
				_, err = b.bot.Send(msg)
				if err != nil {
					slog.Error("Ошибка отправки сообщения",
						slog.String("error", err.Error()))
				}

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

			}

			msg := tgbotapi.NewEditMessageReplyMarkup(
				id, update.CallbackQuery.Message.MessageID,
				tgbotapi.InlineKeyboardMarkup{
					InlineKeyboard: make([][]tgbotapi.InlineKeyboardButton, 0)})
			_, err := b.bot.Send(msg)
			if err != nil {
				slog.Error("Ошибка удаления клавиатуры",
					slog.String("error", err.Error()))
			}

		}
	}
}

func (b *Telegram) feedback(update *tgbotapi.Update) error {
	admin_id, err := strconv.ParseInt(os.Getenv("TG_adminID"), 10, 64)
	if err != nil {
		slog.Error("при выполнении авторизации не удалось распарсить ID в TelegramId",
			slog.String("error", err.Error()))
		return err
	}

	msg_text := fmt.Sprintf("Получена обратная связь от %s содержания: %s",
		update.Message.From.String(), update.Message.Text)
	msg := tgbotapi.NewMessage(admin_id, msg_text)
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("Не удалось отправить обратную связь",
			slog.String("error", err.Error()))
		return err
	}

	forvard_msg := tgbotapi.NewForward(admin_id,
		update.Message.Chat.ID,
		update.Message.MessageID)
	_, err = b.bot.Send(forvard_msg)
	if err != nil {
		slog.Error("Не удалось переслать сообщение",
			slog.String("error", err.Error()))
		return err
	}

	msg_text = "Спасибо, сообщение отправлено разработчику! " +
		"При необходимости можно повторно ввести команду /feedback " +
		"и отправить ещё одно сообщение, в том числе можно отправить файлы, скриншоты и т.п."
	msg.ReplyMarkup = list_kb
	msg = tgbotapi.NewMessage(update.Message.Chat.ID, msg_text)
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("Не удалось отправить обратную связь",
			slog.String("error", err.Error()))
		return err
	}

	err = b.rep.SaveUserStatus(update.Message.Chat.ID, "status", "null")
	if err != nil {
		slog.Error("Ошибка сохранения в БД данных о статусе пользователя",
			slog.String("error", err.Error()))
		return err
	}
	slog.Debug("Успешно изменён статус в БД")

	return nil
}
