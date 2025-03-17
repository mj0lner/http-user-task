package handlers

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Пропускаем авторизацию для регистрации
		if r.URL.Path == "/register" {
			next.ServeHTTP(w, r)
			return
		}

		// Получаем токен из заголовка
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Проверяем формат токена (Bearer <token>)
		if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}
		tokenString = tokenString[7:]

		// Парсим и проверяем токен
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return JWTSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Передаём управление следующему обработчику
		next.ServeHTTP(w, r)
	})
}
