# GoEdge-TTS

[English](README.md) | [Русский](README.ru.md) | [Українська](README.uk.md)

Небольшая Go-библиотека для Microsoft Edge online Text-to-Speech: передаёте текст и голос, получаете MP3-аудио.

GoEdge-TTS портирует основной функционал Python-пакета `edge-tts`, но без Python-зависимостей, HTTP-сервера и playback-обвязки.

## Возможности

- Синтез текста в речь через Microsoft Edge TTS WebSocket API.
- Выбор голоса по `ShortName`, например `en-US-EmmaMultilingualNeural`.
- Получение списка доступных голосов.
- Запись результата в `io.Writer` или получение аудио как `[]byte`.
- Поддержка параметров `rate`, `volume`, `pitch`.
- Опциональная поддержка proxy URL.
- Unit tests и optional integration test с реальным запросом к Edge TTS.

## Установка

### Установить как зависимость

В своём Go-проекте выполните:

```bash
go get github.com/illussioon/GoEdge-TTS
```

Импортируйте пакет так:

```go
import "github.com/illussioon/GoEdge-TTS/edgetts"
```

Пример `go.mod` вашего приложения:

```go
module myapp

go 1.22

require github.com/illussioon/GoEdge-TTS v0.0.0
```

### Клонировать репозиторий

```bash
git clone https://github.com/illussioon/GoEdge-TTS.git
cd GoEdge-TTS
go test ./...
```

## Быстрый запуск demo

В репозитории есть консольный пример:

```bash
go run ./cmd/voice-demo
```

Он делает следующее:

1. загружает и выводит список голосов;
2. просит выбрать голос по номеру или `ShortName`;
3. просит ввести текст;
4. создаёт файл `input.mp3`.

Важно: Microsoft Edge TTS отдаёт MP3, поэтому пример создаёт `input.mp3`, а не WAV.

## Пример использования библиотеки

### Получить аудио как `[]byte`

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

### Записать аудио напрямую в файл

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

    err = edgetts.WriteSpeech(ctx, file, "Привет! Это тест синтеза речи.", edgetts.Options{
        Voice:  "ru-RU-SvetlanaNeural",
        Rate:   "+0%",
        Volume: "+0%",
        Pitch:  "+0Hz",
    })
    if err != nil {
        panic(err)
    }
}
```

### Получить список голосов

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

Параметры валидируются так же, как в Python `edge-tts`:

- `Rate`: `+10%`, `-20%`, `+0%`
- `Volume`: `+10%`, `-20%`, `+0%`
- `Pitch`: `+10Hz`, `-20Hz`, `+0Hz`

Голос можно передавать коротким именем:

```text
en-US-EmmaMultilingualNeural
ru-RU-SvetlanaNeural
uk-UA-PolinaNeural
```

Внутри библиотека преобразует его в формат Microsoft Speech Service:

```text
Microsoft Server Speech Text to Speech Voice (en-US, EmmaMultilingualNeural)
```

### `Synthesize`

```go
func Synthesize(ctx context.Context, text string, opts Options) ([]byte, error)
```

Синтезирует весь текст и возвращает MP3-аудио в памяти.

### `WriteSpeech`

```go
func WriteSpeech(ctx context.Context, w io.Writer, text string, opts Options) error
```

Синтезирует текст и пишет MP3-аудио в переданный `io.Writer`. Это предпочтительный вариант для больших текстов, потому что не держит весь результат в памяти.

### `ListVoices`

```go
func ListVoices(ctx context.Context, proxyURL string) ([]Voice, error)
```

Загружает список доступных голосов Microsoft Edge TTS.

## Как это работает внутри

### 1. Подготовка текста

```text
input text
  ↓
remove unsupported control chars
  ↓
XML escape: &, <, >
  ↓
split into chunks <= 4096 UTF-8 bytes
```

Текст режется так, чтобы не ломать:

- UTF-8 символы;
- XML entities вроде `&amp;`;
- слова, если рядом есть пробел или перенос строки.

### 2. Нормализация настроек

```text
Options
  ↓
defaults
  ↓
voice canonicalization
  ↓
regex validation
```

Например:

```text
en-US-EmmaMultilingualNeural
```

становится:

```text
Microsoft Server Speech Text to Speech Voice (en-US, EmmaMultilingualNeural)
```

### 3. DRM token / headers

Microsoft Edge TTS требует специальные query parameters и headers:

```text
Sec-MS-GEC
Sec-MS-GEC-Version
TrustedClientToken
Cookie: muid=<random>
```

`Sec-MS-GEC` считается так:

```text
current unix time
  + Windows epoch offset
  round down to 5 minutes
  convert to 100ns ticks
  concatenate with trusted client token
  SHA256
  uppercase hex
```

Если сервер отвечает `403`, библиотека читает HTTP header `Date`, корректирует clock skew и повторяет запрос один раз.

### 4. WebSocket synthesis

Для каждого text chunk открывается WebSocket:

```text
wss://speech.platform.bing.com/consumer/speech/synthesize/readaloud/edge/v1
```

Библиотека отправляет два text frame:

```text
Path:speech.config
```

и

```text
Path:ssml
```

SSML содержит выбранный голос, rate, volume, pitch и подготовленный текст.

### 5. Получение MP3

Сервер присылает:

- text frames: `response`, `turn.start`, `audio.metadata`, `turn.end`;
- binary frames: MP3 audio chunks.

Binary frame устроен так:

```text
[2 bytes header length][headers][\r\n][mp3 payload]
```

Библиотека проверяет:

```text
Path:audio
Content-Type:audio/mpeg
```

и пишет только MP3 payload в `io.Writer`.

## Тесты

Обычные unit tests:

```bash
go test ./...
```

Integration test с реальным запросом к Microsoft Edge TTS:

```bash
go test -tags=integration ./...
```

Integration test требует интернет и может зависеть от доступности Microsoft Edge TTS.

## Структура проекта

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

## Ограничения

- Edge TTS — неофициальный API Microsoft Edge, поэтому протокол может измениться.
- На выходе сейчас MP3: `audio-24khz-48kbitrate-mono-mp3`.
- WAV напрямую не генерируется. Если нужен WAV, MP3 нужно конвертировать отдельно, например через `ffmpeg`.
