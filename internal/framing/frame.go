
package framing

import (
	"encoding/binary"
	"fmt"
	"io"
)

const MaxPayloadSize = 16 * 1024 * 1024

func WriteFrame(w io.Writer, payload []byte) error {
	if len(payload) > MaxPayloadSize {
		return fmt.Errorf(
			"frame payload too large: %d bytes exceeds maximum of %d bytes",
			len(payload),
			MaxPayloadSize,
		)
	}
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(payload)))
	if err := writeFull(w, header[:]); err != nil {
		return fmt.Errorf("write frame length prefix: %w", err)
	}
	if len(payload) == 0 {
		return nil
	}
	if err := writeFull(w, payload); err != nil {
		return fmt.Errorf("write frame payload: %w", err)
	}

	return nil
}

func writeFull(w io.Writer, p []byte) error {
	for len(p) > 0 {
		n, err := w.Write(p)

		if n < 0 || n > len(p) {
			return fmt.Errorf(
				"invalid write count %d for buffer of %d bytes",
				n,
				len(p),
			)
		}

		p = p[n:]

		if err != nil {
			return err
		}

		if n == 0 {
			return io.ErrShortWrite
		}
	}

	return nil
}
