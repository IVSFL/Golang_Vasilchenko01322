package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

type CarBrand struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Country       string `json:"country"`
	Year          int    `json:"year"`
	Capitalization int   `json:"capitalization"`
}

type User struct {
	ID int `json:"id"`
	Login string `json:"login"`
	Password string `json:"password"`
}

var db *sql.DB
var jwtKey []byte

// Хэширование пароля
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// Проверка хэшированного пароля
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Генерация JWT токена
func generateToken(login string) (string, error) {
	claims := jwt.RegisteredClaims {
		Subject: login,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt: jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// Валидация JWT токена
func parseToken(tokenStr string) (*jwt.RegisteredClaims, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil // вернем ключ для проверки
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// Middleware для CORS с логированием preflight-запросов
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Разрешаем запросы только с вашего фронтенда
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, X-Requested-With, Authorization")

		if r.Method == "OPTIONS" {
			log.Println("Preflight request:", r.Method, r.URL.Path, r.Header)
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Проверка токена, для защищенных маршрутов
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		_, err := parseToken(tokenStr)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

//Регистрация нового пользователя
func register(w http.ResponseWriter, r *http.Request) {
	var u User
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if u.Login == "" || u.Password == "" {
		http.Error(w, "login and password required", http.StatusBadRequest)
		return
	}

	// Хэшируем пароль
	hash, _ := hashPassword(u.Password)

	// Сохранение в БД
	_, err := db.Exec("INSERT INTO users(login, password) VALUES(?, ?)", u.Login, hash)
	if err != nil {
		http.Error(w, "user exist or insert failed", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// Логин пользователя
func login(w http.ResponseWriter, r *http.Request) {
	var u User
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Ищем пользователя в БД
	var storeHash string
	err := db.QueryRow("SELECT password FROM users WHERE login = ?", u.Login).Scan(&storeHash)
	if err != nil {
		http.Error(w, "no access", http.StatusBadRequest)
		return
	}

	// Проверка пароля
	if !checkPasswordHash(u.Password, storeHash) {
		http.Error(w, "invalid password", http.StatusUnauthorized)
		return
	}

	// Генератор токена
	token, _ := generateToken(u.Login)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// Create
func createCarBrand(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	var cb CarBrand
	if err := json.NewDecoder(r.Body).Decode(&cb); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	res, err := db.Exec("INSERT INTO car_brands(name, country, year, capitalization) VALUES(?, ?, ?, ?)", cb.Name, cb.Country, cb.Year, cb.Capitalization)
	if err != nil {
		http.Error(w, "insert failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	id, _ := res.LastInsertId()
	cb.ID = int(id)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cb)
}

// Read all
func getCarBrands(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	rows, err := db.Query("SELECT id, name, country, year, capitalization FROM car_brands")
	if err != nil {
		http.Error(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	carBrands := []CarBrand{}
	for rows.Next() {
		var cb CarBrand
		if err := rows.Scan(&cb.ID, &cb.Name, &cb.Country, &cb.Year, &cb.Capitalization); err != nil {
			http.Error(w, "scan failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		carBrands = append(carBrands, cb)
	}
	json.NewEncoder(w).Encode(carBrands)
}

// Read one
func getCarBrand(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var cb CarBrand
	err = db.QueryRow("SELECT id, name, country, year, capitalization FROM car_brands WHERE id = ?", id).
		Scan(&cb.ID, &cb.Name, &cb.Country, &cb.Year, &cb.Capitalization)
	if err == sql.ErrNoRows {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(cb)
}

// Update
func updateCarBrand(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var cb CarBrand
	if err := json.NewDecoder(r.Body).Decode(&cb); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE car_brands SET name = ?, country = ?, year = ?, capitalization = ? WHERE id = ?", cb.Name, cb.Country, cb.Year, cb.Capitalization, id)
	if err != nil {
		http.Error(w, "update failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	cb.ID = id
	json.NewEncoder(w).Encode(cb)
}

// Delete
func deleteCarBrand(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid ID", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM car_brands WHERE id = ?", id)
	if err != nil {
		http.Error(w, "delete failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	// Читаем JWT ключ из переменной окружения
	k := os.Getenv("JWT_KEY")
	if k == "" {
		log.Println("WARNING")
		k = "dev_secret_replace_me"
	}
	jwtKey = []byte(k)

	var err error

	db, err = sql.Open("sqlite", "carBrand.db")
	if err != nil {
		log.Fatalf("DB open: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	// Создание таблицы
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS car_brands (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		country TEXT NOT NULL,
		year INTEGER NOT NULL,
		capitalization INTEGER NOT NULL
	);
`)
if err != nil { log.Fatalf("DB create car_brands: %v", err) }

_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		login TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	);
`)
if err != nil { log.Fatalf("DB create users: %v", err) }

	// Роутер
	router := mux.NewRouter()

	// Публичные запросы
	router.HandleFunc("/register", register).Methods("POST")
	router.HandleFunc("/login", login).Methods("POST")

	// Защищенные запросы
	api := router.PathPrefix("/carBrands").Subrouter()
	api.Use(authMiddleware)
	
	api.HandleFunc("/", createCarBrand).Methods("POST")
	api.HandleFunc("/", getCarBrands).Methods("GET")
	api.HandleFunc("/{id}", getCarBrand).Methods("GET")
	api.HandleFunc("/{id}", updateCarBrand).Methods("PUT")
	api.HandleFunc("/{id}", deleteCarBrand).Methods("DELETE")

	// Запуск сервера
	log.Println("Сервер запущен на порту :8080")
	log.Fatal(http.ListenAndServe(":8080", enableCORS(router)))
}

