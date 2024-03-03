package internal

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot/models"
	"log"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
)

var tgBot *bot.Bot

func StartBot() {
	fmt.Println("Starting bot")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var err error

	opts := []bot.Option{
		bot.WithDefaultHandler(defaultHandler),
	}

	tgBot, err = bot.New(os.Getenv("TELEGRAM_BOT_TOKEN"), opts...)

	if err != nil {
		panic(err)
	}

	tgBot.Start(ctx)
}

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Printf("Received message: %s, chatID: %d\n", update.Message.Text, update.Message.Chat.ID)
	SendMessage(update.Message.Chat.ID, "Received message: "+update.Message.Text)
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Bot started",
	})
	if err != nil {
		log.Printf("Enable to send default message")
		return
	}
}

func SendMessage(chatID int64, text string) error {
	_, err := tgBot.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	return err
}
