# GO SDK for logbunny
- Currently supports go-fiber

### Installation:
- Install the SDK using `go get github.com/logbunny/bunny.go`
- Initialize the Logger with your `app id` and `stream id`
- Connect the custom error handler into go-fiber's custom error handler
### Example:
```go
func main() {
  logger := LogBunnyLogger{AppId: "your-app-id", StreamId: "your-stream-id"}
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
```
