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
	Id		int64		`sql:"id"`
	UserId		int64		`sql:"user_id"`
	CreatedAt	time.Time	`sql:"created_at"`
	EventType	int		`sql:"event_type"`
}

type User struct {

	Id              int64       `sql:"id" json:"id"`
	CardKey         int64       `sql:"card_key" json:"card_key"`
	FirstName       string      `sql:"first_name" json:"first_name"`
	LastName        string      `sql:"last_name" json:"last_name"`
	Status          string      `sql:"status" json:"status"`
	LastCheckedIn   time.Time   `sql:"last_checked_in" json:"last_checked_in"`
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

        router.HandleFunc("/{any:.*}", Options).Methods("OPTIONS")
	
	router.HandleFunc("/std/user", GetResources).Methods("GET")
	router.HandleFunc("/std/user/{card_key}", GetResource).Methods("GET")
	router.HandleFunc("/std/user", CreateResource).Methods("POST")
	router.HandleFunc("/std/user/delete/{id}", DeleteResource).Methods("DELETE")
	router.HandleFunc("/std/user/update/{id}", UpdateResource).Methods("PUT")
	
	router.HandleFunc("/std/user/blocked/{id}",BlockedUser).Methods("PUT")
	router.HandleFunc("/std/user/unblocked/{id}",UnblockedUser).Methods("PUT")
	
	router.HandleFunc("/std/auth",AuthUser).Methods("POST")
	router.HandleFunc("/std/exit",UserExit).Methods("POST")
	
	router.HandleFunc("/std/logs", GetLogs).Methods("GET")
	router.HandleFunc("/std/logs/{user_id}", GetLog).Methods("GET")

	http.ListenAndServe(":" + os.Getenv("PORT"), router)

}

func Options(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, PUT, OPTIONS, PATCH, DELETE")
}

func WriteResult(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)

	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func GetResources(w http.ResponseWriter, r *http.Request) {
	var users []User
	if err := db.Find(&users).Error; err != nil {
		WriteResult(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteResult(w, http.StatusOK, users)

}

func GetLog(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var logs []Log
	i, err := strconv.ParseInt(params["user_id"], 10, 64)

	if err != nil {
		WriteResult(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := db.Where("user_id = ?", i).Find(&logs).Error; err != nil {
		WriteResult(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteResult(w, http.StatusOK, logs)

}

func GetLogs(w http.ResponseWriter, r *http.Request) {
	var logs []Log
	if err := db.Find(&logs).Error; err != nil {
		WriteResult(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteResult(w, http.StatusOK, logs)

}

func GetResource(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var resource User
	i, err := strconv.ParseInt(params["card_key"], 10, 64)

	if err != nil {
		WriteResult(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := db.Where("card_key = ?", i).First(&resource).Error; err != nil {
		WriteResult(w, http.StatusNotFound, http.StatusNotFound)
		return
	}

	if err := db.Where("status = ?", 0).First(&resource).Error; err != nil {
		WriteResult(w, http.StatusOK, http.StatusOK)
		return
	}

	WriteResult(w, http.StatusForbidden, http.StatusForbidden)

}

func CreateResource(w http.ResponseWriter, r *http.Request) {
	var resource User
	if err := json.NewDecoder(r.Body).Decode(&resource); err != nil {
		WriteResult(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	if err := db.Create(&resource).Error; err != nil {
		WriteResult(w, http.StatusBadRequest, err.Error())
		return
	}
	log := Log{UserId: resource.Id, CreatedAt: time.Now(), EventType: 6}
	if err := db.Create(&log).Error; err != nil {
		WriteResult(w, http.StatusBadRequest, err.Error())
	}
	WriteResult(w, http.StatusOK, resource)
}

func DeleteResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		WriteResult(w, http.StatusNotFound, err.Error())
		return
	}
	user := User{Id: id}
	if err := db.Delete(&user).Error; err != nil {
		WriteResult(w, http.StatusBadRequest, err.Error)
		return
	}

	log := Log{UserId: id, CreatedAt: time.Now(), EventType: 8}
	if err := db.Create(&log).Error; err != nil {
		WriteResult(w, http.StatusBadRequest, err.Error())
	}

	WriteResult(w, http.StatusOK, id)
}

func UpdateResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		WriteResult(w, http.StatusNotFound, err.Error())
		return
	}

	var resource User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&resource); err != nil {
		WriteResult(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	resource.Id = id
	if err := db.Save(&resource).Error; err != nil {
		WriteResult(w, http.StatusInternalServerError, err)
		return
	}
	log := Log{UserId: id, CreatedAt: time.Now(), EventType: 7}
	if err := db.Create(&log).Error; err != nil {
		WriteResult(w, http.StatusBadRequest, err.Error())
	}

	WriteResult(w, http.StatusOK, resource)
}

func BlockedUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		WriteResult(w, http.StatusNotFound, err.Error())
		return
	}
	var user User

	user.Id = id

	if err := db.First(&user).Error; err != nil {
		WriteResult(w, http.StatusBadRequest, err.Error())
		return
	}

	user.Status = "0"

	if err := db.Save(&user).Error; err != nil {
		WriteResult(w, http.StatusInternalServerError, err)
		return
	}
	log := Log{UserId: id, CreatedAt: time.Now(), EventType: 4}
	if err := db.Create(&log).Error; err != nil {
		WriteResult(w, http.StatusBadRequest, err.Error())
	}

	WriteResult(w, http.StatusOK, user)

}

func UnblockedUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		WriteResult(w, http.StatusNotFound, err.Error())
		return
	}
	var user User

	user.Id = id

	if err := db.First(&user).Error; err != nil {
		WriteResult(w, http.StatusBadRequest, err.Error())
		return
	}

	user.Status = "1"

	if err := db.Save(&user).Error; err != nil {
		WriteResult(w, http.StatusInternalServerError, err)
		return
	}
	log := Log{UserId: id, CreatedAt: time.Now(), EventType: 5}
	if err := db.Create(&log).Error; err != nil {
		WriteResult(w, http.StatusBadRequest, err.Error())
	}
	WriteResult(w, http.StatusOK, user)
}

func AuthUser(w http.ResponseWriter, r *http.Request) {

	var resource User
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&resource);
	if err != nil {
		WriteResult(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	if err := db.Where("card_key = ?", resource.CardKey).First(&resource).Error; err != nil {
		log := Log{UserId: 0, CreatedAt: time.Now(), EventType: 3}
		if err := db.Create(&log).Error; err != nil {
			WriteResult(w, http.StatusBadRequest, err.Error())
		}
		WriteResult(w, http.StatusNotFound, nil)
		return
	}

	if resource.Status == "1" {
		resource.LastCheckedIn = time.Now()
		if err := db.Save(&resource).Error; err != nil {
			WriteResult(w, http.StatusInternalServerError, err)
			return
		}
		log := Log{UserId: resource.Id, CreatedAt: time.Now(), EventType: 1}
		if err := db.Create(&log).Error; err != nil {
			WriteResult(w, http.StatusBadRequest, err.Error())
		}
		WriteResult(w, http.StatusOK, nil)
	} else {
		log := Log{UserId: resource.Id, CreatedAt: time.Now(), EventType: 2}
		if err := db.Create(&log).Error; err != nil {
			WriteResult(w, http.StatusBadRequest, err.Error())
		}
		WriteResult(w, http.StatusForbidden, nil)
	}

}

func UserExit(w http.ResponseWriter, r *http.Request) {
	log := Log{UserId: 0, CreatedAt: time.Now(), EventType: 9}
	if err := db.Create(&log).Error; err != nil {
		WriteResult(w, http.StatusBadRequest, err.Error())
	}
	WriteResult(w, http.StatusOK, nil)
}

