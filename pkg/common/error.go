package common

import (
	"fmt"
	"log"
	"os"
)

func Check(err error, message string) {
	if err != nil {
		fullMessage := fmt.Sprintf("%s - %s", err.Error(), message)
		Fail(fullMessage)
	}
}

func Fail(message string) {
	log.Printf("Error: %s\n", message)
	os.Exit(1)
}

//func Check(err error, message string) {
//	if err != nil {
//		log.Fatalf("Error: %s - %s\n", err.Error(), message)
//	}
//}

func SilentCheck(err error, message string) {
	if err != nil {
		log.Printf("Error: %s - %s\n", err.Error(), message)
	}
}

func IsError(err error, message string) bool {
	if err != nil {
		log.Printf("Error: %s - %s\n", err.Error(), message)
		return true
	}
	return false
}
