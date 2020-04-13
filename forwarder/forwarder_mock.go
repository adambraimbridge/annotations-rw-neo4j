package forwarder

import (
	"github.com/Financial-Times/annotations-rw-neo4j/v3/annotations"

	"github.com/stretchr/testify/mock"
)

type MockForwarder struct {
	mock.Mock
	Forwarder
}

func (mf *MockForwarder) SendMessage(transactionID string, originSystem string, headers map[string]string, uuid string, annotations annotations.Annotations) error {
	args := mf.Called(transactionID, originSystem, headers, uuid, annotations)
	return args.Error(0)
}
