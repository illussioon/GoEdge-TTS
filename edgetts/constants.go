package edgetts

const (
	baseURL            = "speech.platform.bing.com/consumer/speech/synthesize/readaloud"
	trustedClientToken = "6A5AA1D4EAFF4E9FB37E23D68491D6F4"
	wssURL             = "wss://" + baseURL + "/edge/v1?TrustedClientToken=" + trustedClientToken
	voiceListURL       = "https://" + baseURL + "/voices/list?trustedclienttoken=" + trustedClientToken

	DefaultVoice         = "en-US-EmmaMultilingualNeural"
	chromiumFullVersion  = "143.0.3650.75"
	chromiumMajorVersion = "143"
	secMSGECVersion      = "1-" + chromiumFullVersion

	maxTextChunkBytes = 4096
	ticksPerSecond    = 10_000_000
	mp3BitrateBPS     = 48_000
)

func baseHeaders() map[string]string {
	return map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" + chromiumMajorVersion + ".0.0.0 Safari/537.36 Edg/" + chromiumMajorVersion + ".0.0.0",
		"Accept-Encoding": "gzip, deflate, br, zstd",
		"Accept-Language": "en-US,en;q=0.9",
	}
}

func websocketHeaders() map[string]string {
	headers := baseHeaders()
	headers["Pragma"] = "no-cache"
	headers["Cache-Control"] = "no-cache"
	headers["Origin"] = "chrome-extension://jdiccldimpdaibmpdkjnbmckianbfold"
	return headers
}

func voiceHeaders() map[string]string {
	headers := baseHeaders()
	headers["Authority"] = "speech.platform.bing.com"
	headers["Sec-CH-UA"] = `" Not;A Brand";v="99", "Microsoft Edge";v="` + chromiumMajorVersion + `", "Chromium";v="` + chromiumMajorVersion + `"`
	headers["Sec-CH-UA-Mobile"] = "?0"
	headers["Accept"] = "*/*"
	headers["Sec-Fetch-Site"] = "none"
	headers["Sec-Fetch-Mode"] = "cors"
	headers["Sec-Fetch-Dest"] = "empty"
	return headers
}
