package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Broker struct {
	queue []string
	mu    sync.Mutex
}

func (b *Broker) Push(msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.queue = append(b.queue, msg)
}

func (b *Broker) Pop() (string, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.queue) == 0 {
		return "", false
	}
	msg := b.queue[0]
	b.queue = b.queue[1:]
	return msg, true
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	PORT := os.Getenv("PORT")
	broker := &Broker{}
	ln, _ := net.Listen("tcp", ":"+PORT)
	fmt.Println("Custom Message Broker running on :" + PORT)

	for {
		conn, _ := ln.Accept() // For hadling multiple clients concurrently using goroutines
		go handleConnection(conn, broker)
	}
}

func handleConnection(conn net.Conn, b *Broker) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		cmd := scanner.Text()
		if cmd == "POP" {
			if msg, ok := b.Pop(); ok {
				fmt.Fprintln(conn, msg)
			} else {
				fmt.Fprintln(conn, "EMPTY")
			}
		} else {
			// Assume any other text is a JSON payload to PUSH
			b.Push(cmd)
			fmt.Fprintln(conn, "ACK")
		}
	}
}
