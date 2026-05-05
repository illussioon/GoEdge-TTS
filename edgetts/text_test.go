package edgetts

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestPrepareTextChunksEscapesAndCleans(t *testing.T) {
	chunks, err := prepareTextChunks("a\x0bb & <c>")
	if err != nil {
		t.Fatal(err)
	}
	got := string(bytes.Join(chunks, nil))
	want := "a b &amp; &lt;c&gt;"
	if got != want {
		t.Fatalf("prepared text = %q, want %q", got, want)
	}
}

func TestSplitTextByByteLengthKeepsUTF8Valid(t *testing.T) {
	text := []byte(strings.Repeat("ж", 10))
	chunks, err := splitTextByByteLength(text, 5)
	if err != nil {
		t.Fatal(err)
	}
	for _, chunk := range chunks {
		if len(chunk) > 5 {
			t.Fatalf("chunk length = %d, want <= 5", len(chunk))
		}
		if !utf8.Valid(chunk) {
			t.Fatalf("invalid utf8 chunk %q", chunk)
		}
	}
}

func TestSplitTextByByteLengthDoesNotSplitXMLEntity(t *testing.T) {
	chunks, err := splitTextByByteLength([]byte("hello &amp; world"), 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, chunk := range chunks {
		if bytes.Contains(chunk, []byte("&am")) && !bytes.Contains(chunk, []byte("&amp;")) {
			t.Fatalf("split inside entity: %q", chunk)
		}
	}
}
