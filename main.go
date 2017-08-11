package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
)

const (
	configPath = "config.toml"
	roomURL    = "/room"
	messageURL = "/message"
)

var conf Config
var urlPrefix string

// MessageType - message
type MessageType struct {
	ID        int64  `json:"id"`
	Text      string `json:"text"`
	File      string `json:"file,omitempty"`
	Timestamp int64  `json:"time"`
}

// MessagesType - messages
type MessagesType []*MessageType

type sendMessageType struct {
	Text string `json:"text"`
	File string `json:"file,omitempty"`
}

func request(method, url string, data []byte) (response interface{}, err error) {
	fullURL := urlPrefix + url
	log.Printf("%s %s", method, fullURL)
	req, err := http.NewRequest(method, fullURL, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	log.Println("response Status:", resp.Status)
	// log.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)

	log.Println("response Body:", string(body))

	return
}

func getRoom(roomID int) {
	urlPostfix := roomURL
	if roomID >= 0 {
		urlPostfix = fmt.Sprintf("%s/%d", roomURL, roomID)
	}
	request("GET", urlPostfix, []byte(""))
}

func getMessage(roomID int, fromMessageID int64) {
	if roomID < 0 {
		roomID = 0
	}

	urlPostfix := fmt.Sprintf("%s/%d%s?from=%d", roomURL, roomID, messageURL, fromMessageID)
	request("GET", urlPostfix, []byte(""))
}

func addMessage(roomID int, message string) {
	urlPostfix := fmt.Sprintf("%s/%d%s", roomURL, roomID, messageURL)
	data := fmt.Sprintf(`{"text":"%s"}`, message)
	request("POST", urlPostfix, []byte(data))
}

func main() {
	if _, err := toml.DecodeFile(configPath, &conf); err != nil {
		log.Printf("failed to read config: %s", err)
	}

	urlPrefix = fmt.Sprintf("http://%s:%d", conf.Host, conf.Port)

	addMessage(0, "Buy cheese and bread for breakfast.")
	// getRoom(-1)
	// getRoom(0)
	getMessage(0, 0)
}
