# LLM Module Guide

Simple guide for using AI completions and embeddings in your modules.

## Setup

Add to your `.env`:

```bash
OPENAI_API_KEY=your-api-key-here
OPENAI_MODEL=gpt-4
```

## Usage in Your Module

### 1. Inject the LLM Client

```go
import "github.com/moasq/go-b2b-starter/pkg/llm/domain"

type YourService struct {
    llmClient domain.LLMClient
}

func NewYourService(llmClient domain.LLMClient) *YourService {
    return &YourService{llmClient: llmClient}
}
```

### 2. Use Completions (Prompts)

Send a prompt, get AI-generated text:

```go
func (s *YourService) GenerateText(ctx context.Context, prompt string) (string, error) {
    req := domain.CompletionRequest{
        Prompt: prompt,
    }

    response, err := s.llmClient.Complete(ctx, req)
    if err != nil {
        return "", err
    }

    return response.Text, nil
}
```

With options:

```go
maxTokens := 200
temperature := float32(0.7) // 0.0 = focused, 1.0 = creative

req := domain.CompletionRequest{
    Prompt:      prompt,
    MaxTokens:   &maxTokens,
    Temperature: &temperature,
}
```

### 3. Use Embeddings (Vectors)

Convert text to vectors for semantic search:

```go
func (s *DocumentService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
    embedding, err := s.llmClient.GenerateEmbedding(
        ctx,
        text,
        "text-embedding-3-small",
    )
    if err != nil {
        return nil, err
    }

    // Convert []float64 to []float32 for database
    result := make([]float32, len(embedding))
    for i, v := range embedding {
        result[i] = float32(v)
    }

    return result, nil
}
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `OPENAI_API_KEY` | *required* | Your OpenAI API key |
| `OPENAI_MODEL` | `gpt-4` | AI model |
| `OPENAI_MAX_TOKENS` | `500` | Max response length |
| `OPENAI_TEMPERATURE` | `0.7` | Creativity level (0.0-1.0) |
| `LLM_TIMEOUT_SEC` | `60` | Request timeout |

## Common Models

**Completions:** `gpt-4`, `gpt-4-turbo`, `gpt-3.5-turbo`
**Embeddings:** `text-embedding-3-small` (recommended), `text-embedding-3-large`

That's it! Just inject `LLMClient` and you're ready to use AI in your module.
