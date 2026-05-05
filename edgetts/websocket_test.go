package edgetts

import (
	"encoding/binary"
	"testing"
)

func TestParseAudioPayload(t *testing.T) {
	headers := []byte("Path:audio\r\nContent-Type:audio/mpeg")
	payload := []byte{0x49, 0x44, 0x33}
	message := make([]byte, 2, 2+len(headers)+2+len(payload))
	binary.BigEndian.PutUint16(message[:2], uint16(len(headers)))
	message = append(message, headers...)
	message = append(message, '\r', '\n')
	message = append(message, payload...)

	got, ok, err := parseAudioPayload(message)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("payload was not audio")
	}
	if string(got) != string(payload) {
		t.Fatalf("payload = %v, want %v", got, payload)
	}
}

func TestParseAudioPayloadAllowsEmptyTerminator(t *testing.T) {
	headers := []byte("Path:audio")
	message := make([]byte, 2, 2+len(headers)+2)
	binary.BigEndian.PutUint16(message[:2], uint16(len(headers)))
	message = append(message, headers...)
	message = append(message, '\r', '\n')

	_, ok, err := parseAudioPayload(message)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("empty terminator reported as audio")
	}
}
