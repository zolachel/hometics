package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/zolachel/hometic/logger"
)

func main() {
	fmt.Println("hello hometic : I'm Gopher!!")

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	//db, err := sql.Open("postgres", "postgres://gosctihb:CqOz6dVYlooEBPY4quY9KHvySa2OmADZ@arjuna.db.elephantsql.com:5432/gosctihb")

	if err != nil {
		log.Fatal("connect to database error", err)
	}

	defer db.Close()

	r := mux.NewRouter()

	r.Use(logger.Middleware)

	r.Handle("/pair-device", PairDeviceHandler(NewCreatePairDevice(db))).Methods(http.MethodPost)

	addr := fmt.Sprintf("0.0.0.0:%s", os.Getenv("PORT"))

	fmt.Println("addr:", addr)

	server := http.Server{
		Addr:    addr,
		Handler: r,
	}

	log.Println("staring...")
	log.Fatal(server.ListenAndServe())
}

//PairDeviceHandler ...
func PairDeviceHandler(device Device) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		logger.L(r.Context()).Info("pair-device")

		var p Pair

		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(err.Error())
			return
		}

		defer r.Body.Close()
		fmt.Printf("pair: %#v\n", p)

		err = device.Pair(p)
		if err != nil {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(err.Error())
			return
		}

		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`{"status":"active"}`))
	}
}

//Pair ...
type Pair struct {
	DeviceID int64
	UserID   int64
}

//Device ...
type Device interface {
	Pair(p Pair) error
}

//CreatePairDeviceFunc ...
type CreatePairDeviceFunc func(p Pair) error

//Pair ...
func (fn CreatePairDeviceFunc) Pair(p Pair) error {
	return fn(p)
}

//NewCreatePairDevice ...
func NewCreatePairDevice(db *sql.DB) CreatePairDeviceFunc {
	return func(p Pair) error {
		_, err := db.Exec("INSERT INTO pairs VALUES ($1,$2);", p.DeviceID, p.UserID)
		return err

	}
}
