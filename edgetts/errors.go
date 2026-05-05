package edgetts

import "errors"

var (
	ErrNoAudioReceived      = errors.New("no audio was received")
	ErrUnexpectedResponse   = errors.New("unexpected response")
	ErrUnknownResponse      = errors.New("unknown response")
	ErrSkewAdjustmentFailed = errors.New("failed to adjust clock skew")
)
