package database

import (
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/XanderWatson/tasty-pastey/models"
	"golang.org/x/crypto/bcrypt"
)

var DB *gorm.DB

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env")
	}

	dsn := os.Getenv("DATABASE_URL")

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database", err)
	}

	log.Println("Connected to DB successfully!")

	err = DB.AutoMigrate(&models.User{}, &models.Paste{}, &models.PasteAccess{})
	if err != nil {
		log.Fatal("Failed to migrate models", err)
	}

	log.Println("Migrated models successfully!")
}

func CreateUserRecord(user *models.User) error {
	result := DB.Create(&user)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func CheckPassword(providedPassword string, user *models.User) error {
	err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password), []byte(providedPassword),
	)
	if err != nil {
		return err
	}

	return nil
}

func GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}

	result := DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}

func CreatePasteRecord(paste *models.Paste) error {
	result := DB.Create(&paste)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func GetPasteByID(id string) (*models.Paste, error) {
	var paste models.Paste

	result := DB.Where("id = ?", id).First(&paste)
	if result.Error != nil {
		return nil, result.Error
	}

	return &paste, nil
}

func UpdatePasteRecord(pasteId string, paste *models.Paste) error {
	result := DB.Model(&paste).Where(
		"id = ?", pasteId,
	).Updates(&paste)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func DeletePasteRecord(paste *models.Paste) error {
	result := DB.Delete(&paste)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func CreatePasteAccessRecord(pasteAccess *models.PasteAccess) error {
	result := DB.Create(&pasteAccess)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func GetPasteAccessRecordsByUserId(userId uuid.UUID) (
	[]models.PasteAccess, error,
) {
	pasteAccessRecords := []models.PasteAccess{}

	result := DB.Model(&models.PasteAccess{}).Where(
		"user_id = ?", userId,
	).Find(&pasteAccessRecords)
	if result.Error != nil {
		return nil, result.Error
	}

	return pasteAccessRecords, nil
}

func GetPasteAccessRecordsByPasteId(pasteId string) (
	[]models.PasteAccess, error,
) {
	pasteAccessRecords := []models.PasteAccess{}

	result := DB.Model(&models.PasteAccess{}).Where(
		"paste_id = ?", pasteId,
	).Find(&pasteAccessRecords)
	if result.Error != nil {
		return nil, result.Error
	}

	return pasteAccessRecords, nil
}

func GetPasteAccessRecordByUserIdAndPasteId(userId uuid.UUID, pasteId string) (
	*models.PasteAccess, error,
) {
	var pasteAccess models.PasteAccess

	result := DB.Model(&models.PasteAccess{}).Where(
		"paste_id = ? AND user_id = ?", pasteId, userId,
	).First(&models.PasteAccess{})
	if result.Error != nil {
		return nil, result.Error
	}

	return &pasteAccess, nil
}

func DeletePasteAccessRecord(pasteAccess *models.PasteAccess) error {
	result := DB.Delete(&pasteAccess)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
