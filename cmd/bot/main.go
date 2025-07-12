package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"telegram-bot-assistente/config"
	"telegram-bot-assistente/internal/handlers"

	"gopkg.in/telebot.v3"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	bot, err := telebot.NewBot(telebot.Settings{
		Token:  cfg.TelegramBotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	log.Printf("Authorized as @%s", bot.Me.Username)

	setupHandlers(bot)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Println("Bot started and ready...")
		bot.Start()
	}()

	waitForShutdown(func() {
		cancel()
		bot.Stop()
	})
	log.Println("Bot stopped")
}

func setupHandlers(bot *telebot.Bot) {
	h := handlers.NewHandlers()
	h.RegisterRoutes(bot)
}

func waitForShutdown(stopFunc func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal %v, starting graceful shutdown...", sig)

	stopFunc()
	time.Sleep(2 * time.Second)
}
