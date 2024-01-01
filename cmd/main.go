package main

import (
	"among-us-roombot/internals/repository"
	"among-us-roombot/internals/service"
	"among-us-roombot/pkg/base/boltdb"
	"among-us-roombot/pkg/logger"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"log/slog"

	"github.com/boltdb/bolt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// load env
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading env: %v", err)
	}

	// setup logger (use slog)
	err := os.MkdirAll("log", os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}

	f, err := os.OpenFile("log/all.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}()

	l := log.Default()
	wrt := io.MultiWriter(os.Stdout, f)
	l.SetOutput(wrt)

	// setup slog
	log, err := setupLogger(os.Getenv("ENV"))
	if err != nil {
		l.Fatal(err)
	}
	slog.SetDefault(log)

	log.Info("start app", slog.String("env", os.Getenv("ENV")))
	log.Debug("debug level is enabled")

	// creare TG bot
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TG_TOKEN"))
	if err != nil {
		l.Fatal(err.Error())
	}
	bot.Debug = false

	// open DB (use bolt)
	db, err := bolt.Open(os.Getenv("DB_FILE"), 0600, nil)
	if err != nil {
		l.Fatal(err.Error())
	}
	defer func() {
		err := db.Close()
		if err != nil {
			l.Fatal(err.Error())
		}
	}()
	base := boltdb.NewBase(db)

	// init
	rep := repository.NewRepository(base)
	service := service.NewService(bot, rep)
	service.Bot.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	err = db.Close()
	if err != nil {
		l.Fatalf("error closing db: %v", err)
	}

}

func setupLogger(env string) (*slog.Logger, error) {
	var log *slog.Logger

	switch env {
	// для локальной разработки и отладки используется дефолтный json handler
	case "local":
		h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: false,
		})
		log = slog.New(h)
	// для продакшена используется кастомный handler, который отправляет логи на сервер
	case "prod":
		h := logger.NewCustomSlogHandler(slog.NewJSONHandler(
			os.Stdout, &slog.HandlerOptions{
				Level:     slog.LevelInfo,
				AddSource: false,
			}))
		log = slog.New(h)
	default:
		return nil, fmt.Errorf("incorrect error level: %s", env)
	}

	return log, nil
}
