package utils

import (
	"bytes"
	"math/rand"
	"testing"
)

// deterministicBytes returns n pseudo-random bytes generated from a fixed
// seed so tests are fully reproducible.
func deterministicBytes(seed int64, n int) []byte {
	rng := rand.New(rand.NewSource(seed))
	data := make([]byte, n)
	rng.Read(data)
	return data
}

// TestEncodeDecodeRoundTrip verifies that every compression level round-trips
// both a compressible text dataset (small edit of repeated lines) and a
// binary dataset (pseudo-random bytes with two flipped bytes) exactly.
func TestEncodeDecodeRoundTrip(t *testing.T) {
	// Compressible dataset: ~256KB of repeated text lines with a few lines
	// changed and a block appended.
	var buf bytes.Buffer
	line := []byte("The quick brown fox jumps over the lazy dog; pack my box with five dozen liquor jugs.\n")
	for buf.Len() < 256*1024 {
		buf.Write(line)
	}
	textOld := buf.Bytes()
	textNew := append([]byte(nil), textOld...)
	copy(textNew[1024:1024+len(line)], bytes.Repeat([]byte("X"), len(line)))
	copy(textNew[4096:4096+len(line)], bytes.Repeat([]byte("Y"), len(line)))
	textNew = append(textNew, bytes.Repeat([]byte("appended trailer block of text\n"), 64)...)

	// Binary dataset: deterministic pseudo-random bytes with 2 bytes changed.
	binOld := deterministicBytes(42, 256*1024)
	binNew := append([]byte(nil), binOld...)
	binNew[123] ^= 0xFF
	binNew[200000] ^= 0xFF

	for _, level := range []int{1, 2, 3, 4} {
		for _, tc := range []struct {
			name string
			old  []byte
			new  []byte
		}{
			{"text", textOld, textNew},
			{"binary", binOld, binNew},
		} {
			delta, err := EncodeDelta(tc.old, tc.new, level)
			if err != nil {
				t.Fatalf("EncodeDelta(%s, level %d) returned error: %v", tc.name, level, err)
			}
			got, err := DecodeDelta(tc.old, delta)
			if err != nil {
				t.Fatalf("DecodeDelta(%s, level %d) returned error: %v", tc.name, level, err)
			}
			if !bytes.Equal(got, tc.new) {
				t.Errorf("round trip (%s, level %d): decoded %d bytes, want the exact %d-byte new data",
					tc.name, level, len(got), len(tc.new))
			}
		}
	}

	// A two-byte change in pseudo-random data must produce a delta that is
	// substantially smaller than the new data at the default level.
	delta, err := EncodeDelta(binOld, binNew, 3)
	if err != nil {
		t.Fatalf("EncodeDelta(binary, level 3) returned error: %v", err)
	}
	if limit := len(binNew) / 10; len(delta) >= limit {
		t.Errorf("binary delta at level 3 is %d bytes, want < %d (10%% of the %d-byte new data)",
			len(delta), limit, len(binNew))
	}
}

// TestEncodeDeltaEmptyBase verifies that encoding without base data fails.
func TestEncodeDeltaEmptyBase(t *testing.T) {
	newData := []byte("new content that requires a non-empty encoding base")

	if _, err := EncodeDelta(nil, newData, 3); err == nil {
		t.Error("EncodeDelta with nil base data must return an error")
	}
	if _, err := EncodeDelta([]byte{}, newData, 3); err == nil {
		t.Error("EncodeDelta with empty base data must return an error")
	}
}

// TestDecodeDeltaWrongBase verifies that a delta decoded against a base it
// was not encoded with never silently yields the correct new data: the
// decode must either fail or produce different bytes.
func TestDecodeDeltaWrongBase(t *testing.T) {
	oldData := deterministicBytes(7, 64*1024)
	newData := append([]byte(nil), oldData...)
	newData[100] ^= 0xA5
	newData[200] ^= 0x5A
	newData[300] ^= 0xFF

	delta, err := EncodeDelta(oldData, newData, 3)
	if err != nil {
		t.Fatalf("EncodeDelta returned error: %v", err)
	}

	tampered := append([]byte(nil), oldData...)
	tampered[5] ^= 0x01
	tampered[5000] ^= 0x02
	tampered[60000] ^= 0x03

	got, err := DecodeDelta(tampered, delta)
	if err == nil && bytes.Equal(got, newData) {
		t.Error("decoding a delta against a wrong base must never silently yield the correct new data")
	}
}

// TestShouldUseDelta verifies the strict-smaller rule for delta selection.
func TestShouldUseDelta(t *testing.T) {
	tests := []struct {
		deltaSize int64
		fullSize  int64
		expected  bool
	}{
		{10, 11, true},
		{11, 10, false},
		{10, 10, false},
		{0, 10, true},
		{5, 0, false},
	}
	for _, tc := range tests {
		if got := ShouldUseDelta(tc.deltaSize, tc.fullSize); got != tc.expected {
			t.Errorf("ShouldUseDelta(%d, %d) = %v, want %v", tc.deltaSize, tc.fullSize, got, tc.expected)
		}
	}
}

// TestDeltaWindowCoversBase verifies deltaWindowSize returns power-of-two
// windows clamped to the encoder limits that still cover the base size.
func TestDeltaWindowCoversBase(t *testing.T) {
	inputs := []int{
		1,
		1023,
		1024,
		1025,
		1024*1024 + 1,
		511 * 1024 * 1024,
		MaxDeltaBaseSize, // 512 MB
	}
	for _, in := range inputs {
		window := deltaWindowSize(in)
		if window < 1024 {
			t.Errorf("deltaWindowSize(%d) = %d, must be >= 1024", in, window)
		}
		if window > MaxDeltaBaseSize {
			t.Errorf("deltaWindowSize(%d) = %d, must be <= MaxDeltaBaseSize (%d)", in, window, MaxDeltaBaseSize)
		}
		if window&(window-1) != 0 {
			t.Errorf("deltaWindowSize(%d) = %d, must be a power of two", in, window)
		}
		if in <= MaxDeltaBaseSize && window < in {
			t.Errorf("deltaWindowSize(%d) = %d, window must cover the input size", in, window)
		}
	}
}
