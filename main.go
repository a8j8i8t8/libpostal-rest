package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	expand "github.com/openvenues/gopostal/expand"
	parser "github.com/openvenues/gopostal/parser"

)

var (

	parse_duration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "libpostal_parse_request_duration_seconds",
		Help:    "Histogram of the /parser request duration.",
		Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	})

	parse_counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "libpostal_parse_requests_total",
			Help: "Total number of /parser requests.",
		},
		[]string{"status"},
	)

	expand_duration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "libpostal_expand_request_duration_seconds",
		Help:    "Histogram of the /expand request duration.",
		Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	})

	expand_counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "libpostal_expand_requests_total",
			Help: "Total number of /expand requests.",
		},
		[]string{"status"},
	)
)

// init registers Prometheus metrics.
func init() {
	prometheus.MustRegister(parse_duration)
	prometheus.MustRegister(parse_counter)
	prometheus.MustRegister(expand_duration)
	prometheus.MustRegister(expand_counter)
}

type Request struct {
	Query string `json:"query"`
}

func main() {
	host := os.Getenv("LISTEN_HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	port := os.Getenv("LISTEN_PORT")
	if port == "" {
		port = "8080"
	}
	listenSpec := fmt.Sprintf("%s:%s", host, port)

	certFile := os.Getenv("SSL_CERT_FILE")
	keyFile := os.Getenv("SSL_KEY_FILE")

	router := mux.NewRouter()
	router.HandleFunc("/health", HealthHandler).Methods("GET")
	router.HandleFunc("/expand", ExpandHandler).Methods("POST")
	router.HandleFunc("/parser", ParserHandler).Methods("POST")
	// router.Path("/metrics").Handler(promhttp.Handler())
	router.Handle("/metrics", promhttp.Handler())

	s := &http.Server{Addr: listenSpec, Handler: router}
	go func() {
		if certFile != "" && keyFile != "" {
			fmt.Printf("listening on https://%s\n", listenSpec)
			s.ListenAndServeTLS(certFile, keyFile)
		} else {
			fmt.Printf("listening on http://%s\n", listenSpec)
			s.ListenAndServe()
		}
	}()

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	<-stop
	fmt.Println("\nShutting down the server...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	s.Shutdown(ctx)
	fmt.Println("Server stopped")
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func ExpandHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req Request
	var status int

	defer func(begun time.Time) {
		expand_duration.Observe(time.Since(begun).Seconds())

		expand_counter.With(prometheus.Labels{
			"status": fmt.Sprint(status),
		}).Inc()
	}(time.Now())

	q, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(q, &req)

	expansions := expand.ExpandAddress(req.Query)

	expansionThing, _ := json.Marshal(expansions)
	w.Write(expansionThing)
}

func ParserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req Request
	var status int

	defer func(begun time.Time) {
		parse_duration.Observe(time.Since(begun).Seconds())

		parse_counter.With(prometheus.Labels{
			"status": fmt.Sprint(status),
		}).Inc()
	}(time.Now())

	q, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(q, &req)

	parsed := parser.ParseAddress(req.Query)
	parseThing, _ := json.Marshal(parsed)
	w.Write(parseThing)
}
