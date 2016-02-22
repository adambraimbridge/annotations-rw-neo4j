package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	"./annotations"
	"github.com/Financial-Times/base-ft-rw-app-go/baseftrwapp"
	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/http-handlers-go/httphandlers"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"github.com/jmcvetta/neoism"
	"github.com/rcrowley/go-metrics"
)

func main() {
	log.Infof("Application started with args %s", os.Args)
	app := cli.App("annotations-rw-neo4j", "A RESTful API for managing Annotations in neo4j")
	neoURL := app.StringOpt("neo-url", "http://localhost:7474/db/data", "neo4j endpoint URL")
	port := app.IntOpt("port", 8080, "Port to listen on")
	env := app.StringOpt("env", "local", "environment this app is running in")
	batchSize := app.IntOpt("batchSize", 1024, "Maximum number of statements to execute per batch")
	graphiteTCPAddress := app.StringOpt("graphiteTCPAddress", "",
		"Graphite TCP address, e.g. graphite.ft.com:2003. Leave as default if you do NOT want to output to graphite (e.g. if running locally)")
	graphitePrefix := app.StringOpt("graphitePrefix", "",
		"Prefix to use. Should start with content, include the environment, and the host name. e.g. content.test.annotation.rw.neo4j.ftaps58938-law1a-eu-t")
	logMetrics := app.BoolOpt("logMetrics", false, "Whether to log metrics. Set to true if running locally and you want metrics output")
	logLevel := app.StringOpt("log-level", "INFO", "Logging level (DEBUG, INFO, WARN, ERROR)")
	platformVersion := app.StringOpt("platformVersion", "", "Annotation source platform. Possible values are: v1 or v2.")

	app.Action = func() {
		log.Infof("annotations-rw-neo4j will listen on port: %d, connecting to: %s", *port, *neoURL)

		db, err := neoism.Connect(*neoURL)
		if err != nil {
			log.Fatalf("Error connecting to neo4j %s", err)
		}

		if *env != "local" {
			f, err := os.OpenFile("/var/log/apps/annotations-rw-neo4j-go-app.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
			if err == nil {
				log.SetOutput(f)
				log.SetFormatter(&log.TextFormatter{})
			} else {
				log.Fatalf("Failed to initialise log file, %v", err)
			}

			defer f.Close()
		}
		batchRunner := neoutils.NewBatchCypherRunner(neoutils.StringerDb{db}, *batchSize)
		httpHandlers := httpHandlers{annotations.NewAnnotationsService(batchRunner, db, *platformVersion)}
		r := router(httpHandlers)
		http.Handle("/", r)

		if err := http.ListenAndServe(fmt.Sprintf(":%d", *port),
			httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry,
				httphandlers.TransactionAwareRequestLoggingHandler(log.StandardLogger(), r))); err != nil {
			log.Fatalf("Unable to start server: %v", err)
		}
		baseftrwapp.OutputMetricsIfRequired(*graphiteTCPAddress, *graphitePrefix, *logMetrics)

	}
	setLogLevel(strings.ToUpper(*logLevel))
	app.Run(os.Args)
}

func router(hh httpHandlers) *mux.Router {
	r := mux.NewRouter()
	r.Headers("Content-type: application/json")

	// Healthchecks and standards first
	r.HandleFunc("/__health", v1a.Handler("Annotations RW Healthchecks",
		"Checks for accessing neo4j", hh.HealthCheck()))
	r.HandleFunc("/ping", Ping)
	r.HandleFunc("/__ping", Ping)

	// Then API specific ones:
	r.HandleFunc("/content/{uuid}/annotations", hh.GetAnnotations).Methods("GET")
	r.HandleFunc("/content/{uuid}/annotations", hh.PutAnnotations).Methods("PUT")
	r.HandleFunc("/content/{uuid}/annotations", hh.DeleteAnnotations).Methods("DELETE")
	r.HandleFunc("/content/annotations/__count", hh.CountAnnotations).Methods("GET")
	return r
}

func setLogLevel(level string) {
	switch level {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	default:
		log.Errorf("Requested log level %s is not supported, will default to INFO level", level)
		log.SetLevel(log.InfoLevel)
	}
	log.Debugf("Logging level set to %s", level)
}
