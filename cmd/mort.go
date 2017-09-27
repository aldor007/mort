package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
	"mort"
	"mort/config"
	"mort/object"

	"mort/response"
)

func main() {
	imgConfig := config.GetInstance()
	imgConfig.Init("configuration/config.yml")
	// Echo instance
	e := echo.New()

	// Route => handler
	e.GET("/*", func(ctx echo.Context) error {
		obj, err := object.NewFileObject(ctx.Request().URL.Path)
		if err != nil {
			return ctx.NoContent(400)
		}
		// dodac placeholder
		res := mort.Process(obj)
		res.WriteHeaders(ctx.Response())

		return ctx.Stream(res.StatusCode, res.Headers[response.ContentType], res.Stream)
	})

	s := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	}

	e.Logger.Fatal(e.StartServer(s))

}
