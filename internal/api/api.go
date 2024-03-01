package api

import(
	"net/http"
	"sync"
	"github.com/gin-gonic/gin"
	"go_chat/internal/chat"
)

func Start_api(wg *sync.WaitGroup, server *chat.Server) {
	defer wg.Done()
	router := gin.Default()

	//syntax emaple:
	//http://localhost:8080/disconnect/?username=lol
	router.DELETE("/disconnect", func(c *gin.Context) {
		err := server.KickUser(c.Query("username"))
		if err != nil {
			c.JSON(http.StatusAccepted, gin.H{"message": "User not found"})
		} else {
			c.JSON(http.StatusAccepted, gin.H{"message": "Disconnected"})
		}
	})

	router.DELETE("/delete_chat", func(c *gin.Context) {

		c.JSON(http.StatusAccepted, gin.H{"message": "created"})
	})

	router.GET("/logs", func(c *gin.Context) {

	})

	router.POST("/create_chat", func(c *gin.Context) {

	})

	router.GET("/chats", func(c *gin.Context) {
		
	})

	router.Run(":8080")

}