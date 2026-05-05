# GoEdge-TTS

[English](README.md) | [Русский](README.ru.md) | [Українська](README.uk.md)

Невелика Go-бібліотека для Microsoft Edge online Text-to-Speech: передаєте текст і голос, отримуєте MP3-аудіо.

GoEdge-TTS переносить основний функціонал Python-пакета `edge-tts`, але без Python-залежностей, HTTP-сервера та playback-обгорток.

## Можливості

- Синтез тексту в мовлення через Microsoft Edge TTS WebSocket API.
- Вибір голосу за `ShortName`, наприклад `en-US-EmmaMultilingualNeural`.
- Отримання списку доступних голосів.
- Запис результату в `io.Writer` або отримання аудіо як `[]byte`.
- Підтримка параметрів `rate`, `volume`, `pitch`.
- Опціональна підтримка proxy URL.
- Unit tests і optional integration test з реальним запитом до Edge TTS.

## Встановлення

### Додати як залежність

У своєму Go-проєкті виконайте:

```bash
go get github.com/illussioon/GoEdge-TTS
```

Імпортуйте пакет так:

```go
import "github.com/illussioon/GoEdge-TTS/edgetts"
```

Приклад `go.mod` вашого застосунку:

```go
module myapp

go 1.22

require github.com/illussioon/GoEdge-TTS v0.0.0
```

### Клонувати репозиторій

```bash
git clone https://github.com/illussioon/GoEdge-TTS.git
cd GoEdge-TTS
go test ./...
```

## Швидкий запуск demo

У репозиторії є консольний приклад:

```bash
go run ./cmd/voice-demo
```

Він робить таке:

1. завантажує і виводить список голосів;
2. просить вибрати голос за номером або `ShortName`;
3. просить ввести текст;
4. створює файл `input.mp3`.

Важливо: Microsoft Edge TTS повертає MP3, тому приклад створює `input.mp3`, а не WAV.

## Приклади використання бібліотеки

### Отримати аудіо як `[]byte`

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

### Записати аудіо напряму у файл

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

    err = edgetts.WriteSpeech(ctx, file, "Привіт! Це тест синтезу мовлення.", edgetts.Options{
        Voice:  "uk-UA-PolinaNeural",
        Rate:   "+0%",
        Volume: "+0%",
        Pitch:  "+0Hz",
    })
    if err != nil {
        panic(err)
    }
}
```

### Отримати список голосів

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

| Поле | Default |
|---|---|
| `Voice` | `en-US-EmmaMultilingualNeural` |
| `Rate` | `+0%` |
| `Volume` | `+0%` |
| `Pitch` | `+0Hz` |
| `ConnectTimeout` | `10s` |
| `ReceiveTimeout` | `60s` |

Формат параметрів:

- `Rate`: `+10%`, `-20%`, `+0%`
- `Volume`: `+10%`, `-20%`, `+0%`
- `Pitch`: `+10Hz`, `-20Hz`, `+0Hz`

Голос можна передавати коротким ім'ям:

```text
en-US-EmmaMultilingualNeural
ru-RU-SvetlanaNeural
uk-UA-PolinaNeural
```

Всередині бібліотека перетворює його у формат Microsoft Speech Service:

```text
Microsoft Server Speech Text to Speech Voice (en-US, EmmaMultilingualNeural)
```

### `Synthesize`

```go
func Synthesize(ctx context.Context, text string, opts Options) ([]byte, error)
```

Синтезує весь текст і повертає MP3-аудіо в пам'яті.

### `WriteSpeech`

```go
func WriteSpeech(ctx context.Context, w io.Writer, text string, opts Options) error
```

Синтезує текст і пише MP3-аудіо в переданий `io.Writer`. Це кращий варіант для великих текстів, бо він не тримає весь результат у пам'яті.

### `ListVoices`

```go
func ListVoices(ctx context.Context, proxyURL string) ([]Voice, error)
```

Завантажує список доступних голосів Microsoft Edge TTS.

## Як це працює всередині

### 1. Підготовка тексту

```text
input text
  ↓
remove unsupported control chars
  ↓
XML escape: &, <, >
  ↓
split into chunks <= 4096 UTF-8 bytes
```

Розділення тексту не ламає:

- UTF-8 символи;
- XML entities на кшталт `&amp;`;
- слова, якщо поруч є пробіл або перенос рядка.

### 2. Нормалізація налаштувань

```text
Options
  ↓
defaults
  ↓
voice canonicalization
  ↓
regex validation
```

Наприклад:

```text
en-US-EmmaMultilingualNeural
```

стає:

```text
Microsoft Server Speech Text to Speech Voice (en-US, EmmaMultilingualNeural)
```

### 3. DRM token / headers

Microsoft Edge TTS очікує спеціальні query parameters і headers:

```text
Sec-MS-GEC
Sec-MS-GEC-Version
TrustedClientToken
Cookie: muid=<random>
```

`Sec-MS-GEC` рахується так:

```text
current unix time
  + Windows epoch offset
  round down to 5 minutes
  convert to 100ns ticks
  concatenate with trusted client token
  SHA256
  uppercase hex
```

Якщо сервер відповідає `403`, бібліотека читає HTTP header `Date`, коригує clock skew і повторює запит один раз.

### 4. WebSocket synthesis

Для кожного text chunk відкривається WebSocket:

```text
wss://speech.platform.bing.com/consumer/speech/synthesize/readaloud/edge/v1
```

Бібліотека відправляє два text frame:

```text
Path:speech.config
```

і

```text
Path:ssml
```

SSML містить вибраний голос, rate, volume, pitch і підготовлений текст.

### 5. Отримання MP3

Сервер надсилає:

- text frames: `response`, `turn.start`, `audio.metadata`, `turn.end`;
- binary frames: MP3 audio chunks.

Binary frame має таку структуру:

```text
[2 bytes header length][headers][\r\n][mp3 payload]
```

Бібліотека перевіряє:

```text
Path:audio
Content-Type:audio/mpeg
```

і пише тільки MP3 payload в `io.Writer`.

## Тести

Unit tests:

```bash
go test ./...
```

Integration test з реальним запитом до Microsoft Edge TTS:

```bash
go test -tags=integration ./...
```

Integration test потребує інтернету і залежить від доступності Microsoft Edge TTS.

## Структура проєкту

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

## Обмеження

- Edge TTS — неофіційний API Microsoft Edge, тому протокол може змінитися.
- Поточний формат виходу — MP3: `audio-24khz-48kbitrate-mono-mp3`.
- WAV напряму не генерується. Якщо потрібен WAV, конвертуйте MP3 окремо, наприклад через `ffmpeg`.
