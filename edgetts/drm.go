package edgetts

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const winEpochSeconds = 11644473600

var (
	clockSkewMu sync.Mutex
	clockSkew   time.Duration
)

func generateSecMSGEC() string {
	return generateSecMSGECAt(time.Now().UTC().Add(currentClockSkew()))
}

func generateSecMSGECAt(t time.Time) string {
	seconds := t.UTC().Unix() + winEpochSeconds
	seconds -= seconds % 300
	ticks := seconds * 10_000_000
	payload := fmt.Sprintf("%d%s", ticks, trustedClientToken)
	sum := sha256.Sum256([]byte(payload))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}

func generateMUID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(buf)), nil
}

func randomHexID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func currentClockSkew() time.Duration {
	clockSkewMu.Lock()
	defer clockSkewMu.Unlock()
	return clockSkew
}

func adjustClockSkewFromResponse(resp *http.Response) error {
	if resp == nil {
		return ErrSkewAdjustmentFailed
	}
	dateHeader := resp.Header.Get("Date")
	if dateHeader == "" {
		return ErrSkewAdjustmentFailed
	}
	serverTime, err := http.ParseTime(dateHeader)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSkewAdjustmentFailed, err)
	}

	clockSkewMu.Lock()
	defer clockSkewMu.Unlock()
	clientTime := time.Now().UTC().Add(clockSkew)
	clockSkew += serverTime.Sub(clientTime)
	return nil
}

func headersWithMUID(headers map[string]string) (http.Header, error) {
	muid, err := generateMUID()
	if err != nil {
		return nil, err
	}
	out := make(http.Header, len(headers)+1)
	for key, value := range headers {
		out.Set(key, value)
	}
	out.Set("Cookie", "muid="+muid+";")
	return out, nil
}
