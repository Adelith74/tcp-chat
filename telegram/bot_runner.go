package telegram

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go_chat/internal/repository"
	"log"
	"os"
	"strconv"
	"strings"
)

type BotRunner struct {
	key string
	bot *tgbotapi.BotAPI
	RM  *repository.RepositoryManager
	CTX context.Context
}

func getKey(path string) string {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return string(file)
}

func (br *BotRunner) Run() {
	key := getKey("./telegram/api_key.txt")
	br.key = key
	bot, err := tgbotapi.NewBotAPI(key)
	if err != nil {
		log.Panic(err)
	}
	br.bot = bot

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)

	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if strings.Contains(update.Message.Text, "/link") {
				if update.FromChat().IsGroup() {
					symbols := strings.Split(update.Message.Text, " ")
					if len(symbols) > 2 {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Wrong syntax")
						msg.ReplyToMessageID = update.Message.MessageID
						bot.Send(msg)
					} else if len(symbols) > 1 {
						internal := symbols[1]
						id, err := strconv.Atoi(internal)
						if err != nil {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Wrong syntax")
							msg.ReplyToMessageID = update.Message.MessageID
							bot.Send(msg)
						} else {
							err = br.RM.ChatRepository.LinkChat(br.CTX, int(update.FromChat().ID), id)
							if err != nil {
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error occurred")
								msg.ReplyToMessageID = update.Message.MessageID
								bot.Send(msg)
							} else {
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Successful")
								msg.ReplyToMessageID = update.Message.MessageID
								bot.Send(msg)
							}
						}
					}
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You can't use this command here")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
				}
			} else {
				// If we got a message
				log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				msg.ReplyToMessageID = update.Message.MessageID

				bot.Send(msg)
			}
		}
	}
}
