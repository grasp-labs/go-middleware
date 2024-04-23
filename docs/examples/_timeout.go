package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Create server
	e := echo.New()
	e.Use(middleware.ContextTimeout(2 * time.Second))

	e.GET("/", func(c echo.Context) error {
		if err := sleepWithContext(c.Request().Context(), 3*time.Second); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, "hello world")
	})

	if err := e.Start(":8080"); err != http.ErrServerClosed {
		log.Fatal(err)
	}

}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)

	defer func() {
		_ = timer.Stop()
	}()

	select { // wait for task to finish or context to timeout/cancelled
	case <-ctx.Done():
		return context.DeadlineExceeded
	case <-timer.C:
		return nil
	}
}
