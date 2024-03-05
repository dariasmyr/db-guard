package internal

import (
	"context"
	"fmt"
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

	tgBot, err = bot.New(os.Getenv("TELEGRAM_BOT_TOKEN"))

	if err != nil {
		panic(err)
	}

	tgBot.Start(ctx)
}

func SendMessage(chatID int64, text string) error {
	_, err := tgBot.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	return err
}
