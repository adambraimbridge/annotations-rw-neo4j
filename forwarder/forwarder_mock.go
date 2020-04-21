package forwarder

import (
	"github.com/Financial-Times/annotations-rw-neo4j/v3/annotations"

	"github.com/stretchr/testify/mock"
)

type MockForwarder struct {
	mock.Mock
}

func (mf *MockForwarder) SendMessage(transactionID string, originSystem string, uuid string, annotations annotations.Annotations) error {
	args := mf.Called(transactionID, originSystem, uuid, annotations)
	return args.Error(0)
}
