package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	"github.com/aldor007/mort/config"
	"github.com/aldor007/mort/lock"
	"github.com/aldor007/mort/log"
	mortMiddleware "github.com/aldor007/mort/middleware"
	"github.com/aldor007/mort/object"
	"github.com/aldor007/mort/processor"
	"github.com/aldor007/mort/response"
	"github.com/aldor007/mort/throttler"
)

const (
	// Version of mort
	Version = "0.0.1"
	// BANNER just fancy command line banner
	BANNER = `
  /\/\   ___  _ __| |_
 /    \ / _ \| '__| __|
/ /\/\ \ (_) | |  | |_
\/    \/\___/|_|   \__|
 Version: %s
`
)

func main() {
	configPath := flag.String("config", "configuration/config.yml", "Path to configuration")
	listenAddr := flag.String("listen", ":8080", "Listen addr")
	flag.Parse()

	fmt.Printf(BANNER, "v"+Version)
	fmt.Printf("Config file %s listen addr %s\n", *configPath, *listenAddr)

	logger, _ := zap.NewProduction()
	//logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	log.RegisterLogger(logger)
	router := chi.NewRouter()
	rp := processor.NewRequestProcessor(5, lock.NewMemoryLock(), throttler.NewBucketThrottler(10))

	imgConfig := config.GetInstance()
	imgConfig.Load(*configPath)

	s3Auth := mortMiddleware.NewS3AuthMiddleware(imgConfig)

	router.Use(s3Auth.Handler)

	router.Use(func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(resWriter http.ResponseWriter, req *http.Request) {
			debug := req.Header.Get("X-Mort-Debug") != ""
			obj, err := object.NewFileObject(req.URL.Path, imgConfig)
			if err != nil {
				logger.Sugar().Errorf("Unable to create file object err = %s", err)
				response.NewError(400, err).SetDebug(debug).Send(resWriter)
				return
			}

			res := rp.Process(req, obj)
			res.SetDebug(debug)
			// FIXME
			res.Set("Access-Control-Allow-Headers", "Content-Type, X-Amz-Public-Width, X-Amz-Public-Height")
			res.Set("Access-Control-Expose-Headers", "Content-Type, X-Amz-Public-Width, X-Amz-Public-Height")
			res.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, HEAD")
			defer logger.Sync() // flushes buffer, if any
			if res.HasError() {
				log.Log().Warn("github.com/aldor007/mort process error", zap.String("obj.Key", obj.Key), zap.Error(res.Error()))
			}

			res.Send(resWriter)
		})
	})

	router.HandleFunc("/", http.HandlerFunc(func(resWriter http.ResponseWriter, req *http.Request) {
		resWriter.WriteHeader(400)
		log.Log().Warn("github.com/aldor007/mort error request shouldn't go here")
	}))

	s := &http.Server{
		Addr:         *listenAddr,
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 2 * time.Minute,
		Handler:      router,
	}

	s.ListenAndServe()

}
