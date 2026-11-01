
package framing

import (
	"bytes"
	"io"
	"testing"
)

var benchmarkPayloadSink []byte

func BenchmarkWriteFrame(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"Empty", 0},
		{"Small_64B", 64},
		{"Medium_1KB", 1024},
		{"Large_64KB", 64 * 1024},
	}

	for _, tc := range sizes {
		b.Run(tc.name, func(b *testing.B) {
			payload := bytes.Repeat([]byte("a"), tc.size)

			b.ReportAllocs()
			b.SetBytes(int64(len(payload) + 4))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if err := WriteFrame(io.Discard, payload); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkReadFrame(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"Empty", 0},
		{"Small_64B", 64},
		{"Medium_1KB", 1024},
		{"Large_64KB", 64 * 1024},
	}

	for _, tc := range sizes {
		b.Run(tc.name, func(b *testing.B) {
			payload := bytes.Repeat([]byte("a"), tc.size)

			// Prepare one valid encoded frame outside the timed section.
			var encoded bytes.Buffer
			if err := WriteFrame(&encoded, payload); err != nil {
				b.Fatal(err)
			}

			frame := encoded.Bytes()
			var reader bytes.Reader

			b.ReportAllocs()
			b.SetBytes(int64(len(frame)))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				reader.Reset(frame)

				got, err := ReadFrame(&reader)
				if err != nil {
					b.Fatal(err)
				}

				benchmarkPayloadSink = got
			}
		})
	}
}
