package api

import(
	"net/http"
	"io"
	"sync"
	"github.com/gin-gonic/gin"
	"fmt"
)

func Start_api(wg *sync.WaitGroup) {
	defer wg.Done()
	router := gin.Default()

	//syntax emaple:
	//http://localhost:8080/disconnect/?chat_id=1&username=lol
	router.DELETE("/disconnect", func(c *gin.Context) {
		fmt.Println(c.Query("chat_id"))
		fmt.Println(c.Query("username"))
		c.JSON(http.StatusAccepted, gin.H{"message": "Deleted"})
	})

	router.DELETE("/delete_chat", func(c *gin.Context) {

		c.JSON(http.StatusAccepted, gin.H{"message": "created"})
	})

	router.GET("/logs", func(c *gin.Context) {

	})

	router.POST("/create_chat", func(c *gin.Context) {
		jsonData, err := io.ReadAll(c.Request.Body)
		if err != nil {
			// Handle error
		}
		fmt.Println(jsonData)
	})

	router.GET("/chats", func(c *gin.Context) {

	})

	router.Run(":8080")

}