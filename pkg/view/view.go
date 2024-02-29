package view

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"golang-chat/pkg/utils"
)

var reader = bufio.NewReader(os.Stdin)

func Start() {
	fmt.Println("Welcome to the chat app")
	port := DisplayInsertPort()
	log.Println("Port number entered: ", port)
	username := DisplayInsertUsername()
	log.Println("Username entered: ", username)
	DisplayMainMenu()
}

func DisplayInsertPort() int {
	fmt.Print("Enter port number to start the server on (e.g., 8080): ")
	return ReadInt()
}

func DisplayInsertUsername() string {
	fmt.Print("Enter username: ")
	return ReadStringTrimmed()
}

// Function to display the main menu
func DisplayMainMenu() {
	DisplayMenu([]MenuOption{
		{"Add new connections", DisplayAddNewConnectionsMenu},
		{"Create new Group", DisplayCreateNewGroup},
		{"Open Group", DisplayOpenGroup},
	})
}

func DisplayAddNewConnectionsMenu() {
	// use DisplayMenu to display the options
	DisplayMenu([]MenuOption{
		{"Add Connection From File", DisplayAddConnectionFromFile},
		{"Add Connection Manually", DisplayAddConnectionManually},
		{"Back", DisplayMainMenu},
	})
}

func DisplayAddConnectionFromFile() {
	fmt.Print("Enter the file path (\"q\" to go back): ")
	var connections []string
	for {
		filePath := ReadStringTrimmed()
		if filePath == "q" {
			DisplayAddNewConnectionsMenu()
			break
		}

		// call the function to add the connections from the file
		var err error
		connections, err = utils.ReadConnectionsFromFile(filePath)
		if err == nil {
			break
		}
		fmt.Println("Invalid file path. Please try again.")
	}
	log.Println("Connections added successfully")
	log.Print(connections)
	DisplayMainMenu()
}

func DisplayAddConnectionManually() {
	for {
		fmt.Print("Enter Connection: ")
		connection := ReadStringTrimmed()
		// call the function to add the connection
		log.Println("Connection added successfully")
		log.Print(connection)
		// ask the user if they want to add another connection
		fmt.Println("Do you want to add another connection? (y/n)")
		choice := ReadStringTrimmed()
		if choice == "n" {
			break
		}
	}
	DisplayMainMenu()
}

func DisplayCreateNewGroup() {
	fmt.Print("Enter the group name: ")
	groupName := ReadStringTrimmed()
	// call the function to create a new group
	log.Println("Group" + groupName + " created successfully")
	DisplayMainMenu()
}

func DisplayOpenGroup() {
	fmt.Print("Enter the group name: ")
	groupName := ReadStringTrimmed()
	// call the function to open the group
	log.Println("Opening group " + groupName)
}

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

func ReadInt() int {
	text := ReadStringTrimmed()
	num, err := strconv.Atoi(text)
	if err != nil {
		log.Fatal(err)
	}
	return num
}
