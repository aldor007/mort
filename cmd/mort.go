package main

import (
	"net/http"
	"time"
	"flag"
	"fmt"

	"github.com/labstack/echo"
	"mort"
	"mort/config"
	"mort/object"

)

func main() {
	configPath := flag.String("config", "configuration/config.yml", "Path to configuration")
	listenAddr := flag.String("listen", ":8080", "Listen addr")
	flag.Parse()
	fmt.Println(*configPath, *listenAddr)

	imgConfig := config.GetInstance()
	imgConfig.Load(*configPath)
	// Echo instance
	e := echo.New()

	e.Use(mort.S3AuthMiddleware(imgConfig))
	e.Use(mort.S3Middleware(imgConfig))

	// Route => handler
	e.Any ("/*", func(ctx echo.Context) error {
		obj, err := object.NewFileObject(ctx.Request().URL.Path, imgConfig)
		if err != nil {
			return ctx.NoContent(400)
		}
		// dodac placeholder
		res := mort.Process(ctx, obj)
		res.WriteHeaders(ctx.Response())
		defer res.Close()

		return ctx.Stream(res.StatusCode, res.ContentType, res.Stream)
	})


	s := &http.Server{
		Addr:        *listenAddr,
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 2 * time.Minute,
	}

	e.Logger.Fatal(e.StartServer(s))

}
