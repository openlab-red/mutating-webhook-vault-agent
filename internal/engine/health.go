package engine

import "github.com/gin-gonic/gin"

func health(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "UP",
	})
}
