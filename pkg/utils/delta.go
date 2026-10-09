package utils

import (
	"bytes"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
)

// Delta compression uses zstd dictionary encoding to store only the
// difference between two versions of a file: the old file content is
// registered as the encoder dictionary, and the new file content is
// compressed against it. This is the same mechanism as `zstd --patch-from`.
// The decoder reconstructs the new file by registering the same dictionary
// (read from the file on disk at apply time) and decompressing the delta.

// DeltaDictID is the zstd dictionary identifier embedded in every delta
// frame. A fixed non-zero ID lets the decoder reliably select the
// registered base-file dictionary.
const DeltaDictID uint32 = 0x43504D44 // "CPMD"

const (
	// MaxDeltaBaseSize is the largest base (old) file eligible for delta
	// compression. klauspost/compress caps the encoder window at 512MB
	// (zstd.MaxWindowSize), and the window must cover the dictionary for
	// matches to reach the whole base file. Larger files fall back to full
	// file replacement.
	MaxDeltaBaseSize = 512 * 1024 * 1024 // 512 MB

	// MaxDeltaTargetSize is the largest modified (new) file eligible for
	// delta compression. It matches the project's large-file threshold:
	// files above this stay on the existing full-replacement path so the
	// extra decode-time memory of holding base + reconstructed content in
	// memory never applies to the project's "large file" regime.
	MaxDeltaTargetSize = LargeFileThreshold // 1 GB
)

// EncodeDelta compresses newData against oldData using oldData as a zstd
// dictionary (initial history). level uses the same 1-4 mapping as the
// patch compression option (zstd encoder speed presets). The returned bytes
// are a self-contained zstd frame that requires oldData as its dictionary
// to decompress.
func EncodeDelta(oldData, newData []byte, level int) ([]byte, error) {
	if len(oldData) == 0 {
		return nil, fmt.Errorf("delta encoding requires non-empty base data")
	}

	encoder, err := zstd.NewWriter(nil,
		zstd.WithEncoderDictRaw(DeltaDictID, oldData),
		zstd.WithEncoderLevel(mapZstdLevel(level)),
		zstd.WithWindowSize(deltaWindowSize(len(oldData))),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create delta encoder: %w", err)
	}
	defer encoder.Close()

	return encoder.EncodeAll(newData, make([]byte, 0, len(newData)/4)), nil
}

// DecodeDelta reconstructs the new file content from a delta produced by
// EncodeDelta. oldData must be the exact base content the delta was encoded
// against; any other base yields corrupted output or a decode error. Callers
// must verify the result (e.g. against the expected new-file checksum).
func DecodeDelta(oldData, delta []byte) ([]byte, error) {
	decoder, err := zstd.NewReader(bytes.NewReader(delta),
		zstd.WithDecoderDictRaw(DeltaDictID, oldData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create delta decoder: %w", err)
	}
	defer decoder.Close()

	newData, err := io.ReadAll(decoder)
	if err != nil {
		return nil, fmt.Errorf("failed to decode delta: %w", err)
	}
	return newData, nil
}

// ShouldUseDelta decides whether a delta is stored instead of the full new
// file. The full file is preferred unless the delta is strictly smaller, so
// deltas that are larger or equal in size fall back to full replacement.
func ShouldUseDelta(deltaSize, fullSize int64) bool {
	return deltaSize < fullSize
}

// deltaWindowSize returns the power-of-two zstd window, clamped to the
// encoder limits (1KB..512MB). The window must be STRICTLY GREATER than the
// dictionary size: klauspost/compress match finders only accept match
// offsets below the window (`coffset < e.maxMatchOff`), so an identical
// prefix in the new content needs an offset of exactly dictSize to reference
// the start of the dictionary. A window equal to the dictionary therefore
// suppresses all dictionary matches. Bases of exactly 512MB (the encoder
// window cap) cannot get a larger window; those degrade gracefully via the
// ShouldUseDelta full-replacement fallback.
func deltaWindowSize(dictSize int) int {
	window := zstd.MinWindowSize
	for window <= dictSize && window < MaxDeltaBaseSize {
		window <<= 1
	}
	return window
}
