package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/telebot.v3"
	"telegram-bot-assistente/config"
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

	ctx, cancel := context.WithCancel(context.Background())
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
	bot.Handle("/start", func(c telebot.Context) error {
		text := "🤖 Добро пожаловать в Task Assistant Bot!\n\n" +
			"Я помогу вам управлять задачами. Доступные команды:\n" +
			"/help - показать все команды\n" +
			"/add - добавить новую задачу\n" +
			"/list - показать список задач\n" +
			"/done - отметить задачу как выполненную"
		
		log.Printf("Message from @%s: %s", c.Sender().Username, c.Text())
		return c.Send(text)
	})

	bot.Handle("/help", func(c telebot.Context) error {
		text := "📋 Доступные команды:\n\n" +
			"/start - начать работу с ботом\n" +
			"/add \"описание\" срок: ГГГГ-ММ-ДД - добавить задачу\n" +
			"/list - показать все активные задачи\n" +
			"/done <id> - отметить задачу как выполненную\n" +
			"/edit <id> новое описание - редактировать задачу\n" +
			"/help - показать эту справку"
		
		log.Printf("Message from @%s: %s", c.Sender().Username, c.Text())
		return c.Send(text)
	})

	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		text := "📝 Я получил ваше сообщение! Пока что я умею только отвечать на команды. Попробуйте /help"
		log.Printf("Message from @%s: %s", c.Sender().Username, c.Text())
		return c.Send(text)
	})

	bot.Handle(telebot.OnCommand, func(c telebot.Context) error {
		text := "❓ Неизвестная команда. Используйте /help для просмотра доступных команд."
		log.Printf("Unknown command from @%s: %s", c.Sender().Username, c.Text())
		return c.Send(text)
	})
}

func waitForShutdown(stopFunc func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal %v, starting graceful shutdown...", sig)
	
	stopFunc()
	time.Sleep(2 * time.Second)
}
