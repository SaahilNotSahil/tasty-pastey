package controllers

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/XanderWatson/tasty-pastey/database"
	"github.com/XanderWatson/tasty-pastey/internal/fileupload"
	"github.com/XanderWatson/tasty-pastey/internal/keygen"
	"github.com/XanderWatson/tasty-pastey/models"
)

func CreatePasteController(c *gin.Context) {
	log.Println("Inside CreatePasteController")

	email, _ := c.Get("email")

	user, err := database.GetUserByEmail(email.(string))
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error fetching user",
		})

		return
	}

	title := c.Request.Header.Get("Pastey-Title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please provide a title",
		})

		return
	}

	visibilityString := c.Request.Header.Get("Pastey-Visibility")
	if visibilityString == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please provide a visibility",
		})

		return
	}

	visibility, err := strconv.Atoi(visibilityString)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid visibility value",
		})

		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please provide a file",
		})

		return
	}

	f, err := file.Open()
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error opening file",
		})

		return
	}

	defer f.Close()

	var paste models.Paste

	err = c.Bind(&paste)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please provide valid data",
		})

		return
	}

	paste.UserID = user.ID
	paste.ID = keygen.GenerateKey()
	paste.Title = title
	paste.Visibility = visibility

	err = fileupload.UploadFile(paste.ID, &f)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error uploading file",
		})

		return
	}

	err = database.CreatePasteRecord(&paste)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error creating paste",
		})

		return
	}

	var pasteAccess models.PasteAccess

	pasteAccess.ID = uuid.New()
	pasteAccess.PasteID = paste.ID
	pasteAccess.UserID = user.ID

	err = database.CreatePasteAccessRecord(&pasteAccess)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error creating paste access",
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Paste created successfully!",
		"data":    paste,
	})
}

func GetPastesController(c *gin.Context) {
	log.Println("Inside GetPastesController")

	email, _ := c.Get("email")

	user, err := database.GetUserByEmail(email.(string))
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error fetching user",
		})

		return
	}

	userId := user.ID

	var pastes []models.Paste

	pasteAccesses, err := database.GetPasteAccessRecordsByUserId(userId)
	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "No pastes found for this user",
		})

		return
	} else if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error fetching paste accesses",
		})

		return
	}

	for _, pasteAccess := range pasteAccesses {
		paste, err := database.GetPasteByID(pasteAccess.PasteID)
		if err == nil {
			pastes = append(pastes, *paste)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pastes for user with ID: " + userId.String(),
		"data":    pastes,
	})
}

func GetPasteController(c *gin.Context) {
	log.Println("Inside GetPasteController")

	pasteId, found := c.Params.Get("id")
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please provide the ID of the paste",
		})
	}

	paste, err := database.GetPasteByID(pasteId)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error fetching paste",
		})

		return
	}

	if paste.Visibility == 1 {
		email, _ := c.Get("email")

		user, err := database.GetUserByEmail(email.(string))
		if err != nil {
			log.Println(err)

			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error fetching user",
			})

			return
		}

		userId := user.ID

		_, err = database.GetPasteAccessRecordByUserIdAndPasteId(
			userId, pasteId,
		)
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "You are not authorized to view this paste",
			})

			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Paste with ID: " + pasteId,
		"data":    paste,
	})
}

func GetPasteFileController(c *gin.Context) {
	log.Println("Inside GetPasteFileController")

	pasteId, found := c.Params.Get("id")
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please provide the ID of the paste",
		})
	}

	paste, err := database.GetPasteByID(pasteId)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error fetching paste",
		})

		return
	}

	if paste.Visibility == 1 {
		email, _ := c.Get("email")

		user, err := database.GetUserByEmail(email.(string))
		if err != nil {
			log.Println(err)

			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error fetching user",
			})

			return
		}

		userId := user.ID

		_, err = database.GetPasteAccessRecordByUserIdAndPasteId(
			userId, pasteId,
		)
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "You are not authorized to view this paste",
			})

			return
		}
	}

	file, filesize, err := fileupload.GetFile(pasteId)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error fetching file",
		})

		return
	}

	var fileByteArray = make([]byte, filesize)

	_, err = file.Read(fileByteArray)
	if err != nil {
		if err != io.EOF {
			log.Println(err)

			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error reading file",
			})

			return
		}
	}

	c.Data(http.StatusOK, "application/octet-stream", fileByteArray)
}

func UpdatePasteController(c *gin.Context) {
	log.Println("Inside UpdatePasteController")

	email, _ := c.Get("email")

	user, err := database.GetUserByEmail(email.(string))
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error fetching user",
		})

		return
	}

	userId := user.ID

	pasteId, found := c.Params.Get("id")
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please provide the ID of the paste",
		})
	}

	paste, err := database.GetPasteByID(pasteId)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error fetching paste",
		})

		return
	}

	if paste.UserID != userId {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "You are not authorized to update this paste",
		})

		return
	}

	metadataOnly, found := c.GetQuery("metadata")
	if !found || metadataOnly == "false" {
		file, err := c.FormFile("file")
		if err != nil {
			log.Println(err)

			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Please provide a file",
			})

			return
		}

		f, err := file.Open()
		if err != nil {
			log.Println(err)

			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error opening file",
			})

			return
		}

		defer f.Close()

		err = fileupload.UploadFile(pasteId, &f)
		if err != nil {
			log.Println(err)

			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error uploading file",
			})

			return
		}

		title := c.Request.Header.Get("Pastey-Title")
		if title != "" {
			paste.Title = title
		}

		visibilityString := c.Request.Header.Get("Pastey-Visibility")
		if visibilityString != "" {
			visibility, err := strconv.Atoi(visibilityString)
			if err != nil {
				log.Println(err)

				c.JSON(http.StatusBadRequest, gin.H{
					"message": "Invalid visibility value",
				})

				return
			}

			paste.Visibility = visibility
		}

		err = database.UpdatePasteRecord(pasteId, paste)
		if err != nil {
			log.Println(err)

			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error updating paste",
			})

			return
		}
	} else {
		var paste models.Paste

		err = c.Bind(&paste)
		if err != nil {
			log.Println(err)

			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Please provide valid data",
			})

			return
		}

		err = database.UpdatePasteRecord(pasteId, &paste)
		if err != nil {
			log.Println(err)

			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error updating paste",
			})

			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Paste updated successfully!",
	})
}

func DeletePasteController(c *gin.Context) {
	log.Println("Inside DeletePasteController")

	email, _ := c.Get("email")

	user, err := database.GetUserByEmail(email.(string))
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error fetching user",
		})

		return
	}

	userId := user.ID

	pasteId, found := c.Params.Get("id")
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please provide the ID of the paste",
		})

		return
	}

	paste, err := database.GetPasteByID(pasteId)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error fetching paste",
		})

		return
	}

	if paste.UserID != userId {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "You are not authorized to delete this paste",
		})

		return
	}

	err = fileupload.DeleteFile(pasteId)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error deleting file",
		})

		return
	}

	pasteAccesses, err := database.GetPasteAccessRecordsByPasteId(pasteId)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error fetching paste accesses",
		})

		return
	}

	for _, pasteAccess := range pasteAccesses {
		err = database.DeletePasteAccessRecord(&pasteAccess)
		if err != nil {
			log.Println(err)

			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error deleting paste access",
			})

			return
		}
	}

	err = database.DeletePasteRecord(paste)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error deleting paste",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Paste deleted successfully!",
	})
}

func CreatePasteAccessController(c *gin.Context) {
	log.Println("Inside CreatePasteAccessController")

	email, _ := c.Get("email")

	owner, err := database.GetUserByEmail(email.(string))
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error fetching user",
		})

		return
	}

	ownerId := owner.ID

	pasteId, found := c.GetQuery("paste_id")
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please provide the ID of the paste",
		})

		return
	}

	paste, err := database.GetPasteByID(pasteId)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error fetching paste",
		})

		return
	}

	if paste.UserID != ownerId {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "You are not authorized to share this paste",
		})

		return
	}

	userEmail, found := c.GetQuery("user_email")
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please provide the email of the user",
		})

		return
	}

	user, err := database.GetUserByEmail(userEmail)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error fetching user",
		})

		return
	}

	userId := user.ID

	var pasteAccess models.PasteAccess

	pasteAccess.ID = uuid.New()
	pasteAccess.PasteID = pasteId
	pasteAccess.UserID = userId

	err = database.CreatePasteAccessRecord(&pasteAccess)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error creating paste access",
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Paste access created successfully!",
		"data":    pasteAccess,
	})
}

func DeletePasteAccessController(c *gin.Context) {
	log.Println("Inside DeletePasteAccessController")

	email, _ := c.Get("email")

	owner, err := database.GetUserByEmail(email.(string))
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error fetching user",
		})

		return
	}

	ownerId := owner.ID

	pasteId, found := c.GetQuery("paste_id")
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please provide the ID of the paste",
		})

		return
	}

	paste, err := database.GetPasteByID(pasteId)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error fetching paste",
		})

		return
	}

	if paste.UserID != ownerId {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "You are not authorized to share this paste",
		})

		return
	}

	userEmail, found := c.GetQuery("user_email")
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please provide the email of the user",
		})

		return
	}

	user, err := database.GetUserByEmail(userEmail)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error fetching user",
		})

		return
	}

	userId := user.ID

	pasteAccess, err := database.GetPasteAccessRecordByUserIdAndPasteId(
		userId, pasteId,
	)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error fetching paste access",
		})

		return
	}

	err = database.DeletePasteAccessRecord(pasteAccess)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error deleting paste access",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Paste access deleted successfully!",
	})
}
