package main

import (
	"database/sql"
	"encoding/json"
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

// Middleware для CORS с логированием preflight-запросов
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Разрешаем запросы только с вашего фронтенда
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, X-Requested-With")

		if r.Method == "OPTIONS" {
			log.Println("Preflight request:", r.Method, r.URL.Path, r.Header)
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
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
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		login TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
		);`)
	if err != nil {
		log.Fatalf("DB create: %v", err)
	}

	// Роутер
	router := mux.NewRouter()
	
	router.HandleFunc("/carBrands", createCarBrand).Methods("POST")
	router.HandleFunc("/carBrands", getCarBrands).Methods("GET")
	router.HandleFunc("/carBrands/{id}", getCarBrand).Methods("GET")
	router.HandleFunc("/carBrands/{id}", updateCarBrand).Methods("PUT")
	router.HandleFunc("/carBrands/{id}", deleteCarBrand).Methods("DELETE")

	// Запуск сервера
	log.Println("Сервер запущен на порту :8080")
	log.Fatal(http.ListenAndServe(":8080", enableCORS(router)))
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
