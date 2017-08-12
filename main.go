package main

import (
	"log"

	"github.com/SpiritOfWill/chat-client/client"
)

const roomID = 0

func main() {
	// client.GetRoom(-1)

	// if err := client.GetMessages(0, 0); err != nil {
	// 	log.Fatalf("error: %s", err)
	// }

	// if err := client.SendMessage(0, "Buy cheese and bread for breakfast.", ""); err != nil {
	// 	log.Fatalf("error: %s", err)
	// }

	// client.GetRoom(roomID)

	// if err := client.UploadFile(0, "/Users/will/benchmark.bin", "13b33f89-4832-4030-9c75-c7ebc1eaa6ad"); err != nil {
	if err := client.UploadFile(0, "/Users/will/10MB.bin", "13b33f89-4832-4030-9c75-c7ebc1eaa6ad"); err != nil {
		log.Fatalf("error: %s", err)
	}
}
