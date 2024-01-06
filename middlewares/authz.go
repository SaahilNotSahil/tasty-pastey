package middlewares

import (
	"log"
	"net/http"
	"strings"

	"github.com/XanderWatson/tasty-pastey/auth"
	"github.com/XanderWatson/tasty-pastey/database"
	"github.com/XanderWatson/tasty-pastey/models"
	"github.com/gin-gonic/gin"
)

func Authz() gin.HandlerFunc {
	return func(c *gin.Context) {
		pasteId, found := c.Params.Get("id")
		if found {
			var paste models.Paste

			result := database.DB.Where("id = ?", pasteId).First(&paste)
			if result.Error != nil {
				log.Println(result.Error)

				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Error fetching paste",
				})
				c.Abort()

				return
			}

			if paste.Visibility == 0 {
				if c.Request.Method == "GET" {
					c.Next()

					return
				}
			}
		}

		clientToken := c.Request.Header.Get("Authorization")

		if clientToken == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "No Authorization header provided",
			})
			c.Abort()

			return
		}

		extractedToken := strings.Split(clientToken, "Bearer ")
		if len(extractedToken) == 2 {
			clientToken = strings.TrimSpace(extractedToken[1])
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Incorrect Format of Authorization Token",
			})
			c.Abort()

			return
		}

		Jwt := auth.Jwt{
			SecretKey: "verysecretkey",
			Issuer:    "AuthService",
		}

		claims, err := Jwt.ValidateToken(clientToken)
		if err != nil {
			log.Println(err)

			c.JSON(http.StatusUnauthorized, err.Error())
			c.Abort()

			return
		}

		c.Set("email", claims.Email)
		c.Next()
	}
}
