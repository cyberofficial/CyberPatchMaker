package utils

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// SavePatch saves a patch to a file with optional compression
func SavePatch(patch *Patch, filename string, compression string) error {
	// Create output file
	outFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create patch file: %w", err)
	}
	defer outFile.Close()

	// Create a pipe for streaming JSON encoding
	jsonReader, jsonWriter := io.Pipe()
	defer jsonReader.Close()

	// Start custom streaming JSON encoding in a goroutine
	encodeErr := make(chan error, 1)
	go func() {
		defer jsonWriter.Close()
		encodeErr <- encodePatchStreaming(patch, jsonWriter)
	}()

	// Set up compression if needed
	var finalReader io.Reader = jsonReader

	if compression != "none" && compression != "" {
		// Create a pipe for compression
		compressedReader, compressor := io.Pipe()

		// Start compression in a goroutine
		go func() {
			defer compressor.Close()
			err := CompressDataStreaming(jsonReader, compressor, compression, 3) // Default level 3
			if err != nil {
				compressor.CloseWithError(err)
			}
		}()

		finalReader = compressedReader
	}

	// Create a hash writer to calculate checksum while writing
	hasher := sha256.New()
	multiWriter := io.MultiWriter(outFile, hasher)

	// Copy data through compression and hashing
	_, err = io.Copy(multiWriter, finalReader)
	if err != nil {
		return fmt.Errorf("failed to write patch data: %w", err)
	}

	// Close output file
	if err := outFile.Close(); err != nil {
		return fmt.Errorf("failed to close output file: %w", err)
	}

	// Check for encoding errors
	if err := <-encodeErr; err != nil {
		return fmt.Errorf("failed to encode patch: %w", err)
	}

	// Calculate checksum
	checksum := fmt.Sprintf("%x", hasher.Sum(nil))
	patch.Header.Checksum = checksum

	return nil
}

// LoadPatch loads a patch from a file
func LoadPatch(filename string) (*Patch, error) {
	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read patch file: %w", err)
	}

	// Try to parse as uncompressed JSON first
	var patch Patch
	if err := json.Unmarshal(data, &patch); err == nil {
		return &patch, nil
	}

	// Try decompressing with zstd
	decompressed, err := DecompressData(data, "zstd")
	if err == nil {
		if err := json.Unmarshal(decompressed, &patch); err == nil {
			return &patch, nil
		}
	}

	// Try gzip
	decompressed, err = DecompressData(data, "gzip")
	if err == nil {
		if err := json.Unmarshal(decompressed, &patch); err == nil {
			return &patch, nil
		}
	}

	return nil, fmt.Errorf("failed to load and decompress patch")
}

// encodePatchStreaming writes the patch as JSON in a streaming fashion to avoid memory exhaustion
func encodePatchStreaming(patch *Patch, writer io.Writer) error {
	// Create a buffered writer for better performance
	bufWriter := bufio.NewWriterSize(writer, 64*1024) // 64KB buffer
	defer bufWriter.Flush()

	// Write opening brace
	if _, err := bufWriter.WriteString("{\n"); err != nil {
		return err
	}

	// Encode header
	if err := encodeField(bufWriter, "Header", patch.Header, true); err != nil {
		return err
	}

	// Encode simple fields
	if err := encodeField(bufWriter, "FromVersion", patch.FromVersion, true); err != nil {
		return err
	}
	if err := encodeField(bufWriter, "ToVersion", patch.ToVersion, true); err != nil {
		return err
	}
	if err := encodeField(bufWriter, "FromKeyFile", patch.FromKeyFile, true); err != nil {
		return err
	}
	if err := encodeField(bufWriter, "ToKeyFile", patch.ToKeyFile, true); err != nil {
		return err
	}
	if err := encodeField(bufWriter, "RequiredFiles", patch.RequiredFiles, true); err != nil {
		return err
	}
	if err := encodeField(bufWriter, "SimpleMode", patch.SimpleMode, true); err != nil {
		return err
	}

	// Encode multi-part info if present
	if patch.MultiPart != nil {
		if err := encodeField(bufWriter, "MultiPart", patch.MultiPart, true); err != nil {
			return err
		}
	} else {
		// Write null explicitly
		if _, err := bufWriter.WriteString(`  "MultiPart": null,`); err != nil {
			return err
		}
		if _, err := bufWriter.WriteString("\n"); err != nil {
			return err
		}
	}

	// Encode operations array manually to stream large data
	if _, err := bufWriter.WriteString(`  "Operations": [`); err != nil {
		return err
	}

	for i, op := range patch.Operations {
		if i > 0 {
			if _, err := bufWriter.WriteString(",\n"); err != nil {
				return err
			}
		} else {
			if _, err := bufWriter.WriteString("\n"); err != nil {
				return err
			}
		}

		if err := encodeOperation(bufWriter, op); err != nil {
			return err
		}
	}

	if _, err := bufWriter.WriteString("\n  ]\n"); err != nil {
		return err
	}

	// Write closing brace
	if _, err := bufWriter.WriteString("}\n"); err != nil {
		return err
	}

	return bufWriter.Flush()
}

// encodeField encodes a single field with proper JSON formatting
func encodeField(writer io.Writer, name string, value interface{}, addComma bool) error {
	commaStr := ","
	if !addComma {
		commaStr = ""
	}

	// For simple types, use JSON encoding
	data, err := json.MarshalIndent(value, "  ", "  ")
	if err != nil {
		return err
	}

	// Write field name and value
	fieldStr := fmt.Sprintf("  \"%s\": ", name)
	if _, err := writer.Write([]byte(fieldStr)); err != nil {
		return err
	}

	// Write the JSON data, but indent it properly
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if i > 0 {
			if _, err := writer.Write([]byte("\n")); err != nil {
				return err
			}
		}
		if _, err := writer.Write([]byte(line)); err != nil {
			return err
		}
	}

	if _, err := writer.Write([]byte(commaStr + "\n")); err != nil {
		return err
	}

	return nil
}

// encodeOperation encodes a single patch operation with streaming for large binary data
func encodeOperation(writer io.Writer, op PatchOperation) error {
	// Write operation opening
	if _, err := writer.Write([]byte("    {\n")); err != nil {
		return err
	}

	// Encode simple fields
	if err := encodeOperationField(writer, "Type", int(op.Type), true); err != nil {
		return err
	}
	if err := encodeOperationField(writer, "FilePath", op.FilePath, true); err != nil {
		return err
	}
	if err := encodeOperationField(writer, "BinaryDiff", op.BinaryDiff, true); err != nil {
		return err
	}

	// Encode NewFile data - this is the large binary data that needs streaming
	if err := encodeOperationField(writer, "NewFile", op.NewFile, true); err != nil {
		return err
	}

	if err := encodeOperationField(writer, "OldChecksum", op.OldChecksum, true); err != nil {
		return err
	}
	if err := encodeOperationField(writer, "NewChecksum", op.NewChecksum, true); err != nil {
		return err
	}
	if err := encodeOperationField(writer, "Size", op.Size, false); err != nil {
		return err
	}

	// Write operation closing
	if _, err := writer.Write([]byte("\n    }")); err != nil {
		return err
	}

	return nil
}

// encodeOperationField encodes a single field within an operation
func encodeOperationField(writer io.Writer, name string, value interface{}, addComma bool) error {
	commaStr := ","
	if !addComma {
		commaStr = ""
	}

	fieldStr := fmt.Sprintf("      \"%s\": ", name)
	if _, err := writer.Write([]byte(fieldStr)); err != nil {
		return err
	}

	// For byte slices (binary data), encode as base64 using streaming encoder
	if byteData, ok := value.([]byte); ok {
		// Write opening quote
		if _, err := writer.Write([]byte("\"")); err != nil {
			return err
		}

		// Create base64 encoder that writes directly to output
		encoder := base64.NewEncoder(base64.StdEncoding, writer)

		// Write data in chunks to avoid memory exhaustion
		const chunkSize = 64 * 1024 // 64KB chunks
		for i := 0; i < len(byteData); i += chunkSize {
			end := i + chunkSize
			if end > len(byteData) {
				end = len(byteData)
			}
			if _, err := encoder.Write(byteData[i:end]); err != nil {
				encoder.Close()
				return err
			}
		}

		// Close encoder to flush any remaining data
		if err := encoder.Close(); err != nil {
			return err
		}

		// Write closing quote and comma
		if _, err := writer.Write([]byte(fmt.Sprintf("\"%s", commaStr))); err != nil {
			return err
		}
	} else {
		// For other types, use JSON encoding
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		jsonStr := string(data) + commaStr
		if _, err := writer.Write([]byte(jsonStr)); err != nil {
			return err
		}
	}

	if _, err := writer.Write([]byte("\n")); err != nil {
		return err
	}

	return nil
}
