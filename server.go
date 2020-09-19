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
)

//Pair ...
type Pair struct {
	DeviceID int64
	UserID   int64
}

//PairDeviceHandler ...
type PairDeviceHandler struct {
	createPairDevice CreatePairDevice
}

//ServeHTTP ...
func (ph *PairDeviceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var p Pair

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	defer r.Body.Close()

	err = ph.createPairDevice(p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	fmt.Printf("pair: %#v\n", p)
	w.Write([]byte(`{"status":"active"}`))
}

//CreatePairDevice ...
type CreatePairDevice func(p Pair) error

func createPairDevice(p Pair) error {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	//db, err := sql.Open("postgres", "postgres://gosctihb:CqOz6dVYlooEBPY4quY9KHvySa2OmADZ@arjuna.db.elephantsql.com:5432/gosctihb")

	if err != nil {
		log.Fatal("connect to database error", err)
	}

	defer db.Close()

	_, err = db.Exec("INSERT INTO pairs VALUES ($1,$2);", p.DeviceID, p.UserID)

	if err != nil {
		return err
	}

	return nil
}

func main() {
	fmt.Println("hello hometic : I'm Gopher!!")

	r := mux.NewRouter()
	r.Handle("/pair-device", &PairDeviceHandler{createPairDevice: createPairDevice}).Methods(http.MethodPost)

	addr := fmt.Sprintf("0.0.0.0:%s", os.Getenv("PORT"))

	fmt.Println("addr:", addr)

	server := http.Server{
		Addr:    addr,
		Handler: r,
	}

	log.Println("staring...")
	log.Fatal(server.ListenAndServe())
}
