package edgetts

import (
	"fmt"
	"time"
)

func makeSSML(cfg ttsConfig, escapedText []byte) string {
	return "<speak version='1.0' xmlns='http://www.w3.org/2001/10/synthesis' xml:lang='en-US'>" +
		fmt.Sprintf("<voice name='%s'>", cfg.voice) +
		fmt.Sprintf("<prosody pitch='%s' rate='%s' volume='%s'>", cfg.pitch, cfg.rate, cfg.volume) +
		string(escapedText) +
		"</prosody></voice></speak>"
}

func dateToString(t time.Time) string {
	return t.UTC().Format("Mon Jan 02 2006 15:04:05 GMT+0000 (Coordinated Universal Time)")
}

func ssmlHeadersPlusData(requestID, timestamp, ssml string) string {
	return "X-RequestId:" + requestID + "\r\n" +
		"Content-Type:application/ssml+xml\r\n" +
		"X-Timestamp:" + timestamp + "Z\r\n" +
		"Path:ssml\r\n\r\n" +
		ssml
}
