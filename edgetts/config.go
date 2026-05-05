package edgetts

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Options struct {
	Voice          string
	Rate           string
	Volume         string
	Pitch          string
	ProxyURL       string
	ConnectTimeout time.Duration
	ReceiveTimeout time.Duration
}

type ttsConfig struct {
	voice          string
	rate           string
	volume         string
	pitch          string
	connectTimeout time.Duration
	receiveTimeout time.Duration
	proxyURL       string
}

var (
	shortVoicePattern = regexp.MustCompile(`^([a-z]{2,})-([A-Z]{2,})-(.+Neural)$`)
	longVoicePattern  = regexp.MustCompile(`^Microsoft Server Speech Text to Speech Voice \(.+,.+\)$`)
	percentPattern    = regexp.MustCompile(`^[+-]\d+%$`)
	pitchPattern      = regexp.MustCompile(`^[+-]\d+Hz$`)
)

func normalizeOptions(opts Options) (ttsConfig, error) {
	voice := opts.Voice
	if voice == "" {
		voice = DefaultVoice
	}
	voice = canonicalVoiceName(voice)
	if !longVoicePattern.MatchString(voice) {
		return ttsConfig{}, fmt.Errorf("invalid voice %q", opts.Voice)
	}

	rate := opts.Rate
	if rate == "" {
		rate = "+0%"
	}
	if !percentPattern.MatchString(rate) {
		return ttsConfig{}, fmt.Errorf("invalid rate %q", rate)
	}

	volume := opts.Volume
	if volume == "" {
		volume = "+0%"
	}
	if !percentPattern.MatchString(volume) {
		return ttsConfig{}, fmt.Errorf("invalid volume %q", volume)
	}

	pitch := opts.Pitch
	if pitch == "" {
		pitch = "+0Hz"
	}
	if !pitchPattern.MatchString(pitch) {
		return ttsConfig{}, fmt.Errorf("invalid pitch %q", pitch)
	}

	connectTimeout := opts.ConnectTimeout
	if connectTimeout == 0 {
		connectTimeout = 10 * time.Second
	}
	receiveTimeout := opts.ReceiveTimeout
	if receiveTimeout == 0 {
		receiveTimeout = 60 * time.Second
	}

	return ttsConfig{
		voice:          voice,
		rate:           rate,
		volume:         volume,
		pitch:          pitch,
		connectTimeout: connectTimeout,
		receiveTimeout: receiveTimeout,
		proxyURL:       opts.ProxyURL,
	}, nil
}

func canonicalVoiceName(voice string) string {
	match := shortVoicePattern.FindStringSubmatch(voice)
	if match == nil {
		return voice
	}

	lang := match[1]
	region := match[2]
	name := match[3]
	if idx := strings.Index(name, "-"); idx != -1 {
		region += "-" + name[:idx]
		name = name[idx+1:]
	}
	return fmt.Sprintf("Microsoft Server Speech Text to Speech Voice (%s-%s, %s)", lang, region, name)
}
