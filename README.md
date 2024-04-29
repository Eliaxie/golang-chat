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
go run main.go
```

## Logs

You can customize the log level at startup using the `-v` flag. The number of `v` characters indicates the log level, as follows:

- `No flag`: PanicLevel
- `-v`: FatalLevel
- `-vv`: ErrorLevel
- `-vvv`: WarnLevel
- `-vvvv`: InfoLevel
- `-vvvvv`: DebugLevel
- `-vvvvvv`: TraceLevel

![connection protocol](https://www.plantuml.com/plantuml/dpng/ZP8nJyCm48Lt_mgpqI7v0GXG0J6m8Jfag2hasa-9rUHSsPV-Ve8ZYpfr9BhTzRvtvvUt3QmyZqClxhZ30AazLwq7IBpqLaDMp_BL7HzaWsDm-WIMbYmBfTbU5EFtpxyY8c9goSTg0cDvZNQAJEZK2KBCyjDScle46KljMsz17FQJIBs3ly2_apaxItoGJvBz21b_zOt0J3PhQBFx3h6P-7XdUjvYUvIiv-fcVmQchIKaxTMnEMFxpXe3EYHtX47cfLZuhXqHfW577sHP3WX1jOXKPdecidvWCcRrm8GPz63S_yUNJGai-r96ClLaSvg8pQKFNmXdHm4bFmg9pACYowyh1cTJIbkKk8AMoLpzqFbLMv0PX_u7)
