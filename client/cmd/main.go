package main

import (
	"bufio"
	"client/internal/config"
	"context"
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

type ChatClient struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}

func main() {
	cfg := config.MustLoad()

	client, err := NewChatClient(cfg.Host, cfg.Port)
	if err != nil {
		log.Fatalf("Failed to start client: %v", err)
	}
	defer client.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	client.Run(ctx)
}

func NewChatClient(host, port string) (*ChatClient, error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	return &ChatClient{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}, nil
}

func (c *ChatClient) Run(ctx context.Context) {
	go c.readMessages(ctx)
	c.writeMessages(ctx)
}

func (c *ChatClient) Close() {
	if err := c.conn.Close(); err != nil {
		log.Printf("Failed to close connection: %v", err)
	}
}

func (c *ChatClient) readMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nExiting chat (read) ...")
			return
		default:
			message, err := c.reader.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					fmt.Println("\nServer disconnected.")
					return
				}
				log.Printf("Error reading from server: %v", err)
				continue
			}
			message = strings.TrimSpace(message)
			fmt.Printf("\r%s\n>> ", message)
		}
	}
}

func (c *ChatClient) writeMessages(ctx context.Context) {
	reader := bufio.NewReader(os.Stdin)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nExiting chat (write) ...")
			return
		default:
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

			_, err = c.writer.WriteString(message + "\n")
			if err != nil {
				log.Printf("Error sending message: %v", err)
				return
			}

			if err := c.writer.Flush(); err != nil {
				log.Printf("Error flushing message: %v", err)
				return
			}
		}
	}
}
