package main

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
)

type HealthCheckHandletTestSuite struct {
	suite.Suite
	annotationsService *mockAnnotationsService
	httpHandler        httpHandlers
}

func (suite *HealthCheckHandletTestSuite) SetupTest() {
	suite.annotationsService = new(mockAnnotationsService)
	suite.httpHandler = httpHandlers{}
}
func TestHealthCheckHandletTestSuite(t *testing.T) {
	suite.Run(t, new(HealthCheckHandletTestSuite))
}

func (suite *HealthCheckHandletTestSuite) TestHealthCheckHandler_Health_Success() {
	suite.annotationsService.On("Check").Return(nil)
	req, err := http.NewRequest(http.MethodGet, "/__health", nil)
	assert.NoError(suite.T(), err, "Unexpected error")
	healthCheckHandler := healthCheckHandler{annotationsService: suite.annotationsService, consumer: mockConsumer{}}
	rec := httptest.NewRecorder()
	router(suite.httpHandler, healthCheckHandler).ServeHTTP(rec, req)
	assert.True(suite.T(), http.StatusOK == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusOK))
}

func (suite *HealthCheckHandletTestSuite) TestHealthCheckHandler_Health_AnnotationsServiceNotHealthy() {
	suite.annotationsService.On("Check").Return(errors.New("not healthy"))
	req, err := http.NewRequest(http.MethodGet, "/__health", nil)
	assert.NoError(suite.T(), err, "Unexpected error")
	healthCheckHandler := healthCheckHandler{annotationsService: suite.annotationsService, consumer: mockConsumer{}}
	rec := httptest.NewRecorder()
	router(suite.httpHandler, healthCheckHandler).ServeHTTP(rec, req)
	fmt.Println(rec.Body.String())
	assert.True(suite.T(), http.StatusOK == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusOK))
	assert.Contains(suite.T(), rec.Body.String(), `"ok":false`)
}

func (suite *HealthCheckHandletTestSuite) TestHealthCheckHandler_Health_ConsumerNotHealthy() {
	suite.annotationsService.On("Check").Return(nil)
	req, err := http.NewRequest(http.MethodGet, "/__health", nil)
	assert.NoError(suite.T(), err, "Unexpected error")
	healthCheckHandler := healthCheckHandler{annotationsService: suite.annotationsService, consumer: mockConsumer{err: errors.New("consumer error")}}
	rec := httptest.NewRecorder()
	router(suite.httpHandler, healthCheckHandler).ServeHTTP(rec, req)
	assert.True(suite.T(), http.StatusOK == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusOK))
	assert.Contains(suite.T(), rec.Body.String(), `"ok":false`)
}

func (suite *HealthCheckHandletTestSuite) TestHealthCheckHandler_GTG_Success() {
	suite.annotationsService.On("Check").Return(nil)
	req, err := http.NewRequest(http.MethodGet, "/__gtg", nil)
	assert.NoError(suite.T(), err, "Unexpected error")
	healthCheckHandler := healthCheckHandler{annotationsService: suite.annotationsService, consumer: mockConsumer{}}
	rec := httptest.NewRecorder()
	router(suite.httpHandler, healthCheckHandler).ServeHTTP(rec, req)
	assert.True(suite.T(), http.StatusOK == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusOK))
}

func (suite *HealthCheckHandletTestSuite) TestHealthCheckHandler_GTG_AnnotationsServiceNotHealthy() {
	suite.annotationsService.On("Check").Return(errors.New("not healthy"))
	req, err := http.NewRequest(http.MethodGet, "/__gtg", nil)
	assert.NoError(suite.T(), err, "Unexpected error")
	healthCheckHandler := healthCheckHandler{annotationsService: suite.annotationsService, consumer: mockConsumer{}}
	rec := httptest.NewRecorder()
	router(suite.httpHandler, healthCheckHandler).ServeHTTP(rec, req)
	fmt.Println(rec.Body.String())
	assert.True(suite.T(), http.StatusServiceUnavailable == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusServiceUnavailable))
}

func (suite *HealthCheckHandletTestSuite) TestHealthCheckHandler_GTG_ConsumerNotHealthy() {
	suite.annotationsService.On("Check").Return(nil)
	req, err := http.NewRequest(http.MethodGet, "/__gtg", nil)
	assert.NoError(suite.T(), err, "Unexpected error")
	healthCheckHandler := healthCheckHandler{annotationsService: suite.annotationsService, consumer: mockConsumer{err: errors.New("consumer error")}}
	rec := httptest.NewRecorder()
	router(suite.httpHandler, healthCheckHandler).ServeHTTP(rec, req)
	assert.True(suite.T(), http.StatusServiceUnavailable == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusServiceUnavailable))

}
