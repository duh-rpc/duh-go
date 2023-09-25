package test

import (
	"context"
	"errors"
	"github.com/duh-rpc/duh-go"
)

// NewService creates a new service instance
func NewService() *Service {
	return &Service{}
}

// Service is an example of a production ready service implementation
type Service struct{}

// TestErrors will return a variety of errors depending on the req.Case provided.
// Suitable for testing client implementations.
// This method only responds with errors.
func (h *Service) TestErrors(ctx context.Context, req *ErrorsRequest) error {
	switch req.Case {
	case CaseServiceReturnedMessage:
		return duh.NewServiceError(duh.CodeNotFound, "The thing you asked for does not exist", nil, nil)
	case CaseServiceReturnedError:
		return duh.NewServiceError(duh.CodeInternalError, "", errors.New("while reading the database: EOF"), nil)
	}

	return nil
}
