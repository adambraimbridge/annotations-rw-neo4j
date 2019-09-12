package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	logger "github.com/Financial-Times/go-logger/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HealthCheckHandlerTestSuite struct {
	suite.Suite
	annotationsService *mockAnnotationsService
	httpHandler        httpHandler
	log                *logger.UPPLogger
}

func (suite *HealthCheckHandlerTestSuite) SetupTest() {
	suite.annotationsService = new(mockAnnotationsService)
	suite.httpHandler = httpHandler{}
}
func TestHealthCheckHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HealthCheckHandlerTestSuite))
}

func (suite *HealthCheckHandlerTestSuite) TestHealthCheckHandler_Health_Success() {
	suite.annotationsService.On("Check").Return(nil)
	req, err := http.NewRequest(http.MethodGet, "/__health", nil)
	assert.NoError(suite.T(), err, "Unexpected error")
	healthCheckHandler := healthCheckHandler{annotationsService: suite.annotationsService, consumer: mockConsumer{}}
	rec := httptest.NewRecorder()
	router(&suite.httpHandler, &healthCheckHandler, suite.log).ServeHTTP(rec, req)
	assert.True(suite.T(), http.StatusOK == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusOK))
}

func (suite *HealthCheckHandlerTestSuite) TestHealthCheckHandler_Health_AnnotationsServiceNotHealthy() {
	suite.annotationsService.On("Check").Return(errors.New("not healthy"))
	req, err := http.NewRequest(http.MethodGet, "/__health", nil)
	assert.NoError(suite.T(), err, "Unexpected error")
	healthCheckHandler := healthCheckHandler{annotationsService: suite.annotationsService, consumer: mockConsumer{}}
	rec := httptest.NewRecorder()
	router(&suite.httpHandler, &healthCheckHandler, suite.log).ServeHTTP(rec, req)
	assert.True(suite.T(), http.StatusOK == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusOK))
	assert.Contains(suite.T(), rec.Body.String(), `"ok":false`)
}

func (suite *HealthCheckHandlerTestSuite) TestHealthCheckHandler_Health_ConsumerNotHealthy() {
	suite.annotationsService.On("Check").Return(nil)
	req, err := http.NewRequest(http.MethodGet, "/__health", nil)
	assert.NoError(suite.T(), err, "Unexpected error")
	healthCheckHandler := healthCheckHandler{annotationsService: suite.annotationsService, consumer: mockConsumer{err: errors.New("consumer error")}}
	rec := httptest.NewRecorder()
	router(&suite.httpHandler, &healthCheckHandler, suite.log).ServeHTTP(rec, req)
	assert.True(suite.T(), http.StatusOK == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusOK))
	assert.Contains(suite.T(), rec.Body.String(), `"ok":false`)
}

func (suite *HealthCheckHandlerTestSuite) TestHealthCheckHandler_GTG_Success() {
	suite.annotationsService.On("Check").Return(nil)
	req, err := http.NewRequest(http.MethodGet, "/__gtg", nil)
	assert.NoError(suite.T(), err, "Unexpected error")
	healthCheckHandler := healthCheckHandler{annotationsService: suite.annotationsService, consumer: mockConsumer{}}
	rec := httptest.NewRecorder()
	router(&suite.httpHandler, &healthCheckHandler, suite.log).ServeHTTP(rec, req)
	assert.True(suite.T(), http.StatusOK == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusOK))
}

func (suite *HealthCheckHandlerTestSuite) TestHealthCheckHandler_GTG_AnnotationsServiceNotHealthy() {
	suite.annotationsService.On("Check").Return(errors.New("not healthy"))
	req, err := http.NewRequest(http.MethodGet, "/__gtg", nil)
	assert.NoError(suite.T(), err, "Unexpected error")
	healthCheckHandler := healthCheckHandler{annotationsService: suite.annotationsService, consumer: mockConsumer{}}
	rec := httptest.NewRecorder()
	router(&suite.httpHandler, &healthCheckHandler, suite.log).ServeHTTP(rec, req)
	fmt.Println(rec.Body.String())
	assert.True(suite.T(), http.StatusServiceUnavailable == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusServiceUnavailable))
}

func (suite *HealthCheckHandlerTestSuite) TestHealthCheckHandler_GTG_ConsumerNotHealthy() {
	suite.annotationsService.On("Check").Return(nil)
	req, err := http.NewRequest(http.MethodGet, "/__gtg", nil)
	assert.NoError(suite.T(), err, "Unexpected error")
	healthCheckHandler := healthCheckHandler{annotationsService: suite.annotationsService, consumer: mockConsumer{err: errors.New("consumer error")}}
	rec := httptest.NewRecorder()
	router(&suite.httpHandler, &healthCheckHandler, suite.log).ServeHTTP(rec, req)
	assert.True(suite.T(), http.StatusServiceUnavailable == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusServiceUnavailable))
}

func (suite *HealthCheckHandlerTestSuite) TestHealthCheckHandler_GTG_NilConsumer() {
	suite.annotationsService.On("Check").Return(nil)
	req, err := http.NewRequest(http.MethodGet, "/__gtg", nil)
	assert.NoError(suite.T(), err, "Unexpected error")
	healthCheckHandler := healthCheckHandler{annotationsService: suite.annotationsService, consumer: nil}
	rec := httptest.NewRecorder()
	router(&suite.httpHandler, &healthCheckHandler, suite.log).ServeHTTP(rec, req)
	assert.True(suite.T(), http.StatusOK == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusOK))
}
