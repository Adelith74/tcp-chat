package api

import (
	"context"
	"fmt"
	"go_chat/internal/chat"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"go_chat/internal/repository"

	"github.com/gin-gonic/gin"
)

type body struct {
	Tg_chat_id int64  `json:"tg_chat_id"`
	Message    string `json:"message"`
	Author     string `json:"author"`
}

func Start_api(ctx context.Context, server *chat.Server, rManager *repository.RepositoryManager) {
	defer server.Wg.Done()
	router := gin.Default()
	//syntax emaple:
	//http://localhost:8080/disconnect/?chat_id=1&username=lol
	router.POST("/disconnect", func(c *gin.Context) {
		username := c.Query("username")
		err := server.KickUser(username)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		} else {
			c.JSON(http.StatusAccepted, gin.H{"message": "Disconnected"})
		}
	})

	router.POST("/auth", func(c *gin.Context) {

	})

	router.POST("/shutdown", func(c *gin.Context) {
		server.Wg.Done()
	})

	router.POST("/send_message_tg", func(c *gin.Context) {
		var b body
		err := c.BindJSON(&b)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
			return
		}
		if "" == strings.TrimSpace(b.Author) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "No author"})
			return
		}
		chat, err := rManager.ChatRepository.GetIdWithTgID(ctx, b.Tg_chat_id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err})
			return
		}
		for ch, users := range server.Chats {
			if ch.Chat_id == chat.Chat_id {
				msg := b.Message
				for _, user := range users {
					user.Write(fmt.Sprintf("[" + b.Author + " from Telegram]" + ": " + msg + "\n" + "\r"))
				}
				break
			}
			c.JSON(http.StatusBadRequest, gin.H{"message": "Chat not found"})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"message": "Delivered"})

	})

	//send message to all users connected to chat
	router.POST("/send_message", func(c *gin.Context) {
		msg := c.Query("message")
		author := c.Query("author")
		chat_id, err := strconv.Atoi(c.Query("chat_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to parse chat_id"})
			return
		}
		trimmed := strings.TrimSpace(author)
		if trimmed == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "No author"})
			return
		}
		for ch, users := range server.Chats {
			if ch.Chat_id == chat_id {
				for _, user := range users {
					user.Write(author + ": " + msg + "\n" + "\r")
				}
				break
			}
			c.JSON(http.StatusBadRequest, gin.H{"message": "Chat not found"})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"message": "Delivered"})
	})

	//broadcast messages to all users connected to the server
	router.POST("/broadcast", func(c *gin.Context) {
		msg := c.Query("message")
		for _, u := range server.Users {
			u.Write("ADMIN:" + msg + "\n")
		}
		c.JSON(http.StatusAccepted, gin.H{"message": "Delivered"})
	})

	router.POST("/close_chat", func(c *gin.Context) {
		chat_name := c.Query("chatname")
		err := server.CloseChat(ctx, chat_name, rManager)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		} else {
			c.JSON(http.StatusAccepted, gin.H{"message": "Closed"})
		}
	})

	router.GET("/logs", func(c *gin.Context) {

	})

	//creates a new chat with provided name
	router.POST("/create_chat", func(c *gin.Context) {
		chat_name := c.Query("chatname")
		rw := sync.RWMutex{}
		rw.Lock()
		if server.CheckChatName(chat_name) {
			go server.NewChat(ctx, chat_name, rManager)
			server.Wg.Add(1)
			c.JSON(http.StatusCreated, gin.H{"message": "Created"})
		} else {
			c.JSON(http.StatusPreconditionFailed, gin.H{"message": "Can't create chat with this name"})
		}
		rw.Unlock()
	})

	router.GET("/chats", func(c *gin.Context) {
		chats := []string{}
		for chat := range server.Chats {
			if chat.Chat_name != "Lobby" {
				chats = append(chats, chat.Chat_name)
			}
		}
		c.JSON(http.StatusAccepted, gin.H{"chats": chats})
	})

	router.Run(":8080")

}
