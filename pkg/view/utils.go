package view

import (
	"log"
	"strconv"
	"strings"
)


func ReadStringTrimmed() string {
	text, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(text)
}

func ReadInt() (int, error) {
	text := ReadStringTrimmed()
	num, err := strconv.Atoi(text)
	return num, err
}
