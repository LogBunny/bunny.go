package bunny

import (
	"bytes"
	"encoding/json"
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
func (logger *LogBunnyLogger) LogHandler(c *fiber.Ctx) error {
	defer func() {
		if r := recover(); r != nil {
			var errMessage string
			switch r := r.(type) {
			case string:
				errMessage = r
			case error:
				errMessage = r.Error()
			default:
				errMessage = "unknown error"
			}

			status := fiber.StatusInternalServerError

			errorDetails := LogDetails{
				Level: "error",
				Data: map[string]interface{}{
					"error":  errMessage,
					"path":   c.Path(),
					"status": status,
				},
				Timestamp: time.Now().Format(time.RFC3339),
				AppId:     logger.AppId,
				StreamId:  logger.StreamId,
			}

			// Log error details
			log.Printf("Error: %s, Path: %s, Status: %d\n", errMessage, c.Path(), status)

			// Send error details to external server using goroutine
			go sendDetails(errorDetails)

			// Respond with an error message
			c.Status(status).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}
	}()

	err := c.Next()

	if err != nil {
		status := c.Response().StatusCode()
		if status == 0 {
			status = fiber.StatusInternalServerError
		}

		errorDetails := LogDetails{
			Level: "error",
			Data: map[string]interface{}{
				"error":  err.Error(),
				"path":   c.Path(),
				"status": status,
			},
			Timestamp: time.Now().Format(time.RFC3339),
			AppId:     logger.AppId,
			StreamId:  logger.StreamId,
		}

		// Log error details
		log.Printf("Error: %s, Path: %s, Status: %d\n", err.Error(), c.Path(), status)

		// Send error details to external server using goroutine
		go sendDetails(errorDetails)

		// Respond with the appropriate status and error message
		return c.Status(status).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Check for 4xx and 5xx status codes
	status := c.Response().StatusCode()
	if status >= 400 {
		errorDetails := LogDetails{
			Level: "error",
			Data: map[string]interface{}{
				"error":  http.StatusText(status),
				"path":   c.Path(),
				"status": status,
			},
			Timestamp: time.Now().Format(time.RFC3339),
			AppId:     logger.AppId,
			StreamId:  logger.StreamId,
		}

		// Log error details
		log.Printf("Error: %s, Path: %s, Status: %d\n", http.StatusText(status), c.Path(), status)

		// Send error details to external server using goroutine
		go sendDetails(errorDetails)
	}

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
