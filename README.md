# SimpleLLM Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/SimpleLLM/sdk-go.svg)](https://pkg.go.dev/github.com/SimpleLLM/sdk-go)
[![CI](https://github.com/SimpleLLM/sdk-go/actions/workflows/ci.yml/badge.svg)](https://github.com/SimpleLLM/sdk-go/actions/workflows/ci.yml)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Official Go SDK for the [SimpleLLM](https://simplellm.eu) API — EU-hosted, OpenAI-compatible LLM inference. Zero third-party dependencies (stdlib only).

## Install

```bash
go get github.com/SimpleLLM/sdk-go
```

Requires Go 1.22+.

## Quick start

### Non-streaming chat

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/SimpleLLM/sdk-go"
)

func main() {
    client, err := simplellm.New() // reads SIMPLELLM_API_KEY from env
    if err != nil {
        log.Fatal(err)
    }

    resp, err := client.Chat(context.Background(), simplellm.ChatRequest{
        Model: "DeepSeek-Chat-V3.1",
        Messages: []simplellm.Message{
            {Role: "user", Content: simplellm.Ptr("Explain Occam's Razor in one sentence.")},
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(*resp.Choices[0].Message.Content)
}
```

### Streaming chat

```go
import "io"

stream, err := client.ChatStream(context.Background(), simplellm.ChatRequest{
    Model: "DeepSeek-Chat-V3.1",
    Messages: []simplellm.Message{
        {Role: "user", Content: simplellm.Ptr("Count to five.")},
    },
})
if err != nil {
    log.Fatal(err)
}
defer stream.Close()

for {
    chunk, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }
    if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != nil {
        fmt.Print(*chunk.Choices[0].Delta.Content)
    }
}
fmt.Println()
```

## Examples

### List models

```go
models, err := client.Models(context.Background())
// models.Data is []simplellm.Model
for _, m := range models.Data {
    fmt.Println(m.ID)
}
```

### Audio transcription

```go
import "os"

audio, _ := os.ReadFile("speech.mp3")
tx, err := client.Transcribe(context.Background(), simplellm.TranscribeRequest{
    File:     audio,
    Filename: "speech.mp3",
    Model:    "whisper-large-v3",
    Language: "de",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(tx.Text)
```

### Text-to-speech

```go
audio, err := client.Speech(context.Background(), simplellm.SpeechRequest{
    Input: "Hallo Welt!",
    Model: "tts-1",
    Voice: "nova",
})
if err != nil {
    log.Fatal(err)
}
os.WriteFile("output.mp3", audio, 0o644)
```

### Image generation

```go
imgs, err := client.GenerateImage(context.Background(), simplellm.ImageRequest{
    Prompt: "A minimalist logo for an AI company",
    N:      simplellm.Ptr(1),
    Size:   "1024x1024",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(imgs.Data[0].URL)
```

### Account usage

```go
usage, err := client.Usage(context.Background())
if err != nil {
    log.Fatal(err)
}
if usage.Balance != nil {
    fmt.Printf("Balance: %.4f SC\n", *usage.Balance)
}
```

### API key management

```go
// List all keys
keys, err := client.ListKeys(context.Background())

// Current key info
current, err := client.CurrentKey(context.Background())

// Aggregate usage for current key
stats, err := client.CurrentKeyUsage(context.Background())

// Per-day breakdown for current key
daily, err := client.CurrentKeyDailyUsage(context.Background())

// By key ID
stats, err = client.KeyUsage(context.Background(), "key-id")
daily, err = client.KeyDailyUsage(context.Background(), "key-id")
```

## Configuration

```go
import (
    "net/http"
    "time"

    "github.com/SimpleLLM/sdk-go"
)

client, err := simplellm.New(
    simplellm.WithAPIKey("sk-..."),                      // override env
    simplellm.WithBaseURL("https://api.simplellm.eu"),   // default
    simplellm.WithTimeout(60*time.Second),               // default 120s
    simplellm.WithHTTPClient(&http.Client{...}),         // custom transport
)
```

Environment variables (read before options, options override):

| Variable | Default | Description |
|---|---|---|
| `SIMPLELLM_API_KEY` | — | Required if `WithAPIKey` is not used |
| `SIMPLELLM_BASE_URL` | `https://api.simplellm.eu` | Override the base URL |

## Error handling

All API errors are returned as `*simplellm.Error`:

```go
import "errors"

resp, err := client.Chat(ctx, req)
if err != nil {
    var apiErr *simplellm.Error
    if errors.As(err, &apiErr) {
        fmt.Printf("HTTP %d  code=%s  message=%s\n",
            apiErr.Status, apiErr.Code, apiErr.Message)
    }
    log.Fatal(err)
}
```

## SDKs for other languages

| Language | Package | Repo |
|---|---|---|
| Node.js / TypeScript | `@simplellm/sdk` | [sdk-js](https://github.com/SimpleLLM/sdk-js) |
| Rust | `simplellm` | [sdk-rs](https://github.com/SimpleLLM/sdk-rs) |
| Go | `github.com/SimpleLLM/sdk-go` | [sdk-go](https://github.com/SimpleLLM/sdk-go) |
| C++ | `simplellm` | [sdk-cpp](https://github.com/SimpleLLM/sdk-cpp) |

## License

[MIT](LICENSE) — Copyright (c) 2026 SimpleLLM / Stage One Solutions UG
