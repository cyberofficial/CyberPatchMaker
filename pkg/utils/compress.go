package utils

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
)

// CompressData compresses data using the specified algorithm
func CompressData(data []byte, algorithm string, level int) ([]byte, error) {
	switch algorithm {
	case "zstd":
		return compressZstd(data, level)
	case "gzip":
		return compressGzip(data, level)
	case "none":
		return data, nil
	default:
		return nil, fmt.Errorf("unsupported compression algorithm: %s", algorithm)
	}
}

// DecompressData decompresses data using the specified algorithm
func DecompressData(data []byte, algorithm string) ([]byte, error) {
	switch algorithm {
	case "zstd":
		return decompressZstd(data)
	case "gzip":
		return decompressGzip(data)
	case "none":
		return data, nil
	default:
		return nil, fmt.Errorf("unsupported compression algorithm: %s", algorithm)
	}
}

func compressZstd(data []byte, level int) ([]byte, error) {
	var buf bytes.Buffer

	encoderLevel := zstd.SpeedDefault
	switch level {
	case 1:
		encoderLevel = zstd.SpeedFastest
	case 2:
		encoderLevel = zstd.SpeedDefault
	case 3:
		encoderLevel = zstd.SpeedBetterCompression
	case 4:
		encoderLevel = zstd.SpeedBestCompression
	}

	encoder, err := zstd.NewWriter(&buf, zstd.WithEncoderLevel(encoderLevel))
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd encoder: %w", err)
	}

	if _, err := encoder.Write(data); err != nil {
		encoder.Close()
		return nil, fmt.Errorf("failed to compress data: %w", err)
	}

	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("failed to close encoder: %w", err)
	}

	return buf.Bytes(), nil
}

func decompressZstd(data []byte) ([]byte, error) {
	decoder, err := zstd.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd decoder: %w", err)
	}
	defer decoder.Close()

	decompressed, err := io.ReadAll(decoder)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	return decompressed, nil
}

func compressGzip(data []byte, level int) ([]byte, error) {
	var buf bytes.Buffer

	gzipLevel := gzip.DefaultCompression
	switch level {
	case 1:
		gzipLevel = gzip.BestSpeed
	case 2:
		gzipLevel = gzip.DefaultCompression
	case 3:
		gzipLevel = gzip.BestCompression
	}

	writer, err := gzip.NewWriterLevel(&buf, gzipLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip writer: %w", err)
	}

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return nil, fmt.Errorf("failed to compress data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

func decompressGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	return decompressed, nil
}

// CompressDataStreaming compresses data using streaming to handle large data
func CompressDataStreaming(src io.Reader, dst io.Writer, algorithm string, level int) error {
	switch algorithm {
	case "zstd":
		return compressZstdStreaming(src, dst, level)
	case "gzip":
		return compressGzipStreaming(src, dst, level)
	case "none":
		_, err := io.Copy(dst, src)
		return err
	default:
		return fmt.Errorf("unsupported compression algorithm: %s", algorithm)
	}
}

// DecompressDataStreaming decompresses data using streaming
func DecompressDataStreaming(src io.Reader, dst io.Writer, algorithm string) error {
	switch algorithm {
	case "zstd":
		return decompressZstdStreaming(src, dst)
	case "gzip":
		return decompressGzipStreaming(src, dst)
	case "none":
		_, err := io.Copy(dst, src)
		return err
	default:
		return fmt.Errorf("unsupported compression algorithm: %s", algorithm)
	}
}

func compressZstdStreaming(src io.Reader, dst io.Writer, level int) error {
	encoderLevel := zstd.SpeedDefault
	switch level {
	case 1:
		encoderLevel = zstd.SpeedFastest
	case 2:
		encoderLevel = zstd.SpeedDefault
	case 3:
		encoderLevel = zstd.SpeedBetterCompression
	case 4:
		encoderLevel = zstd.SpeedBestCompression
	}

	encoder, err := zstd.NewWriter(dst, zstd.WithEncoderLevel(encoderLevel))
	if err != nil {
		return fmt.Errorf("failed to create zstd encoder: %w", err)
	}
	defer encoder.Close()

	if _, err := io.Copy(encoder, src); err != nil {
		return fmt.Errorf("failed to compress data: %w", err)
	}

	return nil
}

func decompressZstdStreaming(src io.Reader, dst io.Writer) error {
	decoder, err := zstd.NewReader(src)
	if err != nil {
		return fmt.Errorf("failed to create zstd decoder: %w", err)
	}
	defer decoder.Close()

	if _, err := io.Copy(dst, decoder); err != nil {
		return fmt.Errorf("failed to decompress data: %w", err)
	}

	return nil
}

func compressGzipStreaming(src io.Reader, dst io.Writer, level int) error {
	gzipLevel := gzip.DefaultCompression
	switch level {
	case 1:
		gzipLevel = gzip.BestSpeed
	case 2:
		gzipLevel = gzip.DefaultCompression
	case 3:
		gzipLevel = gzip.BestCompression
	}

	writer, err := gzip.NewWriterLevel(dst, gzipLevel)
	if err != nil {
		return fmt.Errorf("failed to create gzip writer: %w", err)
	}
	defer writer.Close()

	if _, err := io.Copy(writer, src); err != nil {
		return fmt.Errorf("failed to compress data: %w", err)
	}

	return nil
}

func decompressGzipStreaming(src io.Reader, dst io.Writer) error {
	reader, err := gzip.NewReader(src)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	if _, err := io.Copy(dst, reader); err != nil {
		return fmt.Errorf("failed to decompress data: %w", err)
	}

	return nil
}
