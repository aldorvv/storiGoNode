package main

import (
	"github.com/gin-gonic/gin"
	"github.com/storiGoNode/internal/views"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	router := gin.Default()
	router.GET("/accounts", views.GetAccounts)
	router.GET("/accounts/:id", views.GetAccounts)
	router.POST("/accounts/", views.PostAccount)
	router.GET("/accounts/:id/summary/", views.GetSummary)

	router.Run("0.0.0.0:8000")
}
