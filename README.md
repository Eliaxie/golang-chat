# golang-chat

## Overview

This project is a chat application built with Go. It uses a client-server architecture, with clients communicating with each other through a central server.

## Project Structure

- `main.go`: The entry point of the application.
- `pkg/controller/`: Contains the logic for handling client connections and messages.
  - `controller.go`: Defines the `Controller` struct and its methods for managing clients and messages.
  - `handler.go`: Contains handler functions for different types of messages.
  - `network.go`: Contains functions for network-related tasks.
  - `utils.go`: Contains utility functions used by the controller.
  - `webserver.go`: Contains the code for running the web server.
- `pkg/model/`: Contains the data structures used in the application.
  - `structs.go`: Defines the message and client structs.
- `pkg/utils/`: Contains general utility functions.
- `pkg/view/`: Contains the code for the user interface.

## Application Flow

1. The server starts and waits for clients to connect.
1. When a client connects, it sends a `ConnectionInitMessage` to a peer.
1. The peer responds with a `ConnectionInitResponseMessage` containing the client's ID.

## How to Run

To run the server:

```sh
go run pkg/main.go