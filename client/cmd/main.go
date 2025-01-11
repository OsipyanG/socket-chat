package main

import (
	"bufio"
	"client/internal/config"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	cfg := config.MustLoad()
	conn, err := net.Dial("tcp", net.JoinHostPort(cfg.Host, cfg.Port))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	go handleExit(conn)

	go readMessages(conn)

	writeMessages(conn)
}

func handleExit(conn net.Conn) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\nExiting chat...")
	conn.Close()
	os.Exit(0)
}

func readMessages(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("\nServer disconnected.")
				os.Exit(0)
			}
			log.Printf("Error reading from server: %v", err)

			continue
		}
		message = strings.TrimSpace(message)

		fmt.Printf("\r%s\n>> ", message)
	}
}

func writeMessages(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">> ")
		message, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("\nExiting chat...")

				return
			}
			log.Printf("Error reading input: %v", err)

			continue
		}

		message = strings.TrimSpace(message)
		if message == "" {
			continue
		}

		_, err = conn.Write([]byte(message + "\n"))
		if err != nil {
			log.Printf("Error sending message: %v", err)

			return
		}
	}
}
