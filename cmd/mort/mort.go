package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	mortMiddleware "github.com/aldor007/mort/pkg/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	"go.uber.org/zap"

	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/lock"
	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/processor"
	"github.com/aldor007/mort/pkg/response"
	"github.com/aldor007/mort/pkg/storage"
	"github.com/aldor007/mort/pkg/throttler"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap/zapcore"

	"github.com/aldor007/mort/pkg/object/cloudinary"
	_ "github.com/aldor007/mort/pkg/object/cloudinary"
)

// variables provided by goreleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

const (
	// BANNER just fancy command line banner
	BANNER = `
  /\/\   ___  _ __| |_
 /    \ / _ \| '__| __|
/ /\/\ \ (_) | |  | |_
\/    \/\___/|_|   \__|
 Version: %s
 Commit: %s
 Date: %s
`
)

func debugListener(mortConfig *config.Config) (s *http.Server, ln net.Listener, socketPath string) {
	router := chi.NewRouter()
	router.Mount("/debug", middleware.Profiler())
	router.Handle("/metrics", promhttp.Handler())
	s = &http.Server{
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 2 * time.Minute,
		Handler:      router,
	}

	network := "tcp"
	address := mortConfig.Server.InternalListen
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

func handleSignals(rp *processor.RequestProcessor, servers []*http.Server, socketPaths []string, wg *sync.WaitGroup) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGUSR2, syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM, os.Kill)
	for {
		sig := <-signalChan
		switch sig {
		case os.Kill, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT:
			// Gracefully shutdown request processor first
			if rp != nil {
				rp.Shutdown()
			}

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

func configureMonitoring(mortConfig *config.Config) {
	var logCfg zap.Config
	if mortConfig.Server.LogLevel == "debug" {
		logCfg = zap.NewDevelopmentConfig()
	} else {
		logCfg = zap.NewProductionConfig()
	}

	logCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, _ := logCfg.Build()

	host, err := os.Hostname()
	if err != nil {
		host = "unknown"
	}

	pid := os.Getpid()
	logger = logger.With(
		zap.String("hostname", host),
		zap.Int("pid", pid),
	)

	zap.ReplaceGlobals(logger)
	monitoring.RegisterLogger(logger)
	if mortConfig.Server.Monitoring == "prometheus" {
		p := monitoring.NewPrometheusReporter()
		p.RegisterCounterVec("cache_ratio", prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "mort_cache_ratio",
			Help: "mort cache ratio",
		},
			[]string{"status"},
		))

		p.RegisterCounter("throttled_count", prometheus.NewCounter(prometheus.CounterOpts{
			Name: "mort_request_throttled_count",
			Help: "mort count of throttled requests",
		}))

		p.RegisterGaugeVec("storage_throughput", prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "mort_storage_throughput",
			Help: "mort requests storage",
		},
			[]string{"method", "storage"},
		))

		p.RegisterCounterVec("storage_request", prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "mort_storage_request",
			Help: "mort requests storage",
		},
			[]string{"method", "bucket", "storage", "object_type"},
		))

		p.RegisterCounter("collapsed_count", prometheus.NewCounter(prometheus.CounterOpts{
			Name: "mort_request_collapsed_count",
			Help: "mort count of collapsed requests",
		}))

		p.RegisterCounter("vips_cleanup_count", prometheus.NewCounter(prometheus.CounterOpts{
			Name: "mort_vips_cleanup_count",
			Help: "mort count of vips cache cleanups",
		}))

		p.RegisterCounter("glacier_error_detected", prometheus.NewCounter(prometheus.CounterOpts{
			Name: "mort_glacier_error_detected",
			Help: "mort count of GLACIER/DEEP_ARCHIVE errors detected",
		}))

		p.RegisterCounter("glacier_restore_initiated", prometheus.NewCounter(prometheus.CounterOpts{
			Name: "mort_glacier_restore_initiated",
			Help: "mort count of GLACIER restore requests initiated",
		}))

		p.RegisterHistogramVec("storage_time", prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "mort_storage_time",
			Help:    "mort storage times",
			Buckets: []float64{10.0, 50.0, 100.0, 200.0, 300.0, 400.0, 500., 1000., 2000., 3000., 4000., 5000., 6000., 10000., 30000., 60000., 70000., 80000.},
		},
			[]string{"method", "storage"},
		))

		p.RegisterHistogramVec("response_time", prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "mort_response_time",
			Help:    "mort response times",
			Buckets: []float64{10.0, 50.0, 100.0, 200.0, 300.0, 400.0, 500., 1000., 2000., 3000., 4000., 5000., 6000., 10000., 30000., 60000., 70000., 80000.},
		},
			[]string{"method"},
		))

		p.RegisterCounterVec("request_type", prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "mort_request_type_count",
			Help: "mort count of given request type",
		},
			[]string{"type"},
		))

		p.RegisterHistogram("generation_time", prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "mort_generation_time",
			Help:    "mort generation times",
			Buckets: []float64{10.0, 50.0, 100.0, 200.0, 300.0, 400.0, 500., 1000., 2000., 3000., 4000., 5000., 6000., 10000., 30000., 60000., 70000., 80000.},
		}))

		monitoring.RegisterReporter(p)
	}
}

func startServer(s *http.Server, ln net.Listener) {
	err := s.Serve(ln)
	if err != nil && err != http.ErrServerClosed {
		fmt.Println("Error listen", err)
	}
	fmt.Println("Server listen", ln.Addr().String())
}

func main() {
	configPath := flag.String("config", "/etc/mort/mort.yml", "Path to configuration")
	versionCmd := flag.Bool("version", false, "get mort version")
	flag.Parse()

	if versionCmd != nil && *versionCmd == true {
		fmt.Println(version, "commit", commit, "date", date)
		return
	}

	router := chi.NewRouter()
	imgConfig := config.GetInstance()
	err := imgConfig.Load(*configPath)
	configureMonitoring(imgConfig)

	if err != nil {
		panic(err)
	}

	fmt.Printf(BANNER, version, commit, date)
	fmt.Printf("Config file %s listen addr %s montoring: and debug listen %s pid: %d \n", *configPath, imgConfig.Server.Listen, imgConfig.Server.InternalListen, os.Getpid())

	// Set default concurrent image processing limit if not configured
	concurrentLimit := imgConfig.Server.ConcurrentImageProcessing
	if concurrentLimit <= 0 {
		concurrentLimit = 100
	}
	monitoring.Log().Info("Image processing concurrency", zap.Int("limit", concurrentLimit))

	rp := processor.NewRequestProcessor(imgConfig.Server, lock.Create(imgConfig.Server.Lock, imgConfig.Server.LockTimeout), throttler.NewBucketThrottler(concurrentLimit))

	// Initialize archive restore cache at startup to prevent lazy initialization panics
	// This ensures the cache is ready before any requests arrive
	_ = storage.GetRestoreCache(imgConfig.Server.Cache)
	monitoring.Log().Info("Archive restore cache initialized")

	if imgConfig.Server.AccessLog {
		logger := httplog.NewLogger("mort-access", httplog.Options{
			JSON: true,
			Tags: map[string]string{
				"version": version,
				"logType": "access", // For filtering access logs vs application logs
			},
			LogLevel: "info",
		})
		router.Use(httplog.RequestLogger(logger))

	}
	cloudinaryUploadInterceptor := cloudinary.NewUploadInterceptorMiddleware(imgConfig)
	router.Use(cloudinaryUploadInterceptor.Handler)

	s3Auth := mortMiddleware.NewS3AuthMiddleware(imgConfig)
	router.Use(s3Auth.Handler)

	router.Use(func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(resWriter http.ResponseWriter, req *http.Request) {
			metric := "response_time;method:" + req.Method
			t := monitoring.Report().Timer(metric)
			defer t.Done()
			debug := req.Header.Get("X-Mort-Debug") != ""
			obj, err := object.NewFileObject(req.URL, imgConfig)
			if err != nil {
				monitoring.Logs().Errorf("Unable to create file object err = %s", err)
				response.NewError(400, err).SetDebug(&object.FileObject{Debug: debug}).Send(resWriter)
				return
			}
			obj.Debug = debug

			res := rp.Process(req, obj)
			defer res.Close()
			res.SetDebug(obj)
			if debug {
				res.Set("X-Mort-Version", version)
				monitoring.Log().Info("Mort processing object", obj.LogData()...)
			}

			// FIXME
			res.Set("Access-Control-Allow-Headers", "Content-Type, X-Amz-Public-Width, X-Amz-Public-Height")
			res.Set("Access-Control-Expose-Headers", "Content-Type, X-Amz-Public-Width, X-Amz-Public-Height")
			res.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, HEAD")
			res.Set("Access-Control-Allow-Origin", "*")
			res.Set("Accept-Ranges", "bytes")

			if res.HasError() {
				monitoring.Log().Error("Mort process error", obj.LogData(zap.Int("res.status", res.StatusCode), zap.String("req.url", req.URL.String()), zap.Error(res.Error()))...)
			}

			res.SendContent(req, resWriter)
		})
	})

	router.HandleFunc("/", http.HandlerFunc(func(resWriter http.ResponseWriter, req *http.Request) {
		resWriter.WriteHeader(400)
		monitoring.Log().Warn("Mort error request shouldn't go here")
	}))

	serversCount := len(imgConfig.Server.Listen) + 1
	servers := make([]*http.Server, serversCount)
	netListeners := make([]net.Listener, serversCount)
	var socketPaths []string

	for i, l := range imgConfig.Server.Listen {
		servers[i] = &http.Server{
			ReadTimeout:  2 * time.Hour, // allow to upload big files
			WriteTimeout: 2 * time.Hour, // Write timeout is for while respone write (not reset after each write like in nginx)
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
	servers[serversCount-1], netListeners[serversCount-1], internalSocketPath = debugListener(imgConfig)
	if internalSocketPath != "" {
		socketPaths = append(socketPaths, internalSocketPath)
	}

	var wg sync.WaitGroup

	monitoring.Logs().Sync()
	wg.Add(1)
	go handleSignals(&rp, servers, socketPaths, &wg)

	for i, s := range servers {
		wg.Add(1)
		go startServer(s, netListeners[i])
	}

	wg.Wait()
	defer monitoring.Log().Sync() // flushes buffer, if any
	fmt.Println("Bye...")
}
