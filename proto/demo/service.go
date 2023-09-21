package demo

import (
	"context"
	"fmt"
)

// NewService creates a new service instance
func NewService() *Service {
	return &Service{}
}

// Service is an example of a production ready service implementation
type Service struct{}

func (h *Service) SayHello(ctx context.Context, req *SayHelloRequest, resp *SayHelloResponse) error {
	// TODO: Validate the payload is valid (If the name isn't capitalized, we should reject it for giggles)
	resp.Message = fmt.Sprintf("Hello, %s", req.Name)
	return nil
}

// RenderPixel returns the color of a Mandelbrot fractal at the given point in the image.
// Code copied from Francesc Campoy's Golang Tracer example (https://tinyurl.com/ery6mfz8)
func (h *Service) RenderPixel(ctx context.Context, req *RenderPixelRequest, resp *RenderPixelResponse) error {
	xi := norm(req.I, req.Width, -1.0, 2)
	yi := norm(req.J, req.Height, -1, 1)

	const maxI = 1000
	x, y := 0., 0.

	for i := 0; (x*x+y*y < req.Complexity) && i < maxI; i++ {
		x, y = x*x-y*y+xi, 2*x*y+yi
	}

	resp.Gray = int64(x)
	return nil
}

func norm(x, total int64, min, max float64) float64 {
	return (max-min)*float64(x)/float64(total) - max
}
