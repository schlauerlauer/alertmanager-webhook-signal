package main

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
)

type Signal struct {
	Attachments		[]string	`json:"base64_attachments"`
	Message			string		`json:"message"`
	Number			string		`json:"number"`
	Recipients 		[]string	`json:"recipients"`
}

var Signals []Signal

func main() {
	Signals = []Signal{
		Signal{Message: "", Number: "", Recipients: []string{"recipients"}},
	}
	handleRequests()
}

func handleRequests() {
    router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/v2/send", receiveMessage).Methods("POST")
    log.Fatal(http.ListenAndServe(":10001", router))
}

func receiveMessage(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var Signal Signal
	json.Unmarshal(reqBody, &Signal)
	fmt.Println(Signal.Message)
}
