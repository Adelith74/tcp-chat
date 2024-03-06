package api

import (
	"go_chat/internal/chat"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

func Start_api(server *chat.Server) {
	defer server.Wg.Done()
	router := gin.Default()

	//syntax emaple:
	//http://localhost:8080/disconnect/?chat_id=1&username=lol
	router.POST("/disconnect", func(c *gin.Context) {
		username := c.Query("username")
		err := server.KickUser(username)
		if err != nil {
			c.JSON(http.StatusAccepted, gin.H{"message": "User not found"})
		} else {
			c.JSON(http.StatusAccepted, gin.H{"message": "Disconnected"})
		}
	})

	router.POST("/auth", func(c *gin.Context) {

	})

	//syntax emaple:
	//http://localhost:8080/delete_chat/?chat_name=hello
	router.DELETE("/delete_chat", func(c *gin.Context) {
		c.JSON(http.StatusAccepted, gin.H{"message": "Deleted"})
	})

	router.POST("/close_chat", func(c *gin.Context) {
		chat_name := c.Query("chatname")
		err := server.CloseChat(chat_name)
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
			go server.NewChat(chat_name)
			c.JSON(http.StatusCreated, gin.H{"message": "Created"})
		} else {
			c.JSON(http.StatusPreconditionFailed, gin.H{"message": "Can't create chat with this name"})
		}
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
