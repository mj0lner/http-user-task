package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
	"time"
	"user-tasks/database"
	"user-tasks/models"
)

var JWTSecret = []byte("supersecretkey")

// Обработчик регистрации
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Referrer *int   `json:"referrer,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Проверка существования пользователя
	var existingUser models.User
	result := database.DB.Db.Where("username = ?", input.Username).First(&existingUser)
	if result.Error == nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Создание пользователя
	newUser := models.User{
		Username: input.Username,
		Password: string(hashedPassword),
		Referrer: input.Referrer,
	}

	result = database.DB.Db.Create(&newUser)
	if result.Error != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Генерация JWT токена
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  newUser.ID,
		"username": newUser.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(JWTSecret)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
		"id":    fmt.Sprintf("%d", newUser.ID),
	})
}

func CompleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	// Преобразование ID пользователя
	uid, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Проверка существования целевого пользователя
	var user models.User
	if err := database.DB.Db.First(&user, uid).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	// Декодирование запроса
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Обновляем баланс: текущий баланс + бонус
	result := database.DB.Db.Model(&user).
		Update("Balance", gorm.Expr("balance + ?", task.UserBonus))

	if result.Error != nil {
		log.Printf("Update error: %v", result.Error)
		http.Error(w, "Failed to update bonus", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Bonus added successfully",
		"data": map[string]interface{}{
			"user_id": uid,
		},
	})
}

func AddReferrer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	// Преобразование ID пользователя
	uid, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Декодирование запроса
	var req models.ReferralRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Проверка существования целевого пользователя
	var user models.User
	if err := database.DB.Db.First(&user, uid).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Обновление реферера
	result := database.DB.Db.Model(&user).Update("Referrer", req.ReferrerID)
	if result.Error != nil {
		log.Printf("Update error: %v", result.Error)
		http.Error(w, "Failed to update referrer", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Referrer added successfully",
		"data": map[string]interface{}{
			"user_id":     uid,
			"referrer_id": req.ReferrerID,
		},
	})
}

func GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	var leaderboard []models.User

	// Добавляем Debug() для вывода SQL-запроса в консоль
	result := database.DB.Db.Debug().Order("balance desc").Limit(10).Find(&leaderboard)
	if result.Error != nil {
		log.Printf("Ошибка при запросе к базе: %v", result.Error)
		http.Error(w, "Failed to fetch leaderboard", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(leaderboard); err != nil {
		log.Printf("Ошибка при кодировании JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func GetUserStatus(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ID из URL-параметров
	vars := mux.Vars(r)
	id := vars["id"]

	// Проверяем валидность ID
	if id == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	// Преобразуем ID в число
	userID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var user models.User
	// Ищем пользователя в базе данных
	result := database.DB.Db.First(&user, userID)

	// Обрабатываем возможные ошибки
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			log.Printf("Database error: %v", result.Error)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Возвращаем данные пользователя
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
