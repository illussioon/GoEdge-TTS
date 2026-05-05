package edgetts

import (
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"
)

var xmlEscaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
)

func prepareTextChunks(text string) ([][]byte, error) {
	cleaned := removeIncompatibleCharacters(text)
	escaped := xmlEscaper.Replace(cleaned)
	return splitTextByByteLength([]byte(escaped), maxTextChunkBytes)
}

func removeIncompatibleCharacters(text string) string {
	return strings.Map(func(r rune) rune {
		if (0 <= r && r <= 8) || (11 <= r && r <= 12) || (14 <= r && r <= 31) {
			return ' '
		}
		return r
	}, text)
}

func splitTextByByteLength(text []byte, byteLength int) ([][]byte, error) {
	if byteLength <= 0 {
		return nil, fmt.Errorf("byte length must be positive")
	}

	var chunks [][]byte
	for len(text) > byteLength {
		splitAt := bytes.LastIndexByte(text[:byteLength], '\n')
		if splitAt < 0 {
			splitAt = bytes.LastIndexByte(text[:byteLength], ' ')
		}
		if splitAt < 0 {
			splitAt = safeUTF8SplitPoint(text[:byteLength])
		}
		splitAt = adjustSplitPointForXMLEntity(text, splitAt)
		if splitAt <= 0 {
			return nil, fmt.Errorf("maximum byte length is too small near an XML entity or UTF-8 sequence")
		}

		chunk := bytes.TrimSpace(text[:splitAt])
		if len(chunk) > 0 {
			chunks = append(chunks, append([]byte(nil), chunk...))
		}
		text = text[splitAt:]
	}

	remaining := bytes.TrimSpace(text)
	if len(remaining) > 0 {
		chunks = append(chunks, append([]byte(nil), remaining...))
	}
	return chunks, nil
}

func safeUTF8SplitPoint(text []byte) int {
	for splitAt := len(text); splitAt > 0; splitAt-- {
		if utf8.Valid(text[:splitAt]) {
			return splitAt
		}
	}
	return 0
}

func adjustSplitPointForXMLEntity(text []byte, splitAt int) int {
	for splitAt > 0 && bytes.Contains(text[:splitAt], []byte("&")) {
		ampersandIndex := bytes.LastIndexByte(text[:splitAt], '&')
		if bytes.IndexByte(text[ampersandIndex:splitAt], ';') != -1 {
			break
		}
		splitAt = ampersandIndex
	}
	return splitAt
}
