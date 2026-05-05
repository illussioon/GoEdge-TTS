# GoEdge-TTS

[English](README.md) | [Русский](README.ru.md) | [Українська](README.uk.md)

A small Go library for Microsoft Edge online Text-to-Speech. Pass text and a voice, get MP3 audio back.

GoEdge-TTS ports the core behavior of the Python `edge-tts` package without requiring Python, an HTTP server, or playback wrappers.

## Features

- Text-to-speech through the Microsoft Edge TTS WebSocket API.
- Voice selection by `ShortName`, for example `en-US-EmmaMultilingualNeural`.
- Fetching the available voice list.
- Writing audio to an `io.Writer` or returning audio as `[]byte`.
- `rate`, `volume`, and `pitch` options.
- Optional proxy URL support.
- Unit tests and an optional integration test against the real Edge TTS service.

## Installation

### Add as a dependency

In your Go project, run:

```bash
go get github.com/illussioon/GoEdge-TTS
```

Import the package:

```go
import "github.com/illussioon/GoEdge-TTS/edgetts"
```

Example `go.mod`:

```go
module myapp

go 1.22

require github.com/illussioon/GoEdge-TTS v0.0.0
```

### Clone the repository

```bash
git clone https://github.com/illussioon/GoEdge-TTS.git
cd GoEdge-TTS
go test ./...
```

## Quick demo

The repository includes a console demo:

```bash
go run ./cmd/voice-demo
```

The demo:

1. loads and prints the voice list;
2. asks you to choose a voice by number or `ShortName`;
3. asks you to enter text;
4. creates `input.mp3`.

Microsoft Edge TTS returns MP3 audio, so the demo creates `input.mp3`, not WAV.

## Library examples

### Get audio as `[]byte`

```go
package main

import (
    "context"
    "os"
    "time"

    "github.com/illussioon/GoEdge-TTS/edgetts"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()

    audio, err := edgetts.Synthesize(ctx, "Hello from Go", edgetts.Options{
        Voice: edgetts.DefaultVoice,
    })
    if err != nil {
        panic(err)
    }

    if err := os.WriteFile("output.mp3", audio, 0644); err != nil {
        panic(err)
    }
}
```

### Write audio directly to a file

```go
package main

import (
    "context"
    "os"
    "time"

    "github.com/illussioon/GoEdge-TTS/edgetts"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()

    file, err := os.Create("output.mp3")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    err = edgetts.WriteSpeech(ctx, file, "Hello! This is a TTS test.", edgetts.Options{
        Voice:  "en-US-EmmaMultilingualNeural",
        Rate:   "+0%",
        Volume: "+0%",
        Pitch:  "+0Hz",
    })
    if err != nil {
        panic(err)
    }
}
```

### List voices

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/illussioon/GoEdge-TTS/edgetts"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    voices, err := edgetts.ListVoices(ctx, "")
    if err != nil {
        panic(err)
    }

    for _, voice := range voices {
        fmt.Printf("%s %s %s\n", voice.ShortName, voice.Gender, voice.FriendlyName)
    }
}
```

## API

### `Options`

```go
type Options struct {
    Voice          string
    Rate           string
    Volume         string
    Pitch          string
    ProxyURL       string
    ConnectTimeout time.Duration
    ReceiveTimeout time.Duration
}
```

Defaults:

| Field | Default |
|---|---|
| `Voice` | `en-US-EmmaMultilingualNeural` |
| `Rate` | `+0%` |
| `Volume` | `+0%` |
| `Pitch` | `+0Hz` |
| `ConnectTimeout` | `10s` |
| `ReceiveTimeout` | `60s` |

Parameter format:

- `Rate`: `+10%`, `-20%`, `+0%`
- `Volume`: `+10%`, `-20%`, `+0%`
- `Pitch`: `+10Hz`, `-20Hz`, `+0Hz`

You can pass a short voice name:

```text
en-US-EmmaMultilingualNeural
ru-RU-SvetlanaNeural
uk-UA-PolinaNeural
```

Internally it is converted to the Microsoft Speech Service format:

```text
Microsoft Server Speech Text to Speech Voice (en-US, EmmaMultilingualNeural)
```

### `Synthesize`

```go
func Synthesize(ctx context.Context, text string, opts Options) ([]byte, error)
```

Synthesizes the full text and returns MP3 audio in memory.

### `WriteSpeech`

```go
func WriteSpeech(ctx context.Context, w io.Writer, text string, opts Options) error
```

Synthesizes text and writes MP3 audio to the provided `io.Writer`. Prefer this for larger text because it does not keep the whole result in memory.

### `ListVoices`

```go
func ListVoices(ctx context.Context, proxyURL string) ([]Voice, error)
```

Loads the available Microsoft Edge TTS voices.

## How it works

### 1. Text preparation

```text
input text
  ↓
remove unsupported control chars
  ↓
XML escape: &, <, >
  ↓
split into chunks <= 4096 UTF-8 bytes
```

The text splitter avoids breaking:

- UTF-8 characters;
- XML entities like `&amp;`;
- words when a nearby space or newline exists.

### 2. Option normalization

```text
Options
  ↓
defaults
  ↓
voice canonicalization
  ↓
regex validation
```

Example:

```text
en-US-EmmaMultilingualNeural
```

becomes:

```text
Microsoft Server Speech Text to Speech Voice (en-US, EmmaMultilingualNeural)
```

### 3. DRM token and headers

Microsoft Edge TTS expects special query parameters and headers:

```text
Sec-MS-GEC
Sec-MS-GEC-Version
TrustedClientToken
Cookie: muid=<random>
```

`Sec-MS-GEC` is generated like this:

```text
current unix time
  + Windows epoch offset
  round down to 5 minutes
  convert to 100ns ticks
  concatenate with trusted client token
  SHA256
  uppercase hex
```

If the server returns `403`, the library reads the HTTP `Date` header, adjusts clock skew, and retries once.

### 4. WebSocket synthesis

For each text chunk, the library opens a WebSocket connection:

```text
wss://speech.platform.bing.com/consumer/speech/synthesize/readaloud/edge/v1
```

It sends two text frames:

```text
Path:speech.config
```

and

```text
Path:ssml
```

The SSML contains the selected voice, rate, volume, pitch, and prepared text.

### 5. MP3 output

The server sends:

- text frames: `response`, `turn.start`, `audio.metadata`, `turn.end`;
- binary frames: MP3 audio chunks.

Binary frame layout:

```text
[2 bytes header length][headers][\r\n][mp3 payload]
```

The library checks:

```text
Path:audio
Content-Type:audio/mpeg
```

and writes only the MP3 payload to the `io.Writer`.

## Tests

Unit tests:

```bash
go test ./...
```

Integration test with a real Microsoft Edge TTS request:

```bash
go test -tags=integration ./...
```

The integration test requires internet access and depends on Microsoft Edge TTS availability.

## Project structure

```text
.
├── cmd/
│   └── voice-demo/
│       └── main.go
├── edgetts/
│   ├── config.go
│   ├── constants.go
│   ├── drm.go
│   ├── edgetts.go
│   ├── errors.go
│   ├── ssml.go
│   ├── text.go
│   ├── voices.go
│   ├── websocket.go
│   └── *_test.go
├── go.mod
├── go.sum
├── README.md
├── README.ru.md
└── README.uk.md
```

## Limitations

- Edge TTS is an unofficial Microsoft Edge API, so the protocol can change.
- The current output format is MP3: `audio-24khz-48kbitrate-mono-mp3`.
- WAV is not generated directly. If you need WAV, convert MP3 separately, for example with `ffmpeg`.
