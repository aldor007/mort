package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"github.com/aldor007/mort/config"
	"github.com/aldor007/mort/lock"
	"github.com/aldor007/mort/log"
	mortMiddleware "github.com/aldor007/mort/middleware"
	"github.com/aldor007/mort/object"
	"github.com/aldor007/mort/processor"
	"github.com/aldor007/mort/response"
	"github.com/aldor007/mort/throttler"
	"os"
	"os/signal"
	"syscall"
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

var debugServer *http.Server

func debugListener() {
	if debugServer != nil {
		debugServer.Close()
		return
	}

	router := chi.NewRouter()
	router.Mount("/", middleware.Profiler())
	s := &http.Server{
		Addr:         "localhost:8081",
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 2 * time.Minute,
		Handler:      router,
	}

	debugServer = s
	s.ListenAndServe()
}

func main() {
	configPath := flag.String("config", "configuration/config.yml", "Path to configuration")
	listenAddr := flag.String("listen", ":8080", "Listen addr")
	debug := flag.Bool("debug", false, "enable debug mode")
	flag.Parse()

	fmt.Printf(BANNER, "v"+Version)
	fmt.Printf("Config file %s listen addr %s debug: %t pid: %d \n", *configPath, *listenAddr, *debug, os.Getpid())

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
			obj, err := object.NewFileObject(req.URL, imgConfig)
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
				log.Log().Warn("Mort process error", zap.String("obj.Key", obj.Key), zap.Error(res.Error()))
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

	if debug != nil && *debug {
		go debugListener()
	}

	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan, syscall.SIGUSR2)

	go func() {
		for {
			sig := <-signal_chan
			switch sig {
			// kill -SIGHUP XXXX
			case syscall.SIGUSR2:
				if debugServer != nil {
					log.Log().Info("Stop debug server on port 8081")
				} else {
					log.Log().Info("Start debug server on port 8081")
				}
				go debugListener()
			default:
			}
		}
	}()

	s.ListenAndServe()

}
