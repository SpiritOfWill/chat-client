package main

import (
	"flag"
	"log"
	"os"

	"github.com/SpiritOfWill/chat-client/client"
	"github.com/SpiritOfWill/chat-client/config"
)

const roomID = 0

func main() {
	var r = flag.Int("r", config.Conf.DefaultRoom, "Set room ID")
	var a = flag.String("a", "", "Add new message")
	var f = flag.String("f", "", "Path to file for upload")
	var s = flag.Bool("s", true, "Subscribe to messages")

	flag.Parse()

	if *a != "" {
		if err := client.SendMessage(*r, *a); err != nil {
			log.Fatalf("error: %s", err)
		}
		os.Exit(0)
	}

	if *f != "" {
		if err := client.UploadFile(*r, *f); err != nil {
			log.Fatalf("error: %s", err)
		}
		os.Exit(0)
	}

	if *s {
		err := client.Subscribe(*r)
		if err != nil {
			log.Fatalf("error: %s", err)
		}
	}
}
