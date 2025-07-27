package main

import (
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/hugmouse/scan24/internal/handler"
	"github.com/hugmouse/scan24/internal/job"
	"github.com/hugmouse/scan24/internal/workerPool"
	"github.com/hugmouse/scan24/static"
	"log"
	"net"
	"net/http"
	"time"
)

type config struct {
	HTTPServe                   string `env:"HTTP_SERVE"                      envDefault:":8080"`
	HTTPClientTimeout           int    `env:"HTTP_CLIENT_TIMEOUT"             envDefault:"5"`
	HTTPServerReadHeaderTimeout int    `env:"HTTP_SERVER_READ_HEADER_TIMEOUT" envDefault:"3"`
	DialTimeout                 int    `env:"DIAL_TIMEOUT"                    envDefault:"30"`
	DialKeepAlive               int    `env:"DIAL_KEEP_ALIVE"                 envDefault:"30"`
	TLSHandshakeTimeout         int    `env:"TLS_HANDSHAKE_TIMEOUT"           envDefault:"10"`
	ResponseHeaderTimeout       int    `env:"RESPONSE_HEADER_TIMEOUT"         envDefault:"10"`
	ExpectContinueTimeout       int    `env:"EXPECT_CONTINUE_TIMEOUT"         envDefault:"1"`
	MaxIdleConns                int    `env:"MAX_IDLE_CONNS"                  envDefault:"100"`
	IdleConnTimeout             int    `env:"IDLE_CONN_TIMEOUT"               envDefault:"90"`
	MaxRedirects                int    `env:"MAX_REDIRECTS"                   envDefault:"3"`
	RateLimit                   int    `env:"RATE_LIMIT"                      envDefault:"2"`
	WorkerPool                  int    `env:"WORKER_POOL" envDefault:"5"`
	QueueCapacity               int    `env:"QUEUE_CAPACITY" envDefault:"100"`
}

func main() {
	cfg, err := env.ParseAs[config]()
	if err != nil {
		log.Fatal(err)
	}

	wp := workerPool.NewWorkerPool(cfg.WorkerPool, cfg.QueueCapacity)
	wp.StartWorkers()

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(cfg.DialTimeout) * time.Second,
			KeepAlive: time.Duration(cfg.DialKeepAlive) * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   time.Duration(cfg.TLSHandshakeTimeout) * time.Second,
		ResponseHeaderTimeout: time.Duration(cfg.ResponseHeaderTimeout) * time.Second,
		ExpectContinueTimeout: time.Duration(cfg.ExpectContinueTimeout) * time.Second,
		MaxIdleConns:          cfg.MaxIdleConns,
		IdleConnTimeout:       time.Duration(cfg.IdleConnTimeout) * time.Second,
		ForceAttemptHTTP2:     true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(cfg.HTTPClientTimeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= cfg.MaxRedirects {
				return http.ErrUseLastResponse
			}

			return nil
		},
		Jar: nil,
	}

	h := &handler.Handler{
		Client:     client,
		RateLimit:  cfg.RateLimit,
		WorkerPool: wp,
		JobResults: make(map[string]job.Result),
	}

	go func() {
		for result := range wp.Results() {
			if result.Error != nil {
				fmt.Printf("Job %s (%s) by Worker %d FAILED: %v (Result Data: %v)\n",
					result.JobID, result.JobType, result.WorkerID, result.Error, result.Data)
			} else {
				fmt.Printf("Job %s (%s) by Worker %d COMPLETED: %v\n",
					result.JobID, result.JobType, result.WorkerID, result.Data)
			}
			h.JobResults[result.JobID] = result
		}
	}()

	go func() {
		for {
			time.Sleep(time.Duration(1 * time.Second))
			total, budy, free := wp.GetStatus()
			log.Printf("[worker stats] Total %d, busy %d, free %d", total, budy, free)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.IndexHandler)
	mux.HandleFunc("/analyze", h.AnalyzeHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.FS))))

	log.Printf("Starting server on %s", cfg.HTTPServe)

	server := &http.Server{
		Addr:              cfg.HTTPServe,
		ReadHeaderTimeout: time.Duration(cfg.HTTPServerReadHeaderTimeout) * time.Second,
		Handler:           mux,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
