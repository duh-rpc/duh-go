package test

import (
	"bufio"
	"github.com/harbor-pkgs/duh"
	"io"
	"log"
	"math/rand"
	"net/http"
)

type Handler struct {
	Service *Service
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/v1/test.errors":
		h.handleTestErrors(w, r)
		return
	}
}

func (h *Handler) handleTestErrors(w http.ResponseWriter, r *http.Request) {
	var req ErrorsRequest

	if err := duh.ReadRequest(r, &req); err != nil {
		duh.ReplyError(w, r, err)
		return
	}

	switch req.Case {
	case CaseClientIOError:
		// Force a chunked response by sending a ton of garbage to the client, this
		// ensures the client will receive response headers such that the panic
		//  abruptly cuts off the stream of chunks.
		if err := chunkWriter(w, 100_000); err != nil {
			log.Printf("during chunkWriter: %s", err)
		}
		panic("Panic in a Handler should terminate the connection")
	}

	duh.ReplyError(w, r, h.Service.TestErrors(r.Context(), &req))
}

func chunkWriter(w io.Writer, length int) error {
	const (
		charset = "abcdefghijklmnopqrstuvwxyz" +
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		size = 4096
	)
	buf := bufio.NewWriter(w)

	for length > 0 {
		chunk := make([]byte, size)
		for i := 0; i < size && length > 0; i++ {
			chunk[i] = charset[rand.Intn(len(charset))]
			length--
		}
		if _, err := buf.Write(chunk); err != nil {
			return err
		}
	}
	return buf.Flush()
}
