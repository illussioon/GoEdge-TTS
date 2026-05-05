package edgetts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type VoiceTag struct {
	ContentCategories  []string `json:"ContentCategories"`
	VoicePersonalities []string `json:"VoicePersonalities"`
}

type Voice struct {
	Name           string   `json:"Name"`
	ShortName      string   `json:"ShortName"`
	Gender         string   `json:"Gender"`
	Locale         string   `json:"Locale"`
	SuggestedCodec string   `json:"SuggestedCodec"`
	FriendlyName   string   `json:"FriendlyName"`
	Status         string   `json:"Status"`
	VoiceTag       VoiceTag `json:"VoiceTag"`
}

func ListVoices(ctx context.Context, proxyURL string) ([]Voice, error) {
	voices, resp, err := listVoices(ctx, proxyURL)
	if err == nil {
		return voices, nil
	}
	if resp == nil || resp.StatusCode != http.StatusForbidden {
		return nil, err
	}
	if skewErr := adjustClockSkewFromResponse(resp); skewErr != nil {
		return nil, skewErr
	}
	voices, _, err = listVoices(ctx, proxyURL)
	return voices, err
}

func listVoices(ctx context.Context, proxyURL string) ([]Voice, *http.Response, error) {
	client, err := httpClient(proxyURL)
	if err != nil {
		return nil, nil, err
	}
	headers, err := headersWithMUID(voiceHeaders())
	if err != nil {
		return nil, nil, err
	}
	requestURL := fmt.Sprintf("%s&Sec-MS-GEC=%s&Sec-MS-GEC-Version=%s", voiceListURL, generateSecMSGEC(), secMSGECVersion)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header = headers

	resp, err := client.Do(req)
	if err != nil {
		return nil, resp, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp, fmt.Errorf("voice list request failed: %s", resp.Status)
	}

	var voices []Voice
	if err := json.NewDecoder(resp.Body).Decode(&voices); err != nil {
		return nil, resp, err
	}
	for i := range voices {
		if voices[i].VoiceTag.ContentCategories == nil {
			voices[i].VoiceTag.ContentCategories = []string{}
		}
		if voices[i].VoiceTag.VoicePersonalities == nil {
			voices[i].VoiceTag.VoicePersonalities = []string{}
		}
	}
	return voices, resp, nil
}

func httpClient(proxyURL string) (*http.Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if proxyURL != "" {
		parsed, err := url.Parse(proxyURL)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(parsed)
	} else {
		transport.Proxy = http.ProxyFromEnvironment
	}
	return &http.Client{Transport: transport}, nil
}
