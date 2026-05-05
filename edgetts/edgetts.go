package edgetts

import (
	"bytes"
	"context"
	"io"
)

func Synthesize(ctx context.Context, text string, opts Options) ([]byte, error) {
	var buf bytes.Buffer
	if err := WriteSpeech(ctx, &buf, text, opts); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func WriteSpeech(ctx context.Context, w io.Writer, text string, opts Options) error {
	if w == nil {
		return io.ErrClosedPipe
	}
	cfg, err := normalizeOptions(opts)
	if err != nil {
		return err
	}
	chunks, err := prepareTextChunks(text)
	if err != nil {
		return err
	}
	if len(chunks) == 0 {
		return ErrNoAudioReceived
	}
	return writeSpeechChunks(ctx, w, chunks, cfg)
}
