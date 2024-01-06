package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/XanderWatson/tasty-pastey/auth"
	"github.com/XanderWatson/tasty-pastey/database"
	"github.com/XanderWatson/tasty-pastey/models"
)

type LoginPayload struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshtoken"`
}

func SignupController(c *gin.Context) {
	var user models.User

	err := c.ShouldBindJSON(&user)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Invalid Inputs",
		})
		c.Abort()

		return
	}

	hashedPassword, err := database.HashPassword(user.Password)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Error Hashing Password",
		})
		c.Abort()

		return
	}

	user.Password = hashedPassword
	user.ID = uuid.New()

	err = database.CreateUserRecord(&user)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Error Creating User",
		})
		c.Abort()

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Message": "Sucessfully Registered User",
	})
}

func LoginController(c *gin.Context) {
	var payload LoginPayload
	var user models.User

	err := c.ShouldBindJSON(&payload)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Invalid Inputs",
		})
		c.Abort()

		return
	}

	result := database.DB.Where("email = ?", payload.Email).First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Error": "Invalid User Credentials",
		})
		c.Abort()

		return
	}

	err = database.CheckPassword(payload.Password, &user)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusUnauthorized, gin.H{
			"Error": "Invalid User Credentials",
		})
		c.Abort()

		return
	}

	jwt := auth.Jwt{
		SecretKey:         "verysecretkey",
		Issuer:            "AuthService",
		ExpirationMinutes: 60,
		ExpirationHours:   12,
	}

	signedToken, err := jwt.GenerateToken(user.Email)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Error Signing Token",
		})
		c.Abort()

		return
	}

	signedRefreshToken, err := jwt.RefreshToken(user.Email)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Error Signing Token",
		})
		c.Abort()

		return
	}

	tokenResponse := LoginResponse{
		Token:        signedToken,
		RefreshToken: signedRefreshToken,
	}

	c.JSON(http.StatusOK, tokenResponse)
}
