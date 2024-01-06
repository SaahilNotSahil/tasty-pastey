package main

import (
	"log"
	"os"

	"github.com/XanderWatson/tasty-pastey/controllers"
	"github.com/XanderWatson/tasty-pastey/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	mode := os.Getenv("MODE")
	if mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else if mode == "development" {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()

	auth := r.Group("/auth/v1")
	{
		auth.POST("/signup", controllers.SignupController)
		auth.POST("/login", controllers.LoginController)
	}

	v1 := r.Group("/api/v1").Use(middlewares.Authz())
	{
		v1.POST("/paste", controllers.CreatePasteController)
		v1.GET("/paste", controllers.GetPastesController)
		v1.GET("/paste/:id", controllers.GetPasteController)
		v1.GET("/paste/:id/file", controllers.GetPasteFileController)
		v1.PUT("/paste/:id", controllers.UpdatePasteController)
		v1.DELETE("/paste/:id", controllers.DeletePasteController)
		v1.POST("/share", controllers.CreatePasteAccessController)
		v1.DELETE("/share", controllers.DeletePasteAccessController)
	}

	r.Run("0.0.0.0:8000")
}
