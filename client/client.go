package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/SpiritOfWill/chat-client/config"
)

const (
	roomURL    = "/room"
	messageURL = "/message"
	fileIDURL  = "/file/id"
	uploadURL  = "/upload"
)

var emptyByte []byte
var urlPrefix string
var jsonHeader map[string]string

// RoomType - chat room
type RoomType struct {
	Name          string `json:"name"`
	LastMessageID int64  `json:"last_message_id"`
	// Messages      MessagesType `json:"-"`
}

// RoomsType - rooms type
type RoomsType map[int]*RoomType

// MessageType - message
type MessageType struct {
	ID        int64  `json:"id"`
	Text      string `json:"text"`
	File      string `json:"file,omitempty"`
	Timestamp int64  `json:"time"`
}

// MessagesType - messages
type MessagesType []*MessageType

type newMessageType struct {
	Text string `json:"text"`
	File string `json:"file,omitempty"`
}

func requestRaw(method, url string, headers map[string]string, r io.Reader, i interface{}) error {
	fullURL := urlPrefix + url
	req, _ := http.NewRequest(method, fullURL, r)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make %s request to %s: %s", method, fullURL, err)
	}
	defer resp.Body.Close()
	// TODO: add resp.Status checking
	// log.Println("response Status:", resp.Status)

	body, _ := ioutil.ReadAll(resp.Body)
	if i != nil {
		err := json.Unmarshal(body, i)
		if err != nil {
			return fmt.Errorf("failed to unmarshal body: %s", err)
		}
	}
	return nil
}

func request(method, url string, headers map[string]string, data []byte, i interface{}) (err error) {
	r := bytes.NewBuffer(data)
	return requestRaw(method, url, headers, r, i)
}

// GetRoom - get room
func GetRoom(roomID int) (room *RoomType, err error) {
	urlPostfix := fmt.Sprintf("%s/%d", roomURL, roomID)
	room = &RoomType{}

	err = request("GET", urlPostfix, jsonHeader, emptyByte, room)
	if err != nil {
		err = fmt.Errorf("failed to get room(s): %s", err)
		return
	}
	return
}

// GetMessages - get messages from room
func GetMessages(roomID int, fromMessageID int64) (messages *MessagesType, err error) {
	if roomID < 0 {
		roomID = 0
	}
	urlPostfix := fmt.Sprintf("%s/%d%s?from=%d", roomURL, roomID, messageURL, fromMessageID)
	messages = &MessagesType{}
	err = request("GET", urlPostfix, jsonHeader, emptyByte, messages)
	if err != nil {
		err = fmt.Errorf("failed to get messages: %s", err)
		return
	}
	// TODO: add sort.Sort() - https://golang.org/pkg/sort/#example__sortKeys

	for _, m := range *messages {
		file := ""
		if m.File != "" {
			file = fmt.Sprintf(" (file ID:%s)", m.File)
		}
		fmt.Printf("%v: %s%s\n", time.Unix(m.Timestamp, 0), m.Text, file)
	}
	return
}

func getLatestMessageID(messages *MessagesType) (lastSeenMessageID int64) {
	for _, m := range *messages {
		if m.ID > lastSeenMessageID {
			lastSeenMessageID = m.ID
		}
	}
	return
}

// Subscribe to room
func Subscribe(roomID int) (err error) {
	fmt.Println("Subscribing to room:", roomID)
	messages, err := GetMessages(roomID, 0)
	if err != nil {
		return fmt.Errorf("failed to get messages from room(%d): %s", roomID, err)
	}
	lastSeenMessageID := getLatestMessageID(messages)
	for {
		time.Sleep(config.Conf.Poll * time.Second)

		room, err := GetRoom(roomID)
		if err != nil {
			return fmt.Errorf("failed to get room(%d): %s", roomID, err)
		}

		if room.LastMessageID > lastSeenMessageID {
			fmt.Printf("You have %d new message(s):\n", room.LastMessageID-lastSeenMessageID)
			messages, err := GetMessages(roomID, lastSeenMessageID+1)
			if err != nil {
				return fmt.Errorf("failed to get messages from room(%d): %s", roomID, err)
			}
			lastSeenMessageID = getLatestMessageID(messages)
		}
	}
}

// SendMessage - send message to room
func SendMessage(roomID int, message string) (err error) {
	err = sendMessage(roomID, message, "")
	if err != nil {
		return
	}
	fmt.Println("Sent message to room:", roomID)
	return
}

func sendMessage(roomID int, message, file string) (err error) {
	urlPostfix := fmt.Sprintf("%s/%d%s", roomURL, roomID, messageURL)

	// TODO: add json marshaling newMessageType:
	data := fmt.Sprintf(`{"text":"%s","file":"%s"}`, message, file)

	err = request("POST", urlPostfix, jsonHeader, []byte(data), nil)
	if err != nil {
		err = fmt.Errorf("failed to send message: %s", err)
		return
	}
	return
}

func getFileID() (fileID string, err error) {
	type responseType struct {
		ID string `json:"id"`
	}
	response := &responseType{}
	err = request("GET", fileIDURL, jsonHeader, emptyByte, response)
	if err != nil {
		err = fmt.Errorf("failed to get ID: %s", err)
		return
	}
	fileID = response.ID
	return
}

// UploadFile - upload file to room
func UploadFile(roomID int, filePath string) error {
	fileID, err := getFileID()
	if err != nil {
		return err
	}

	err = uploadFile(roomID, filePath, fileID)
	if err != nil {
		return fmt.Errorf("failed to upload chunks of file(%s): %s", filePath, err)
	}

	err = sendMessage(roomID, fmt.Sprintf("I have sended file: %s", filepath.Base(filePath)), fileID)
	if err != nil {
		return err
	}
	fmt.Println("Sent file to room:", roomID)
	return nil
}

func uploadFile(roomID int, filePath, fileID string) error {
	var from int64

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file(%s): %s", filePath, err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to get stats from file(%s): %s", filePath, err)
	}

	total := fi.Size()
	last := false
	b := make([]byte, config.Conf.MaxChunkSize)
	for {
		copied, err := io.ReadFull(f, b)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			last = true
			// shrink slice:
			b = b[:copied]
		} else if err != nil {
			return fmt.Errorf("failed to read from file(%s): %s", filePath, err)
		}

		if copied == 0 {
			break
		}

		err = uploadChunk(b, fileID, from, total)
		if err != nil {
			return fmt.Errorf("failed to upload chunk(from - %d) for file(%s): %s", from, filePath, err)
		}

		if last {
			break
		}

		from += int64(copied)
	}

	return nil
}

func uploadChunk(data []byte, fileID string, from, total int64) (err error) {
	urlPostfix := fmt.Sprintf("%s/%s", uploadURL, fileID)
	to := from + int64(len(data)) - 1
	contentRange := fmt.Sprintf("bytes %d-%d/%d", from, to, total)
	// log.Println("Content-Range:", contentRange)
	headers := map[string]string{
		"Content-Type":  "application/octet-stream",
		"Content-Range": contentRange,
	}
	err = request("POST", urlPostfix, headers, data, nil)
	if err != nil {
		return
	}
	return
}

func init() {
	urlPrefix = fmt.Sprintf("http://%s:%d", config.Conf.Host, config.Conf.Port)
	jsonHeader = map[string]string{"Content-Type": "application/json"}

	// if config.Conf.LogPath != "" {
	// 	l, err := os.OpenFile(config.Conf.LogPath, os.O_CREATE|os.O_WRONLY, 0644)
	// 	if err != nil {
	// 		log.Panicf("failed to open log file(%s): %s", config.Conf.LogPath, err)
	// 	}
	// 	defer l.Close()
	// 	// logger := log.New(l, "", 0)
	// 	log.SetOutput(l)
	// 	log.Println("start logging...")
	// }
}
