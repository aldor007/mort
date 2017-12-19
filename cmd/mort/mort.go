package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/lock"
	"github.com/aldor007/mort/pkg/log"
	mortMiddleware "github.com/aldor007/mort/pkg/middleware"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/processor"
	"github.com/aldor007/mort/pkg/response"
	"github.com/aldor007/mort/pkg/throttler"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

const (
	// Version of mort
	Version = "0.2.1"
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

func debugListener(mortConfig *config.Config) {
	if debugServer != nil {
		debugServer.Close()
		return
	}

	router := chi.NewRouter()
	router.Mount("/debug", middleware.Profiler())
	s := &http.Server{
		Addr:         mortConfig.Server.DebugListen,
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 2 * time.Minute,
		Handler:      router,
	}

	debugServer = s
	s.ListenAndServe()
}

func main() {
	configPath := flag.String("config", "/etc/mort/mort.yml", "Path to configuration")
	debug := flag.Bool("debug", false, "enable debug mode")
	flag.Parse()

	logger, _ := zap.NewProduction()
	//logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	log.RegisterLogger(logger)
	router := chi.NewRouter()
	imgConfig := config.GetInstance()
	err := imgConfig.Load(*configPath)

	if err != nil {
		panic(err)
	}

	fmt.Printf(BANNER, "v"+Version)
	fmt.Printf("Config file %s listen addr %s debug: %t pid: %d \n", *configPath, imgConfig.Server.Listen, *debug, os.Getpid())

	rp := processor.NewRequestProcessor(imgConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(10))
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

			res.SendContent(req, resWriter)
		})
	})

	router.HandleFunc("/", http.HandlerFunc(func(resWriter http.ResponseWriter, req *http.Request) {
		resWriter.WriteHeader(400)
		log.Log().Warn("Mort error request shouldn't go here")
	}))

	s := &http.Server{
		Addr:         imgConfig.Server.Listen,
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 2 * time.Minute,
		Handler:      router,
	}

	if debug != nil && *debug {
		go debugListener(imgConfig)
	}

	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan, syscall.SIGUSR2)

	go func() {
		for {
			// FIXME: move it to prometheus
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			log.Log().Info("Runtime stats", zap.Uint64("alloc", m.Alloc/1024), zap.Uint64("total-alloc", m.TotalAlloc/1024),
				zap.Uint64("sys", m.Sys/1021), zap.Uint32("numGC", m.NumGC), zap.Uint64("last-gc-pause", m.PauseNs[(m.NumGC+255)%256]))
			time.Sleep(300 * time.Second)
		}
	}()

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
				go debugListener(imgConfig)
			default:
			}
		}
	}()

	s.ListenAndServe()

}
