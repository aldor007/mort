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
	"mort/log"
	"mort/lock"

	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "configuration/config.yml", "Path to configuration")
	listenAddr := flag.String("listen", ":8080", "Listen addr")
	flag.Parse()
	fmt.Println(*configPath, *listenAddr)
	logger, _ := zap.NewProduction()
	//logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	log.RegisterLogger(logger.Sugar())
	rp := mort.NewRequestProcessor(5, lock.NewMemoryLock())

	imgConfig := config.GetInstance()
	imgConfig.Load(*configPath)
	// Echo instance
	e := echo.New()

	e.Use(mort.S3AuthMiddleware(imgConfig))
	// TODO: change echo to pressly/chi


	// Route => handler
	e.Any ("/*", func(ctx echo.Context) error {
		obj, err := object.NewFileObject(ctx.Request().URL.Path, imgConfig)
		if err != nil {
			logger.Sugar().Errorf("Unable to create file object err = %s", err)
			return ctx.NoContent(400)
		}

		res := rp.Process(ctx.Request(), obj)
		if res == nil {
			logger.Sugar().Error("WTF response nil")
			return ctx.NoContent(500)

		}
		res.SetDebug(ctx.Request().Header.Get("X-Mort-Debug"))
		// FIXME
		res.Set("Access-Control-Allow-Headers", "Content-Type, X-Amz-Public-Width, X-Amz-Public-Height")
		res.Set("Access-Control-Expose-Headers", "Content-Type, X-Amz-Public-Width, X-Amz-Public-Height")
		res.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, HEAD")
		res.WriteHeaders(ctx.Response())
		defer logger.Sync() // flushes buffer, if any
		if res.HasError() {
			log.Log().Warnw("Mort process error", "obj.Key", obj.Key, "error", res.Error())
		}

		return res.Write(ctx)

	})


	s := &http.Server{
		Addr:        *listenAddr,
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 2 * time.Minute,
	}

	e.Logger.Fatal(e.StartServer(s))

}
