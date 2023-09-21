package test

import "context"

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
	return nil
}
