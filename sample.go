package main

import (
	"bufio"   // for reading input from the client
	"fmt"     // for printing messages to the console
	"net"     // for creating a TCP server
	"strings" // for string manipulation
	"sync"    // for synchronization
)

// holds the client name and message text
type Message struct {
	name, text string
}

var (
	clients  = make(map[net.Conn]string) // track connected clients
	messages = make(chan Message)        // channel for broadcast messages
	mu       sync.Mutex                  // allows for safe client map access, mutually exclusive
)

// handle client communication
func handleClient(conn net.Conn) {
	defer conn.Close()

	// ask for client's name
	fmt.Fprintln(conn, "Enter your name:")
	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	name := scanner.Text()

	// announce new client joining chat
	broadcastMessage(name, "has joined the chat")

	// store client's connection and name
	mu.Lock()
	clients[conn] = name
	mu.Unlock()

	// continuously read and broadcast client messages
	for scanner.Scan() {
		text := scanner.Text()
		if strings.ToLower(text) == "/exit" {
			broadcastMessage(name, "has left the chat")
			break
		}
		broadcastMessage(name, text)
	}

	// remove client from the list when disconnected
	mu.Lock()
	delete(clients, conn)
	mu.Unlock()
}

// broadcast message to all connected clients
func broadcastMessage(name, text string) {
	message := Message{name, text}
	messages <- message
}

// distribute messages to all clients
func distributeMessages() {
	for msg := range messages {
		mu.Lock()
		for conn := range clients {
			fmt.Fprintf(conn, "%s: %s\n", msg.name, msg.text)
		}
		mu.Unlock()
	}
}

func main() {
	// start listening for incoming TCP connections
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Chat room started. Listening on port 8080...")

	// start message distribution goroutine
	go distributeMessages()

	// accept new clients and handle each in a separate goroutine
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleClient(conn)
	}
}
