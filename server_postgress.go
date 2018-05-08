package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"os"
	"github.com/jinzhu/gorm"
	"time"
	"strconv"
)

type Log struct	{
	Id			int64		`sql:"id"`
	UserId		int64		`sql:"user_id"`
	CreatedAt	time.Time	`sql:"created_at"`
	EventType	string		`sql:"event_type"`
}

type User struct {

	Id              int64       `sql:"id"`
	CardKey         int64       `sql:"card_key"`
	FirstName       string      `sql:"first_name"`
	LastName        string      `sql:"last_name"`
	Status          string      `sql:"status"`
	LastCheckedIn   time.Time	 `sql:"last_checked_in"`
}


var db  *gorm.DB


func main() {

	host := os.Getenv("HOST")
	user := os.Getenv("USER")
	password := os.Getenv("PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%v user=%v dbname=%v sslmode=require password=%v", host, user, dbName, password)
	ddb, err := gorm.Open("postgres", connStr)
	db = ddb
	if err != nil {
		panic("Failed to connect database")
	}
	ddb.LogMode(true)

	defer ddb.Close()

	router := mux.NewRouter()

	router.HandleFunc("/std", GetResources).Methods("GET")
	router.HandleFunc("/std/user/{card_key}", GetResource).Methods("GET")
	router.HandleFunc("/std/create", CreateResource).Methods("POST")
	router.HandleFunc("/std/delete/{id}", DeleteResource).Methods("DELETE")

	router.HandleFunc("/std/logs/{id}", GetLogs).Methods("GET")

	http.ListenAndServe(":8000", router)

}

func GetResources(w http.ResponseWriter, r *http.Request) {
	var users []User
	if err := db.Find(&users).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)

}

func GetResource(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var resource User
	db.First(&resource, params["card_key"])
	i,_:=strconv.ParseInt(params["card_key"], 10, 64)
	if err := db.Where("card_key = ?", i).First(&resource).Error; err != nil {
		json.NewEncoder(w).Encode(http.StatusNotFound)
		return
	}
	if err := db.Where("status = ?", 1).First(&resource).Error; err != nil {
		json.NewEncoder(w).Encode(http.StatusForbidden)
		return
	}

	if err := db.Where("status = ?", 0).First(&resource).Error; err != nil {
	json.NewEncoder(w).Encode(http.StatusOK)
	return
	}

}

func CreateResource(w http.ResponseWriter, r *http.Request) {
	var resource User
	if err := json.NewDecoder(r.Body).Decode(&resource); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	defer r.Body.Close()

	if err := db.Create(&resource).Error; err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(&resource)
}

func DeleteResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user := User{Id: id}
	if err := db.Delete(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func GetLogs(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var resource Log
	db.First(&resource, params["id"])
	if err := db.Where("id = ?", params).First(&resource).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(&resource)
}

