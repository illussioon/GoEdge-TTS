//go:build integration

package edgetts

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func TestIntegrationSynthesize(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	audio, err := Synthesize(ctx, "hello", Options{Voice: DefaultVoice})
	if err != nil {
		t.Fatal(err)
	}
	if len(audio) == 0 {
		t.Fatal("empty audio")
	}
	if !bytes.HasPrefix(audio, []byte("ID3")) && !hasMPEGFrameSync(audio[:min(len(audio), 32)]) {
		t.Fatalf("audio does not look like mp3: first bytes % x", audio[:min(len(audio), 8)])
	}
}

func hasMPEGFrameSync(data []byte) bool {
	for i := 0; i+1 < len(data); i++ {
		if data[i] == 0xff && data[i+1]&0xe0 == 0xe0 {
			return true
		}
	}
	return false
}
