package client

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

func requestRaw(method, url string, headers map[string]string, r io.Reader) (response interface{}, err error) {
	fullURL := urlPrefix + url
	log.Printf("%s %s", method, fullURL)
	req, _ := http.NewRequest(method, fullURL, r)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	log.Println("req ContentLength:", req.ContentLength)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to make %s request to %s: %s", method, fullURL, err)
		return
	}
	defer resp.Body.Close()

	log.Println("response Status:", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)

	log.Println("response Body:", string(body))

	return
}

func request(method, url string, headers map[string]string, data []byte) (response interface{}, err error) {
	r := bytes.NewBuffer(data)

	return requestRaw(method, url, headers, r)
}

// GetRoom - get room(s)
func GetRoom(roomID int) (err error) {
	urlPostfix := roomURL
	if roomID >= 0 {
		urlPostfix = fmt.Sprintf("%s/%d", roomURL, roomID)
	}

	_, err = request("GET", urlPostfix, jsonHeader, emptyByte)
	if err != nil {
		err = fmt.Errorf("failed to get rooms: %s", err)
		return
	}
	return
}

// GetMessages - get messages from room
func GetMessages(roomID int, fromMessageID int64) (err error) {
	if roomID < 0 {
		roomID = 0
	}
	urlPostfix := fmt.Sprintf("%s/%d%s?from=%d", roomURL, roomID, messageURL, fromMessageID)

	_, err = request("GET", urlPostfix, jsonHeader, emptyByte)
	if err != nil {
		err = fmt.Errorf("failed to get messages: %s", err)
		return
	}
	// TODO: sort.Sort() - https://golang.org/pkg/sort/#example__sortKeys
	return
}

// Subscribe to room
func Subscribe(roomID int) (err error) {

	time.Sleep(config.Conf.Poll * time.Second)
	return
}

// SendMessage - send message to room
func SendMessage(roomID int, message, file string) (err error) {
	urlPostfix := fmt.Sprintf("%s/%d%s", roomURL, roomID, messageURL)
	data := fmt.Sprintf(`{"text":"%s"}`, message)

	_, err = request("POST", urlPostfix, jsonHeader, []byte(data))
	if err != nil {
		err = fmt.Errorf("failed to send message: %s", err)
		return
	}
	return
}

func getFileID() (response string, err error) {
	_, err = request("GET", fileIDURL, jsonHeader, emptyByte)
	if err != nil {
		err = fmt.Errorf("failed to send message: %s", err)
		return
	}
	return
}

// UploadFile - upload file to room
func UploadFile(roomID int, filePath, fileID string) error {
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

			log.Println("new buf size:", len(b))

		} else if err != nil {
			return fmt.Errorf("failed to read from file(%s): %s", filePath, err)
		}

		if copied == 0 {
			break
		}

		err = uploadFile(b, fileID, from, total)
		if err != nil {
			return fmt.Errorf("failed to upload file(%s): %s", filePath, err)
		}

		if last {
			break
		}

		from += int64(copied)
	}

	return nil
}

func uploadFile(data []byte, fileID string, from, total int64) (err error) {
	urlPostfix := fmt.Sprintf("%s/%s", uploadURL, fileID)
	to := from + int64(len(data)) - 1
	contentRange := fmt.Sprintf("bytes %d-%d/%d", from, to, total)
	log.Println("Content-Range:", contentRange)
	headers := map[string]string{
		"Content-Type":  "application/octet-stream",
		"Content-Range": contentRange,
	}
	_, err = request("POST", urlPostfix, headers, data)
	if err != nil {
		err = fmt.Errorf("failed to upload file: %s", err)
		return
	}
	return
}

func init() {
	// log.SetOutput()
	urlPrefix = fmt.Sprintf("http://%s:%d", config.Conf.Host, config.Conf.Port)
	jsonHeader = map[string]string{"Content-Type": "application/json"}
}
