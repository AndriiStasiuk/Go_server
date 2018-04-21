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

var db *gorm.DB
var err error

func main() {
	router := mux.NewRouter()

	host := os.Getenv("HOST")
	user := os.Getenv("USER")
	password := os.Getenv("PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprint("host=%v user=%v dbname=%v sslmode=disable password=%v", host, user, dbName, password)

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
	db.Find(&resources)
	json.NewEncoder(w).Encode(&resources)

	err := json.NewEncoder(w).Encode(&resources)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetResource(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var resource Product
	db.First(&resource, params["key"])
	json.NewEncoder(w).Encode(&resource)

	err := json.NewEncoder(w).Encode(&resource)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}

func CreateResource(w http.ResponseWriter, r *http.Request) {
	var resource Product
	json.NewDecoder(r.Body).Decode(&resource)
	db.Create(&resource)
	json.NewEncoder(w).Encode(&resource)

	err := json.NewEncoder(w).Encode(&resource)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func DeleteResource(w http.ResponseWriter, r *http.Request) {

	var resource Product
	db.First(&resource)
	db.Delete(&resource)

	var resources []Product
	db.Find(&resources)
	json.NewEncoder(w).Encode(&resources)

	err := json.NewEncoder(w).Encode(&resources)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
	}
}
