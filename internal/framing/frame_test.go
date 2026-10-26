
package framing

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"
)

type chunkedReader struct {
    reader *bytes.Reader
    maxRead int
}

func (r *chunkedReader) Read(p []byte) (int, error) {
    if len(p) > r.maxRead {
        p = p[:r.maxRead]
    }

    return r.reader.Read(p)
}

func TestWriteFrame(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		want    []byte
	}{
		{
			name:    "normal payload",
			payload: []byte("hello"),
			want:    []byte{0, 0, 0, 5, 'h', 'e', 'l', 'l', 'o'},
		},
		{
			name:    "empty payload",
			payload: []byte{},
			want:    []byte{0, 0, 0, 0},
		},
		{
			name:    "binary payload",
			payload: []byte{0x00, 0xFF, 0x10},
			want:    []byte{0, 0, 0, 3, 0x00, 0xFF, 0x10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			if err := WriteFrame(&buf, tt.payload); err != nil {
				t.Fatalf("WriteFrame() error = %v", err)
			}

			if !bytes.Equal(buf.Bytes(), tt.want) {
				t.Fatalf("WriteFrame() = %v, want %v",
					buf.Bytes(), tt.want)
			}
		})
	}
}

func TestWriteFrameRejectsOversizedPayload(t *testing.T) {
	// The extra byte puts the payload just over the limit.
	payload := make([]byte, MaxPayloadSize+1)

	var buf bytes.Buffer
	err := WriteFrame(&buf, payload)

	if err == nil {
		t.Fatal("expected oversized payload error")
	}

	if buf.Len() != 0 {
		t.Fatalf("expected no bytes written, got %d", buf.Len())
	}
}

func TestWriteFrameMaximumPayload(t *testing.T) {
	payload := make([]byte, MaxPayloadSize)

	var buf bytes.Buffer
	if err := WriteFrame(&buf, payload); err != nil {
		t.Fatalf("WriteFrame() rejected maximum payload: %v", err)
	}

	if got, want := buf.Len(), 4+MaxPayloadSize; got != want {
		t.Fatalf("encoded length = %d, want %d", got, want)
	}

	gotLength := binary.BigEndian.Uint32(buf.Bytes()[:4])
	if gotLength != uint32(MaxPayloadSize) {
		t.Fatalf("encoded payload length = %d, want %d",
			gotLength, MaxPayloadSize)
	}
}

// shortWriter accepts at most max bytes per Write call.
type shortWriter struct {
	buf bytes.Buffer
	max int
}

func (w *shortWriter) Write(p []byte) (int, error) {
	if len(p) > w.max {
		p = p[:w.max]
	}
	return w.buf.Write(p)
}

func TestWriteFrameHandlesShortWrites(t *testing.T) {
	w := &shortWriter{max: 2}
	payload := []byte("hello")

	if err := WriteFrame(w, payload); err != nil {
		t.Fatalf("WriteFrame() error = %v", err)
	}

	want := []byte{0, 0, 0, 5, 'h', 'e', 'l', 'l', 'o'}
	if !bytes.Equal(w.buf.Bytes(), want) {
		t.Fatalf("encoded frame = %v, want %v", w.buf.Bytes(), want)
	}
}

var errTestWriter = errors.New("simulated writer failure")

type failingWriter struct {
	remaining int
}

func (w *failingWriter) Write(p []byte) (int, error) {
	if w.remaining == 0 {
		return 0, errTestWriter
	}

	n := len(p)
	if n > w.remaining {
		n = w.remaining
	}

	w.remaining -= n

	if w.remaining == 0 {
		// Simulate a failure after some bytes have been written.
		return n, errTestWriter
	}

	return n, nil
}

func TestWriteFramePropagatesWriterError(t *testing.T) {
	// Allow part of the prefix to be written, then fail.
	w := &failingWriter{remaining: 2}

	err := WriteFrame(w, []byte("hello"))
	if !errors.Is(err, errTestWriter) {
		t.Fatalf("expected writer error, got %v", err)
	}
}

type noProgressWriter struct{}

func (noProgressWriter) Write([]byte) (int, error) {
	return 0, nil
}

func TestWriteFrameNoProgress(t *testing.T) {
	err := WriteFrame(noProgressWriter{}, []byte("hello"))

	if !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("expected io.ErrShortWrite, got %v", err)
	}
}
func TestReadFrame(t *testing.T) {
    tests := []struct {
        name    string
        payload []byte
    }{
        {
            name:    "normal payload",
            payload: []byte("hello"),
        },
        {
            name:    "empty payload",
            payload: []byte{},
        },
        {
            name:    "binary payload",
            payload: []byte{0x00, 0xFF, 0x10, 0x80},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var buf bytes.Buffer

            if err := WriteFrame(&buf, tt.payload); err != nil {
                t.Fatalf("WriteFrame() error = %v", err)
            }

            got, err := ReadFrame(&buf)
            if err != nil {
                t.Fatalf("ReadFrame() error = %v", err)
            }

            if !bytes.Equal(got, tt.payload) {
                t.Fatalf("ReadFrame() = %v, want %v", got, tt.payload)
            }
        })
    }
}

func TestReadFrameTruncatedInput(t *testing.T) {
    tests := []struct {
        name string
        data []byte
    }{
        {
            name: "truncated header",
            data: []byte{0x00, 0x00},
        },
        {
            name: "truncated payload",
            data: []byte{
                0x00, 0x00, 0x00, 0x05,
                'h', 'i',
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := ReadFrame(bytes.NewReader(tt.data))
            if err == nil {
                t.Fatal("ReadFrame() expected an error, got nil")
            }

            if !errors.Is(err, io.ErrUnexpectedEOF) {
                t.Fatalf(
                    "ReadFrame() error = %v, want io.ErrUnexpectedEOF",
                    err,
                )
            }
        })
    }
}
func TestReadFrameRejectsOversizedPayload(t *testing.T) {
    var header [4]byte
    binary.BigEndian.PutUint32(
        header[:],
        uint32(MaxPayloadSize)+1,
    )

    _, err := ReadFrame(bytes.NewReader(header[:]))
    if err == nil {
        t.Fatal("ReadFrame() expected oversized payload error")
    }
}

func TestReadFrameHandlesPartialReads(t *testing.T) {
    payload := []byte("production-ready framing")

    var buf bytes.Buffer
    if err := WriteFrame(&buf, payload); err != nil {
        t.Fatalf("WriteFrame() error = %v", err)
    }

    reader := &chunkedReader{
        reader:  bytes.NewReader(buf.Bytes()),
        maxRead: 2,
    }

    got, err := ReadFrame(reader)
    if err != nil {
        t.Fatalf("ReadFrame() error = %v", err)
    }

    if !bytes.Equal(got, payload) {
        t.Fatalf("ReadFrame() = %q, want %q", got, payload)
    }
}