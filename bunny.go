package bunny

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

// LogDetails represents the structure of the error message to be sent
type LogDetails struct {
	Level     string                 `json:"level"`
	Data      map[string]interface{} `json:"data"`
	Timestamp string                 `json:"timestamp"`
	AppId     string                 `json:"appId"`
	StreamId  string                 `json:"streamId"`
}

type LogBunnyLogger struct {
	AppId    string
	StreamId string
}

// LogHandler middleware catches all errors and sends them to the specified URL
func (logger *LogBunnyLogger) LogHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}
	details := LogDetails{
		Level: "error",
		Data: map[string]interface{}{
			"error":  err.Error(),
			"status": code,
		},
		Timestamp: time.Now().GoString(),
		AppId:     logger.AppId,
		StreamId:  logger.StreamId,
	}
	fmt.Println(err)
	go sendDetails(details)
	return nil
}

// sendDetails sends the error details to the specified external server
func sendDetails(details LogDetails) {
	url := "https://sabertooth.fly.dev/ingest"
	jsonData, err := json.Marshal(details)
	if err != nil {
		log.Printf("Failed to marshal error details: %v", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to send error details: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Printf("Failed to send error details, status code: %d", resp.StatusCode)
	}
}

/*
func main() {
	logger := LogBunnyLogger{AppId: "2345", StreamId: "12345"}
	app := fiber.New(fiber.Config{
		ErrorHandler: logger.LogHandler,
	})

	// Define a route for the "Hello, World!" endpoint
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// Start the server on port 3000
	app.Listen(":3000")
}
*/
