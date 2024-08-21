package limit

import (
	"fmt"
	"github.com/duh-rpc/duh-go"
	v1 "github.com/duh-rpc/duh-go/proto/v1"
	"google.golang.org/protobuf/proto"
	"io"
)

// NewReader returns a Reader that returns ErrDataLimitExceeded after n bytes have been read.
func NewReader(r io.Reader, n int64) io.Reader {
	return &Reader{r: r, remain: n, max: n}
}

type Reader struct {
	r      io.Reader
	remain int64
	max    int64
}

func (l *Reader) Read(p []byte) (n int, err error) {
	if l.remain <= 0 {
		return 0, &ErrDataLimitExceeded{Max: l.max}
	}
	if int64(len(p)) > l.remain {
		p = p[0:l.remain]
	}
	n, err = l.r.Read(p)
	l.remain -= int64(n)
	return
}

type ErrDataLimitExceeded struct {
	Max int64
}

func (e *ErrDataLimitExceeded) ProtoMessage() proto.Message {
	return &v1.Reply{
		CodeText: duh.CodeText(duh.CodeBadRequest),
		Code:     duh.CodeBadRequest,
		Message:  e.Message(),
		Details:  nil,
	}
}

func (e *ErrDataLimitExceeded) Code() int {
	return duh.CodeBadRequest
}

func (e *ErrDataLimitExceeded) Message() string {
	return fmt.Sprintf("exceeds %s limit", byteToSI(e.Max))
}

func (e *ErrDataLimitExceeded) Error() string {
	return e.Message()
}

func (e *ErrDataLimitExceeded) Details() map[string]string {
	return nil
}

func byteToSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
