package main

import (
	"github.com/gorilla/mux"
	"os"
	"log"
	"net/http"
	"encoding/json"
	"github.com/jinzhu/gorm"
	"fmt"
)

type Product struct {
	gorm.Model
    Id         int    `sql:"id"`
    Key        string `sql:"key"`
    FirstName  string `sql:"first_name"`
    SecondName string `sql:"second_name"`
    Status     string `sql:"status"`
}

var db	*gorm.DB
var err	error

func main() {
	router := mux.NewRouter()

	host := os.Getenv("HOST")
	user := os.Getenv("USER")
	password := os.Getenv("PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%v user=%v dbname=%v sslmode=enable password=%v", host, user, dbName, password)

	db, err := gorm.Open("postgres", connStr)

	if err != nil {
		panic("Failed to connect database")
	}

	defer db.Close()

	router.HandleFunc("/std", GetResources).Methods("GET")
	router.HandleFunc("/std/{key}", GetResource).Methods("GET")
	router.HandleFunc("/std", CreateResource).Methods("POST")
	router.HandleFunc("/std/{key}", DeleteResource).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), router))
}

func GetResources(w http.ResponseWriter, r *http.Request) {
	var resources []Product

	if err := db.Find(&resources).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(&resources); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}

func GetResource(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var resource Product
	db.First(&resource, params["key"])
	if err := json.NewEncoder(w).Encode(&resource); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(&resource)
}

func CreateResource(w http.ResponseWriter, r *http.Request) {
	var resource Product
	if err := json.NewDecoder(r.Body).Decode(&resource); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if err := db.Create(&resource).Error; err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(&resource)
}

func DeleteResource(w http.ResponseWriter, r *http.Request) {

	var resource Product
	db.First(&resource)
	db.Delete(&resource)

	var resources []Product
	if err := db.Find(&resources).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(&resources); err != nil {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte(err.Error()))
		return
	}
}
