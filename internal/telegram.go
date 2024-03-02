package internal

import (
	"context"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
)

func StartBot() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := godotenv.Load()
	opts := []bot.Option{
		bot.WithDefaultHandler(defaultHandler),
	}

	b, err := bot.New(os.Getenv("TELEGRAM_BOT_TOKEN"), opts...)

	if err != nil {
		panic(err)
	}

	b.Start(ctx)
}

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
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
	ctx := context.Background()
	b, err := bot.New(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		return err
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	if err != nil {
		return err
	}

	return nil
}
