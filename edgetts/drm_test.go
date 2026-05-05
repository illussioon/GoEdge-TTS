package edgetts

import (
	"testing"
	"time"
)

func TestGenerateSecMSGECAt(t *testing.T) {
	fixed := time.Date(2026, 5, 5, 12, 34, 56, 0, time.UTC)
	got := generateSecMSGECAt(fixed)
	want := "2A326A40FF0F2758715E52274BA3576AD61661C0C79AC588493CE40ECF3393A8"
	if got != want {
		t.Fatalf("generateSecMSGECAt() = %q, want %q", got, want)
	}
}

func TestGenerateMUID(t *testing.T) {
	got, err := generateMUID()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 32 {
		t.Fatalf("MUID length = %d, want 32", len(got))
	}
}
