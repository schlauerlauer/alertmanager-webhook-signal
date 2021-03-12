package main

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
)

var port = "10001"

type Signal struct {
	Attachments		[]string	`json:"base64_attachments"`
	Message			string		`json:"message"`
	Number			string		`json:"number"`
	Recipients 		[]string	`json:"recipients"`
}

var Signals []Signal

func main() {
	fmt.Println("Starting signal server for testing. Listening on port", port)
	Signals = []Signal{
		{Message: "", Number: "", Recipients: []string{"recipients"}},
	}
	handleRequests()
}

func handleRequests() {
    router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/v2/send", receiveMessage).Methods("POST")
    log.Fatal(http.ListenAndServe(":"+port, router))
}

func receiveMessage(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var Signal Signal
	json.Unmarshal(reqBody, &Signal)
	fmt.Println(Signal.Message, Signal.Recipients, Signal.Attachments)
}
