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

	if err := run(); err != nil {
		log.Fatal("can't start application", err)
	}
}

func run() error {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	//db, err := sql.Open("postgres", "postgres://gosctihb:CqOz6dVYlooEBPY4quY9KHvySa2OmADZ@arjuna.db.elephantsql.com:5432/gosctihb")

	if err != nil {
		//log.Fatal("connect to database error", err)
		return err
	}

	defer db.Close()

	r := mux.NewRouter()

	r.Use(logger.Middleware)

	r.Handle("/pair-device", CustomHandlerFunc(PairDeviceHandler(NewCreatePairDevice(db)))).Methods(http.MethodPost)

	addr := fmt.Sprintf("0.0.0.0:%s", os.Getenv("PORT"))

	fmt.Println("addr:", addr)

	server := http.Server{
		Addr:    addr,
		Handler: r,
	}

	log.Println("staring...")
	return server.ListenAndServe()
}

//CustomResponseWriter ...
type CustomResponseWriter interface {
	JSON(statusCode int, data interface{})
}

//CustomHandlerFunc ...
type CustomHandlerFunc func(CustomResponseWriter, *http.Request)

//ServerHTTP ...
func (handler CustomHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler(&JSONResponseWriter{w}, r)
}

//JSONResponseWriter ...
type JSONResponseWriter struct {
	http.ResponseWriter
}

//JSON ...
func (w *JSONResponseWriter) JSON(statusCode int, data interface{}) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

//PairDeviceHandler ...
func PairDeviceHandler(device Device) func(w CustomResponseWriter, r *http.Request) {

	return func(w CustomResponseWriter, r *http.Request) {
		logger.L(r.Context()).Info("pair-device")

		var p Pair

		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			w.JSON(http.StatusBadRequest, err.Error())
			return
		}

		defer r.Body.Close()
		fmt.Printf("pair: %#v\n", p)

		err = device.Pair(p)
		if err != nil {
			w.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		w.JSON(http.StatusOK, map[string]interface{}{"status": "active"})
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

//DB ...
type DB interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

//NewCreatePairDevice ...
func NewCreatePairDevice(db DB) CreatePairDeviceFunc {
	return func(p Pair) error {
		_, err := db.Exec("INSERT INTO pairs VALUES ($1,$2);", p.DeviceID, p.UserID)
		return err

	}
}
