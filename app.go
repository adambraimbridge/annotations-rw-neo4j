package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/Financial-Times/annotations-rw-neo4j/annotations"
	"github.com/Financial-Times/base-ft-rw-app-go/baseftrwapp"
	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/http-handlers-go/httphandlers"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/Financial-Times/service-status-go/gtg"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"github.com/rcrowley/go-metrics"
)

func main() {

	app := cli.App("annotations-rw-neo4j", "A RESTful API for managing Annotations in neo4j")
	neoURL := app.String(cli.StringOpt{
		Name:   "neo-url",
		Value:  "http://localhost:7474/db/data",
		Desc:   "neo4j endpoint URL",
		EnvVar: "NEO_URL",
	})
	graphiteTCPAddress := app.String(cli.StringOpt{
		Name:   "graphiteTCPAddress",
		Value:  "",
		Desc:   "Graphite TCP address, e.g. graphite.ft.com:2003. Leave as default if you do NOT want to output to graphite (e.g. if running locally",
		EnvVar: "GRAPHITE_ADDRESS",
	})
	graphitePrefix := app.String(cli.StringOpt{
		Name:   "graphitePrefix",
		Value:  "",
		Desc:   "Prefix to use. Should start with content, include the environment, and the host name. e.g. coco.pre-prod.roles-rw-neo4j.1 or content.test.people.rw.neo4j.ftaps58938-law1a-eu-t",
		EnvVar: "GRAPHITE_PREFIX",
	})
	port := app.Int(cli.IntOpt{
		Name:   "port",
		Value:  8080,
		Desc:   "Port to listen on",
		EnvVar: "APP_PORT",
	})
	batchSize := app.Int(cli.IntOpt{
		Name:   "batchSize",
		Value:  1024,
		Desc:   "Maximum number of statements to execute per batch",
		EnvVar: "BATCH_SIZE",
	})
	logMetrics := app.Bool(cli.BoolOpt{
		Name:   "logMetrics",
		Value:  false,
		Desc:   "Whether to log metrics. Set to true if running locally and you want metrics output",
		EnvVar: "LOG_METRICS",
	})
	logLevel := app.String(cli.StringOpt{
		Name:   "log-level",
		Value:  "INFO",
		Desc:   "Logging level (DEBUG, INFO, WARN, ERROR)",
		EnvVar: "LOG_LEVEL",
	})
	platformVersion := app.String(cli.StringOpt{
		Name:   "platformVersion",
		Value:  "",
		Desc:   "Annotation source platform. Possible values are: v1 or v2.",
		EnvVar: "PLATFORM_VERSION",
	})

	app.Action = func() {
		parsedLogLevel, err := log.ParseLevel(*logLevel)
		if err != nil {
			log.WithFields(log.Fields{"logLevel": logLevel, "err": err}).Fatal("Incorrect log level")
		}
		log.SetLevel(parsedLogLevel)

		log.Infof("annotations-rw-neo4j will listen on port: %d, connecting to: %s", *port, *neoURL)

		conf := neoutils.DefaultConnectionConfig()
		conf.BatchSize = *batchSize
		db, err := neoutils.Connect(*neoURL, conf)

		if err != nil {
			log.Fatalf("Error connecting to neo4j %s", err)
		}

		annotationsService := annotations.NewCypherAnnotationsService(db, *platformVersion)
		httpHandlers := httpHandlers{annotationsService}

		// Healthchecks and standards first
		http.HandleFunc("/__health", v1a.Handler("Annotations RW Healthchecks",
			"Checks for accessing neo4j", httpHandlers.HealthCheck()))
		http.HandleFunc(status.PingPath, status.PingHandler)
		http.HandleFunc(status.PingPathDW, status.PingHandler)
		http.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)
		http.HandleFunc(status.BuildInfoPathDW, status.BuildInfoHandler)

		gtgChecker := make([]gtg.StatusChecker, 0)
		gtgChecker = append(gtgChecker, func() gtg.Status {
			if err := httpHandlers.AnnotationsService.Check(); err != nil {
				return gtg.Status{GoodToGo: false, Message: err.Error()}
			}

			return gtg.Status{GoodToGo: true}
		})
		http.HandleFunc(status.GTGPath, status.NewGoodToGoHandler(gtg.FailFastParallelCheck(gtgChecker)))

		r := router(httpHandlers)
		http.Handle("/", r)
		baseftrwapp.OutputMetricsIfRequired(*graphiteTCPAddress, *graphitePrefix, *logMetrics)

		if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
			log.Fatalf("Unable to start server: %v", err)
		}

	}
	log.Infof("Application started with args %s", os.Args)
	app.Run(os.Args)
}

func router(hh httpHandlers) http.Handler {
	servicesRouter := mux.NewRouter()
	servicesRouter.Headers("Content-type: application/json")
	// Then API specific ones:
	servicesRouter.HandleFunc("/content/{uuid}/annotations", hh.GetAnnotations).Methods("GET")
	servicesRouter.HandleFunc("/content/{uuid}/annotations", hh.PutAnnotations).Methods("PUT")
	servicesRouter.HandleFunc("/content/{uuid}/annotations", hh.DeleteAnnotations).Methods("DELETE")
	servicesRouter.HandleFunc("/content/annotations/__count", hh.CountAnnotations).Methods("GET")

	var monitoringRouter http.Handler = servicesRouter
	monitoringRouter = httphandlers.TransactionAwareRequestLoggingHandler(log.StandardLogger(), monitoringRouter)
	monitoringRouter = httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry, monitoringRouter)

	return monitoringRouter
}
