package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go_chat/internal/repository"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type BotRunner struct {
	key    string
	bot    *tgbotapi.BotAPI
	RM     *repository.RepositoryManager
	CTX    context.Context
	ApiURL string
}

func (br *BotRunner) SendMessage(text string, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ReplyToMessageID = update.Message.MessageID
	br.bot.Send(msg)
}

func (br *BotRunner) SendMessageWithTgID(text string, id int64) {
	msg := tgbotapi.NewMessage(id, text)
	_, err := br.bot.Send(msg)
	if err != nil {
		return
	}
}

// SendMessageWithID Requires TgChatID to send message to linked internal chat
func (br *BotRunner) SendMessageWithID(text string, chatID int64, username string) {

	url := fmt.Sprintf("%s/send_message_tg", br.ApiURL)

	json, err := json.Marshal(map[string]any{"message": text, "tg_chat_id": chatID, "author": username})
	if err != nil {
		log.Println("SendMessageWithID" + err.Error())
		return
	}

	reader := strings.NewReader(string(json))
	req, err := http.NewRequest("POST", url, reader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	fmt.Println("response Body:", resp.Body)

}

func getKey(path string) string {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return string(file)
}

func (br *BotRunner) ProcessCommand(update tgbotapi.Update) {
	if update.FromChat().IsGroup() {
		symbols := strings.Split(update.Message.Text, " ")
		if len(symbols) > 2 {
			br.SendMessage("Wrong syntax", update)
		} else if len(symbols) > 1 {
			internal := symbols[1]
			id, err := strconv.Atoi(internal)
			if err != nil {
				br.SendMessage("Wrong syntax", update)
			} else {
				err = br.RM.ChatRepository.LinkChat(br.CTX, int(update.FromChat().ID), id)
				if err != nil {
					br.SendMessage("Error occurred", update)
				} else {
					br.SendMessage("Successful", update)
				}
			}
		}
	} else {
		br.SendMessage("You can't use this command here", update)
	}
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
				br.ProcessCommand(update)
			} else {
				// If we got a message and it's not a command
				log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

				go br.SendMessageWithID(update.Message.Text, update.FromChat().ID, update.Message.From.UserName)

			}
		}
	}
}
