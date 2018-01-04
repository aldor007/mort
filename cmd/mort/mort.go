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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
)

const (
	// Version of mort
	Version = "0.5.0"
	// BANNER just fancy command line banner
	BANNER = `
  /\/\   ___  _ __| |_
 /    \ / _ \| '__| __|
/ /\/\ \ (_) | |  | |_
\/    \/\___/|_|   \__|
 Version: %s
`
)


func debugListener(mortConfig *config.Config) (s *http.Server, ln net.Listener, socketPath string) {
	router := chi.NewRouter()
	router.Mount("/debug", middleware.Profiler())
	router.HandleFunc("/metrics", promhttp.Handler())
	s = &http.Server{
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 2 * time.Minute,
		Handler:      router,
	}

	network := "tcp"
	address := mortConfig.Server.InternalListener
	socketPath = ""
	if strings.HasPrefix(address, "unix:") {
		network = "unix"
		socketPath = address
		address = strings.Replace(address, "unix:", "", 1)
	}

	ln, err := net.Listen(network, address)
	if err != nil {
		panic(err)
	}

	return
}

func handleSignals(servers []*http.Server, socketPaths []string, wg *sync.WaitGroup) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGUSR2, syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM, os.Kill)
	imgConfig := config.GetInstance()
	for {
		sig := <-signalChan
		switch sig {
		case syscall.SIGTERM:
		case syscall.SIGKILL:
		case syscall.SIGINT:
		case os.Kill:
			for _, s := range servers {
				s.Close()
				wg.Done()
			}

			for _, socketPath := range socketPaths {
				os.Remove(socketPath)
			}
			wg.Done()
			return
		default:
		}
	}
}

func startServer(s *http.Server, ln net.Listener) {
	err := s.Serve(ln)
	if err != nil {
		fmt.Println("Error listen", err)
		panic(err)
	}
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
				response.NewError(400, err).SetDebug(debug, nil).Send(resWriter)
				return
			}

			res := rp.Process(req, obj)
			res.SetDebug(debug, obj)
			if debug {
				res.Set("X-Mort-Version", Version)
			}

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

	serversCount := len(imgConfig.Server.Listen) + 1
	servers := make([]*http.Server, serversCount)
	netListeners := make([]net.Listener, serversCount)
	var socketPaths []string

	for i, l := range imgConfig.Server.Listen {
		servers[i] = &http.Server{
			ReadTimeout:  2 * time.Minute,
			WriteTimeout: 2 * time.Minute,
			Handler:      router,
		}

		network := "tcp"
		address := l
		if strings.HasPrefix(l, "unix:") {
			network = "unix"
			address = strings.Replace(l, "unix:", "", 1)
			socketPaths = append(socketPaths, address)
		}

		ln, err := net.Listen(network, address)
		if err != nil {
			panic(err)
		}
		netListeners[i] = ln
	}

	var internalSocketPath string
	servers[serversCount - 1], netListeners[serversCount - 1], internalSocketPath = debugListener(imgConfig)
	if internalSocketPath != "" {
		socketPaths = append(socketPaths, internalSocketPath)
	}

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

	var wg sync.WaitGroup

	wg.Add(1)
	go handleSignals(servers, socketPaths, &wg)

	for i, s := range servers {
		wg.Add(1)
		go startServer(s, netListeners[i])
	}

	wg.Wait()
	fmt.Println("Bye...")
}
