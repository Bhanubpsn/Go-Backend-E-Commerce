package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Payload struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func SendEmail(to string, name string) error {
	from := os.Getenv("EMAIL")
	password := os.Getenv("PASSWORD")

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	auth := smtp.PlainAuth("", from, password, smtpHost)

	subject := "Welcome!"
	body := fmt.Sprintf("Hello %s,\n\nWelcome to our service!", name)

	message := []byte(
		"Subject: " + subject + "\r\n" +
			"\r\n" +
			body + "\r\n")

	err := smtp.SendMail(
		smtpHost+":"+smtpPort,
		auth,
		from,
		[]string{to},
		message,
	)

	if err != nil {
		log.Println("Error sending email: ", err)
		return err
	}

	fmt.Println("Email Sent Successfully")
	log.Println("Email Sent to: ", to, name)
	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	PORT := os.Getenv("PORT")
	for {
		conn, _ := net.Dial("tcp", "localhost:"+PORT)
		fmt.Fprintln(conn, "POP")

		scanner := bufio.NewScanner(conn)
		if scanner.Scan() {
			msg := scanner.Text()
			if msg != "EMPTY" {
				var data Payload
				json.Unmarshal([]byte(msg), &data)
				log.Printf("Worker: Sending email to %s", data.Email)
				SendEmail(data.Email, data.Name)
			}
		}
		conn.Close()
		time.Sleep(1 * time.Second) // Poll every second
	}
}
