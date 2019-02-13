package utils

import (
	"log"
	"os"
)

func SetupLog(name string) {
	f, err := os.OpenFile("log.txt", os.O_RDWR | os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	//log.SetOutput(f)
	log.SetPrefix(name + " ")
	log.Print("#########################################")
	log.Println("Start server...")
}
