package main

import (
	"fmt"
	"github.com/Financial-Times/annotations-rw-neo4j/annotations"
	"github.com/Financial-Times/base-ft-rw-app-go"
	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/http-handlers-go"
	"github.com/Financial-Times/neo-cypher-runner-go"
	"github.com/Financial-Times/neo-utils-go"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"github.com/jmcvetta/neoism"
	"github.com/rcrowley/go-metrics"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
)

func main() {
	log.SetLevel(log.InfoLevel)
	log.Infof("Application started with args %s", os.Args)

	app := cli.App("people-rw-neo4j", "A RESTful API for managing People in neo4j")
	neoURL := app.StringOpt("neo-url", "http://localhost:7474/db/data", "neo4j endpoint URL")
	port := app.IntOpt("port", 8080, "Port to listen on")
	batchSize := app.IntOpt("batchSize", 1024, "Maximum number of statements to execute per batch")
	graphiteTCPAddress := app.StringOpt("graphiteTCPAddress", "",
		"Graphite TCP address, e.g. graphite.ft.com:2003. Leave as default if you do NOT want to output to graphite (e.g. if running locally)")
	graphitePrefix := app.StringOpt("graphitePrefix", "",
		"Prefix to use. Should start with content, include the environment, and the host name. e.g. content.test.annotation.rw.neo4j.ftaps58938-law1a-eu-t")
	logMetrics := app.BoolOpt("logMetrics", false, "Whether to log metrics. Set to true if running locally and you want metrics output")
	logLevel := app.StringOpt("log-level", "INFO", "Logging level (DEBUG, INFO, WARN, ERROR)")

	app.Action = func() {
		setLogLevel(strings.ToUpper(*logLevel))
		db, err := neoism.Connect(*neoURL)
		if err != nil {
			log.Fatalf("Error connecting to neo4j %s", err)
		}
		batchRunner := neocypherrunner.NewBatchCypherRunner(neoutils.StringerDb{db}, *batchSize)
		annotations.AnnotationsDriver = annotations.NewCypherDriver(batchRunner, db)
		r := mux.NewRouter()
		r.Headers("Content-type: application/json")

		// Healthchecks and standards first
		r.HandleFunc("/__health", v1a.Handler("PeopleReadWriteNeo4j Healthchecks",
			"Checks for accessing neo4j", annotations.HealthCheck()))
		r.HandleFunc("/ping", annotations.Ping)
		r.HandleFunc("/__ping", annotations.Ping)

		// Then API specific ones:
		r.HandleFunc("/content/{uuid}/annotations", annotations.GetAnnotations).Methods("GET")
		r.HandleFunc("/content/{uuid}/annotations", annotations.PutAnnotations).Methods("PUT")

		if err := http.ListenAndServe(fmt.Sprintf(":%d", *port),
			httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry,
				httphandlers.TransactionAwareRequestLoggingHandler(log.StandardLogger(), r))); err != nil {
			log.Fatalf("Unable to start server: %v", err)
		}
		baseftrwapp.OutputMetricsIfRequired(*graphiteTCPAddress, *graphitePrefix, *logMetrics)
		log.Infof("public-people-api will listen on port: %d, connecting to: %s\n", *port, *neoURL)
	}
	app.Run(os.Args)
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
