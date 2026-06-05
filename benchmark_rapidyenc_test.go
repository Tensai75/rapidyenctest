package rapidyenctest

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"strings"
	"testing"

	"github.com/mnightingale/rapidyenc"
)

var tests = []struct {
	lines      int
	lineLength int
	bufferSize int
}{
	{lines: 1000, lineLength: 512, bufferSize: -1},
	{lines: 1000, lineLength: 512, bufferSize: 512 * 1024},
	{lines: 1000, lineLength: 512, bufferSize: 1024 * 1024},
	{lines: 1000, lineLength: 512, bufferSize: 1024 * 1024 * 5},
	{lines: 1000, lineLength: 512, bufferSize: 1024 * 1024 * 10},
	{lines: 1000, lineLength: 512, bufferSize: 1024 * 1024 * 10 * 5},
	{lines: 1000, lineLength: 512, bufferSize: 1024 * 1024 * 10 * 10},
	{lines: 1000, lineLength: 512, bufferSize: 1024 * 1024 * 10 * 20},
	{lines: 10000, lineLength: 512, bufferSize: -1},
	{lines: 10000, lineLength: 512, bufferSize: 512 * 1024},
	{lines: 10000, lineLength: 512, bufferSize: 1024 * 1024},
	{lines: 10000, lineLength: 512, bufferSize: 1024 * 1024 * 5},
	{lines: 10000, lineLength: 512, bufferSize: 1024 * 1024 * 10},
	{lines: 10000, lineLength: 512, bufferSize: 1024 * 1024 * 10 * 5},
	{lines: 10000, lineLength: 512, bufferSize: 1024 * 1024 * 10 * 10},
	{lines: 10000, lineLength: 512, bufferSize: 1024 * 1024 * 10 * 20},
	{lines: 100000, lineLength: 512, bufferSize: -1},
	{lines: 100000, lineLength: 512, bufferSize: 512 * 1024},
	{lines: 100000, lineLength: 512, bufferSize: 1024 * 1024},
	{lines: 100000, lineLength: 512, bufferSize: 1024 * 1024 * 5},
	{lines: 100000, lineLength: 512, bufferSize: 1024 * 1024 * 10},
	{lines: 100000, lineLength: 512, bufferSize: 1024 * 1024 * 10 * 5},
	{lines: 100000, lineLength: 512, bufferSize: 1024 * 1024 * 10 * 10},
	{lines: 100000, lineLength: 512, bufferSize: 1024 * 1024 * 10 * 50},
	{lines: 100000, lineLength: 512, bufferSize: 1024 * 1024 * 10 * 100},
}

func BenchmarkRapidyenc(b *testing.B) {
	for _, test := range tests {
		name := fmt.Sprintf("%dx%d", test.lines, test.lineLength)
		if test.bufferSize > 0 {
			name += fmt.Sprintf("_%s", byteCountIEC(test.bufferSize))
		} else {
			name += "_default_buffer"
		}
		input := generateInput(test.lines, test.lineLength)
		b.Run(name, func(b *testing.B) {
			b.SetBytes(int64(test.lines * test.lineLength))
			b.ResetTimer()
			for range b.N {
				b.StopTimer()
				reader := bytes.NewReader(input)
				var decoder *rapidyenc.Decoder
				if test.bufferSize > 0 {
					decoder = rapidyenc.NewDecoder(reader, rapidyenc.WithStatusLineAlreadyRead(), rapidyenc.WithBufferSize(test.bufferSize))
				} else {
					decoder = rapidyenc.NewDecoder(reader, rapidyenc.WithStatusLineAlreadyRead())
				}
				b.StartTimer()
				_, err := decoder.Next()
				if err != nil {
					b.Fatalf("Failed to yenc decode: %v", err)
				}
			}
		})
	}
}

func generateInput(lines, lineLength int) (input []byte) {
	input = append(input, fmt.Sprintf("=ybegin line=%d size=-1\r\n", lineLength)...) // size is not known in advance, so set to -1
	var originalLines []byte
	for range lines {
		input = append(input, []byte(strings.Repeat("x", lineLength)+"\r\n")...)
		originalLines = append(originalLines, []byte(strings.Repeat("N", lineLength))...)
	}
	originalLinesCRC32 := crc32.ChecksumIEEE(originalLines)
	input = append(input, []byte("=yend crc32="+fmt.Sprintf("%08x", originalLinesCRC32)+"\r\n")...)
	input = append(input, []byte(".\r\n")...)
	return input
}

func byteCountIEC(b int) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
