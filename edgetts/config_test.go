package edgetts

import "testing"

func TestNormalizeOptionsDefaultsAndVoiceCanonicalization(t *testing.T) {
	cfg, err := normalizeOptions(Options{})
	if err != nil {
		t.Fatal(err)
	}
	wantVoice := "Microsoft Server Speech Text to Speech Voice (en-US, EmmaMultilingualNeural)"
	if cfg.voice != wantVoice {
		t.Fatalf("voice = %q, want %q", cfg.voice, wantVoice)
	}
	if cfg.rate != "+0%" || cfg.volume != "+0%" || cfg.pitch != "+0Hz" {
		t.Fatalf("unexpected defaults: %#v", cfg)
	}
}

func TestCanonicalVoiceNameWithRegionSuffix(t *testing.T) {
	got := canonicalVoiceName("zh-CN-liaoning-XiaobeiNeural")
	want := "Microsoft Server Speech Text to Speech Voice (zh-CN-liaoning, XiaobeiNeural)"
	if got != want {
		t.Fatalf("canonicalVoiceName() = %q, want %q", got, want)
	}
}

func TestNormalizeOptionsRejectsInvalidParams(t *testing.T) {
	tests := []Options{
		{Voice: "bad"},
		{Rate: "0%"},
		{Volume: "+x%"},
		{Pitch: "+1%"},
	}
	for _, opts := range tests {
		if _, err := normalizeOptions(opts); err == nil {
			t.Fatalf("normalizeOptions(%#v) returned nil error", opts)
		}
	}
}
