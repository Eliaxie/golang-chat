package view

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type MenuOption struct {
	Option string
	Action func()
}

func DisplayMenu(options []MenuOption) {
	for i, option := range options {
		fmt.Printf("%d. %s\n", i+1, option.Option)
	}

	for {
		choice := ReadStringTrimmed()

		if choiceInt, err := strconv.Atoi(choice); err == nil && choiceInt > 0 && choiceInt <= len(options) {
			MoveScreenUp()
			options[choiceInt-1].Action()
			break
		} else {
			fmt.Println("Invalid choice")
		}
	}
}

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

func ClearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func MoveScreenUp() {
	fmt.Print("\033[H\033[2J")
}