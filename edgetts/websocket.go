package edgetts

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type metadataEnvelope struct {
	Metadata []metadataObject `json:"Metadata"`
}

type metadataObject struct {
	Type string       `json:"Type"`
	Data metadataData `json:"Data"`
}

type metadataData struct {
	Offset   int64        `json:"Offset"`
	Duration int64        `json:"Duration"`
	Text     metadataText `json:"text"`
}

type metadataText struct {
	Text string `json:"Text"`
}

func writeSpeechChunks(ctx context.Context, w io.Writer, chunks [][]byte, cfg ttsConfig) error {
	var cumulativeAudioBytes int64
	for _, chunk := range chunks {
		written, err := writeSpeechChunkWithRetry(ctx, w, chunk, cfg)
		if err != nil {
			return err
		}
		cumulativeAudioBytes += written
		_ = cumulativeAudioBytes * 8 * ticksPerSecond / mp3BitrateBPS
	}
	return nil
}

func writeSpeechChunkWithRetry(ctx context.Context, w io.Writer, chunk []byte, cfg ttsConfig) (int64, error) {
	written, resp, err := writeSpeechChunk(ctx, w, chunk, cfg)
	if err == nil {
		return written, nil
	}
	if resp == nil || resp.StatusCode != http.StatusForbidden {
		return written, err
	}
	if skewErr := adjustClockSkewFromResponse(resp); skewErr != nil {
		return written, skewErr
	}
	written, _, err = writeSpeechChunk(ctx, w, chunk, cfg)
	return written, err
}

func writeSpeechChunk(ctx context.Context, w io.Writer, chunk []byte, cfg ttsConfig) (int64, *http.Response, error) {
	conn, resp, err := dialWebsocket(ctx, cfg)
	if err != nil {
		return 0, resp, err
	}
	defer conn.Close()

	if err := conn.SetWriteDeadline(time.Now().Add(cfg.connectTimeout)); err != nil {
		return 0, resp, err
	}
	if err := conn.WriteMessage(websocket.TextMessage, []byte(speechConfigMessage())); err != nil {
		return 0, resp, err
	}

	requestID, err := randomHexID()
	if err != nil {
		return 0, resp, err
	}
	ssml := makeSSML(cfg, chunk)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(ssmlHeadersPlusData(requestID, dateToString(time.Now()), ssml))); err != nil {
		return 0, resp, err
	}

	var audioReceived bool
	var audioBytes int64
	for {
		if err := ctx.Err(); err != nil {
			return audioBytes, resp, err
		}
		if err := conn.SetReadDeadline(time.Now().Add(cfg.receiveTimeout)); err != nil {
			return audioBytes, resp, err
		}

		messageType, data, err := conn.ReadMessage()
		if err != nil {
			return audioBytes, resp, err
		}

		switch messageType {
		case websocket.TextMessage:
			headers, body, err := parseHeadersAndData(data, lenBeforeSeparator(data))
			if err != nil {
				return audioBytes, resp, err
			}
			switch headers["Path"] {
			case "audio.metadata":
				if err := validateMetadata(body); err != nil {
					return audioBytes, resp, err
				}
			case "turn.end":
				if !audioReceived {
					return audioBytes, resp, ErrNoAudioReceived
				}
				return audioBytes, resp, nil
			case "response", "turn.start":
			default:
				return audioBytes, resp, fmt.Errorf("%w: unknown text path %q", ErrUnknownResponse, headers["Path"])
			}
		case websocket.BinaryMessage:
			payload, ok, err := parseAudioPayload(data)
			if err != nil {
				return audioBytes, resp, err
			}
			if !ok {
				continue
			}
			n, err := w.Write(payload)
			if err != nil {
				return audioBytes, resp, err
			}
			if n != len(payload) {
				return audioBytes, resp, io.ErrShortWrite
			}
			audioReceived = true
			audioBytes += int64(len(payload))
		case websocket.CloseMessage:
			return audioBytes, resp, fmt.Errorf("%w: websocket closed", ErrUnexpectedResponse)
		default:
			return audioBytes, resp, fmt.Errorf("%w: websocket message type %d", ErrUnexpectedResponse, messageType)
		}
	}
}

func dialWebsocket(ctx context.Context, cfg ttsConfig) (*websocket.Conn, *http.Response, error) {
	connectionID, err := randomHexID()
	if err != nil {
		return nil, nil, err
	}
	headers, err := headersWithMUID(websocketHeaders())
	if err != nil {
		return nil, nil, err
	}
	dialer := websocket.Dialer{
		HandshakeTimeout: cfg.connectTimeout,
		Proxy:            http.ProxyFromEnvironment,
		NetDialContext: (&net.Dialer{
			Timeout: cfg.connectTimeout,
		}).DialContext,
		EnableCompression: true,
	}
	if cfg.proxyURL != "" {
		proxyURL, err := url.Parse(cfg.proxyURL)
		if err != nil {
			return nil, nil, err
		}
		dialer.Proxy = http.ProxyURL(proxyURL)
	}

	url := fmt.Sprintf("%s&ConnectionId=%s&Sec-MS-GEC=%s&Sec-MS-GEC-Version=%s", wssURL, connectionID, generateSecMSGEC(), secMSGECVersion)
	return dialer.DialContext(ctx, url, headers)
}

func speechConfigMessage() string {
	return "X-Timestamp:" + dateToString(time.Now()) + "\r\n" +
		"Content-Type:application/json; charset=utf-8\r\n" +
		"Path:speech.config\r\n\r\n" +
		`{"context":{"synthesis":{"audio":{"metadataoptions":{"sentenceBoundaryEnabled":"true","wordBoundaryEnabled":"false"},"outputFormat":"audio-24khz-48kbitrate-mono-mp3"}}}}` + "\r\n"
}

func parseAudioPayload(data []byte) ([]byte, bool, error) {
	if len(data) < 2 {
		return nil, false, fmt.Errorf("%w: binary message missing header length", ErrUnexpectedResponse)
	}
	headerLength := int(binary.BigEndian.Uint16(data[:2]))
	if headerLength > len(data)-2 {
		return nil, false, fmt.Errorf("%w: header length exceeds message length", ErrUnexpectedResponse)
	}
	headers, payload, err := parseHeadersAndData(data[2:], headerLength)
	if err != nil {
		return nil, false, err
	}
	if headers["Path"] != "audio" {
		return nil, false, fmt.Errorf("%w: binary path %q", ErrUnexpectedResponse, headers["Path"])
	}
	contentType, hasContentType := headers["Content-Type"]
	if !hasContentType {
		if len(payload) == 0 {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("%w: binary data without content type", ErrUnexpectedResponse)
	}
	if contentType != "audio/mpeg" {
		return nil, false, fmt.Errorf("%w: content type %q", ErrUnexpectedResponse, contentType)
	}
	if len(payload) == 0 {
		return nil, false, fmt.Errorf("%w: empty audio payload", ErrUnexpectedResponse)
	}
	return payload, true, nil
}

func parseHeadersAndData(data []byte, headerLength int) (map[string]string, []byte, error) {
	if headerLength < 0 || headerLength > len(data) {
		return nil, nil, fmt.Errorf("%w: invalid header length", ErrUnexpectedResponse)
	}
	headers := map[string]string{}
	for _, line := range strings.Split(string(data[:headerLength]), "\r\n") {
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return nil, nil, fmt.Errorf("%w: malformed header line", ErrUnexpectedResponse)
		}
		headers[key] = value
	}
	bodyStart := headerLength
	if len(data) >= headerLength+2 && string(data[headerLength:headerLength+2]) == "\r\n" {
		bodyStart += 2
	}
	return headers, data[bodyStart:], nil
}

func lenBeforeSeparator(data []byte) int {
	idx := strings.Index(string(data), "\r\n\r\n")
	if idx == -1 {
		return len(data)
	}
	return idx
}

func validateMetadata(data []byte) error {
	var envelope metadataEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return err
	}
	for _, item := range envelope.Metadata {
		switch item.Type {
		case "WordBoundary", "SentenceBoundary", "SessionEnd":
		default:
			return fmt.Errorf("%w: metadata type %q", ErrUnknownResponse, item.Type)
		}
	}
	return nil
}
