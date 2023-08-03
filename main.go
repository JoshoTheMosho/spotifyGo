package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	router := gin.Default()

	router.LoadHTMLGlob("templates/*.tmpl") // Load templates from the "templates" folder
	router.Static("/assets", "./assets")

	router.GET("", func(c *gin.Context) {
		// Render a template
		c.HTML(http.StatusOK, "main.tmpl", gin.H{
			"title": "Joshua Machado",
		})
	})

	router.GET("/spotifyGo", func(c *gin.Context) {
		// Render a template
		c.HTML(http.StatusOK, "spotifyGo.tmpl", gin.H{
			"title": "spotifyGo",
		})
	})

	router.Run(":8080")
}
