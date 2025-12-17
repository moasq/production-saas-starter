package infra

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/moasq/go-b2b-starter/pkg/llm/domain"
	loggerDomain "github.com/moasq/go-b2b-starter/pkg/logger/domain"
)

type Config struct {
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float32
	TimeoutSec  int
	MaxRetries  int
	DebugMode   bool
}

func (c Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	if c.Model == "" {
		return fmt.Errorf("model is required")
	}
	return nil
}

// CircuitBreaker implements a simple circuit breaker pattern
type CircuitBreaker struct {
	mu              sync.RWMutex
	failureCount    int64
	successCount    int64
	lastFailureTime time.Time
	state           string // "closed", "open", "half-open"
	maxFailures     int
	resetTimeout    time.Duration
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        "closed",
	}
}

// CanExecute checks if a request can be executed based on circuit breaker state
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	if cb.state == "closed" {
		return true
	}
	
	if cb.state == "open" {
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = "half-open"
			cb.mu.Unlock()
			cb.mu.RLock()
			return true
		}
		return false
	}
	
	// half-open state - allow one request to test
	return true
}

// RecordSuccess records a successful execution
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	atomic.AddInt64(&cb.successCount, 1)
	if cb.state == "half-open" {
		cb.state = "closed"
		atomic.StoreInt64(&cb.failureCount, 0)
	}
}

// RecordFailure records a failed execution
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	atomic.AddInt64(&cb.failureCount, 1)
	cb.lastFailureTime = time.Now()
	
	if int(atomic.LoadInt64(&cb.failureCount)) >= cb.maxFailures {
		cb.state = "open"
	}
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	return map[string]interface{}{
		"state":         cb.state,
		"failures":      atomic.LoadInt64(&cb.failureCount),
		"successes":     atomic.LoadInt64(&cb.successCount),
		"last_failure":  cb.lastFailureTime,
	}
}

type OpenAIClient struct {
	config         Config
	client         *http.Client
	logger         loggerDomain.Logger
	circuitBreaker *CircuitBreaker
}

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature *float32        `json:"temperature,omitempty"`
	Stop        []string        `json:"stop,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // "function"
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type openAIMessage struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	Refusal   string     `json:"refusal,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type openAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []openAIChoice `json:"choices"`
	Usage   *openAIUsage   `json:"usage,omitempty"`
	Error   *openAIError   `json:"error,omitempty"`
}

type openAIChoice struct {
	Index        int           `json:"index"`
	Message      openAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

type CompletionTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

type openAIUsage struct {
	PromptTokens            int                      `json:"prompt_tokens"`
	CompletionTokens        int                      `json:"completion_tokens"`
	TotalTokens             int                      `json:"total_tokens"`
	CachedTokens            int                      `json:"cached_tokens,omitempty"`
	CompletionTokensDetails *CompletionTokensDetails `json:"completion_tokens_details,omitempty"`
}

type openAIError struct {
	Message string      `json:"message"`
	Type    string      `json:"type"`
	Param   any `json:"param"` // can be string or null
	Code    any `json:"code"`  // can be string, number, or null
}

func NewLLMConfig() Config {
	maxTokens, _ := strconv.Atoi(getEnvOrDefault("OPENAI_MAX_TOKENS", "150"))
	temperature, _ := strconv.ParseFloat(getEnvOrDefault("OPENAI_TEMPERATURE", "0.1"), 32)
	timeoutSec, _ := strconv.Atoi(getEnvOrDefault("LLM_TIMEOUT_SEC", "60")) // Increased default for GPT-5
	maxRetries, _ := strconv.Atoi(getEnvOrDefault("LLM_MAX_RETRIES", "2"))
	debugMode, _ := strconv.ParseBool(getEnvOrDefault("LLM_DEBUG_MODE", "false"))

	return Config{
		APIKey:      os.Getenv("OPENAI_API_KEY"),
		Model:       getEnvOrDefault("OPENAI_MODEL", "gpt-5-mini"),
		MaxTokens:   maxTokens,
		Temperature: float32(temperature),
		TimeoutSec:  timeoutSec,
		MaxRetries:  maxRetries,
		DebugMode:   debugMode,
	}
}

func NewOpenAIClient(config Config, logger loggerDomain.Logger) (domain.LLMClient, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Configure transport with keep-alive and proper timeouts
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		ForceAttemptHTTP2:   true,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	// No global timeout - let per-request context control deadlines
	client := &http.Client{
		Timeout:   0,
		Transport: transport,
	}

	// Initialize circuit breaker if enabled
	var circuitBreaker *CircuitBreaker
	if os.Getenv("LLM_CIRCUIT_BREAKER_ENABLED") == "true" {
		maxFailures := 3 // Default failure threshold
		resetTimeout := 30 * time.Second // Default reset timeout
		
		if val := os.Getenv("LLM_CIRCUIT_BREAKER_MAX_FAILURES"); val != "" {
			if parsed, err := strconv.Atoi(val); err == nil {
				maxFailures = parsed
			}
		}
		
		if val := os.Getenv("LLM_CIRCUIT_BREAKER_RESET_TIMEOUT"); val != "" {
			if parsed, err := time.ParseDuration(val); err == nil {
				resetTimeout = parsed
			}
		}
		
		circuitBreaker = NewCircuitBreaker(maxFailures, resetTimeout)
		logger.Info("Circuit breaker enabled for OpenAI client", map[string]interface{}{
			"max_failures":   maxFailures,
			"reset_timeout":  resetTimeout,
		})
	}

	return &OpenAIClient{
		config:         config,
		client:         client,
		logger:         logger,
		circuitBreaker: circuitBreaker,
	}, nil
}

func (c *OpenAIClient) Complete(ctx context.Context, request domain.CompletionRequest) (*domain.CompletionResponse, error) {
	if request.Prompt == "" {
		return nil, domain.ErrInvalidPrompt
	}

	maxTokens := c.config.MaxTokens
	if request.MaxTokens != nil {
		maxTokens = *request.MaxTokens
	}

	// Right-size tokens for field extraction - avoid excessive budgets
	if strings.HasPrefix(c.config.Model, "gpt-5") {
		// For GPT-5, use smaller budgets unless explicitly requested
		if maxTokens == c.config.MaxTokens && maxTokens > 200 {
			maxTokens = 128 // Reasonable default for most extraction tasks
			if c.config.DebugMode {
				fmt.Println("[DEBUG] Using optimized token budget for GPT-5:", maxTokens)
			}
		}
	}

	temperature := c.config.Temperature
	if request.Temperature != nil {
		temperature = *request.Temperature
	}

	openAIReq := openAIRequest{
		Model: c.config.Model,
		Messages: []openAIMessage{
			{
				Role:    "user",
				Content: request.Prompt,
			},
		},
		MaxTokens: maxTokens,
		Stream:    false, // Default to non-streaming for backward compatibility
	}

	// Only set temperature for models that support it (GPT-5 models don't accept custom temperature)
	if supportsTemperature(c.config.Model) {
		openAIReq.Temperature = &temperature
	}

	// Only set stop sequences for models that support them (GPT-5 models don't accept stop parameter)
	if supportsStop(c.config.Model) {
		openAIReq.Stop = []string{"\n\n", "\n---"}
	}

	// Enhanced request logging
	if c.config.DebugMode {
		logData := map[string]any{
			"endpoint":              "https://api.openai.com/v1/chat/completions",
			"model":                 c.config.Model,
			"input_length":          len(request.Prompt),
			"max_tokens":           maxTokens,
			"supports_temperature":  supportsTemperature(c.config.Model),
			"supports_stop":         supportsStop(c.config.Model),
		}
		if supportsTemperature(c.config.Model) {
			logData["temperature"] = temperature
		}
		if supportsStop(c.config.Model) {
			logData["stop_sequences"] = []string{"\n\n", "\n---"}
		}
		c.logger.Info("Starting OpenAI request", logData)

		debugMsg := fmt.Sprintf("[DEBUG] Starting OpenAI request - Model: %s | MaxTokens: %d", c.config.Model, maxTokens)
		if supportsTemperature(c.config.Model) {
			debugMsg += fmt.Sprintf(" | Temperature: %.1f", temperature)
		} else {
			debugMsg += " | Temperature: OMITTED"
		}
		if supportsStop(c.config.Model) {
			debugMsg += " | Stop: [\\n\\n, \\n---]"
		} else {
			debugMsg += " | Stop: OMITTED"
		}
		fmt.Println(debugMsg)
	}

	var response *domain.CompletionResponse
	var err error

	// Check circuit breaker before attempting requests
	if c.circuitBreaker != nil && !c.circuitBreaker.CanExecute() {
		stats := c.circuitBreaker.GetStats()
		c.logger.Warn("Circuit breaker is open, request blocked", map[string]any{
			"model":         c.config.Model,
			"breaker_state": stats["state"],
			"failures":      stats["failures"],
			"successes":     stats["successes"],
		})
		return nil, fmt.Errorf("circuit breaker is open due to repeated failures")
	}

	// Retry with fresh context per attempt and exponential backoff with jitter
	for i := 0; i <= c.config.MaxRetries; i++ {
		// Create fresh context per attempt - THIS FIXES THE MAIN BUG
		callTimeout := time.Duration(c.config.TimeoutSec) * time.Second
		if strings.HasPrefix(c.config.Model, "gpt-5") {
			callTimeout += 30 * time.Second // Extra time for reasoning models
		}
		callCtx, cancel := context.WithTimeout(ctx, callTimeout)
		
		response, err = c.makeRequest(callCtx, openAIReq)
		cancel() // Always cancel to free resources
		
		if err == nil {
			// Record success in circuit breaker
			if c.circuitBreaker != nil {
				c.circuitBreaker.RecordSuccess()
			}
			break
		}

		// Categorize error and decide on retry strategy
		isTemp := isTemporaryError(err)
		isPerm := isPermanentError(err)
		
		// Only record failure in circuit breaker for temporary errors
		// Permanent errors (like invalid API key) shouldn't trip the breaker
		if c.circuitBreaker != nil && isTemp {
			c.circuitBreaker.RecordFailure()
		}

		// Don't retry permanent errors
		if isPerm {
			c.logger.Error("Permanent error detected, not retrying", map[string]any{
				"model":       c.config.Model,
				"error":       err.Error(),
				"error_type":  "permanent",
				"attempt":     i + 1,
			})
			break
		}

		if i < c.config.MaxRetries {
			c.logger.Warn("OpenAI request failed, retrying", map[string]any{
				"attempt":     i + 1,
				"max_retries": c.config.MaxRetries,
				"model":       c.config.Model,
				"error":       err.Error(),
				"error_type":  map[bool]string{true: "temporary", false: "unknown"}[isTemp],
				"will_retry":  true,
			})
			
			// Exponential backoff with jitter
			backoff := time.Duration(1<<i) * time.Second
			jitter := time.Duration(generateJitter(int64(backoff))) * time.Millisecond
			time.Sleep(backoff + jitter)
		}
	}

	if err != nil {
		c.logger.Error("OpenAI request failed after all retries", map[string]any{
			"error":       err.Error(),
			"model":       c.config.Model,
			"endpoint":    "https://api.openai.com/v1/chat/completions",
			"max_retries": c.config.MaxRetries,
		})
		fmt.Println("[ERROR] OpenAI request failed after all retries:", err.Error(), "Model:", c.config.Model)
		return nil, err
	}

	return response, nil
}

func (c *OpenAIClient) CompleteStream(ctx context.Context, request domain.CompletionRequest, callback func(domain.StreamChunk) error) (*domain.CompletionResponse, error) {
	if request.Prompt == "" {
		return nil, domain.ErrInvalidPrompt
	}

	maxTokens := c.config.MaxTokens
	if request.MaxTokens != nil {
		maxTokens = *request.MaxTokens
	}

	// Right-size tokens for field extraction - avoid excessive budgets
	if strings.HasPrefix(c.config.Model, "gpt-5") {
		// For GPT-5, use smaller budgets unless explicitly requested
		if maxTokens == c.config.MaxTokens && maxTokens > 200 {
			maxTokens = 128 // Reasonable default for most extraction tasks
			if c.config.DebugMode {
				fmt.Println("[DEBUG] Using optimized token budget for GPT-5 streaming:", maxTokens)
			}
		}
	}

	temperature := c.config.Temperature
	if request.Temperature != nil {
		temperature = *request.Temperature
	}

	openAIReq := openAIRequest{
		Model: c.config.Model,
		Messages: []openAIMessage{
			{
				Role:    "user",
				Content: request.Prompt,
			},
		},
		MaxTokens: maxTokens,
		Stream:    true, // Enable streaming
	}

	// Only set temperature for models that support it
	if supportsTemperature(c.config.Model) {
		openAIReq.Temperature = &temperature
	}

	// Only set stop sequences for models that support them
	if supportsStop(c.config.Model) {
		openAIReq.Stop = []string{"\n\n", "\n---"}
	}

	if c.config.DebugMode {
		c.logger.Info("Starting OpenAI streaming request", map[string]any{
			"model":      c.config.Model,
			"max_tokens": maxTokens,
			"stream":     true,
		})
	}

	var response *domain.CompletionResponse
	var err error

	// Retry with fresh context per attempt
	for i := 0; i <= c.config.MaxRetries; i++ {
		callTimeout := time.Duration(c.config.TimeoutSec) * time.Second
		if strings.HasPrefix(c.config.Model, "gpt-5") {
			callTimeout += 30 * time.Second
		}
		callCtx, cancel := context.WithTimeout(ctx, callTimeout)
		
		response, err = c.makeStreamRequest(callCtx, openAIReq, callback)
		cancel()
		
		if err == nil {
			break
		}

		if i < c.config.MaxRetries {
			c.logger.Warn("OpenAI streaming request failed, retrying", map[string]any{
				"attempt":     i + 1,
				"max_retries": c.config.MaxRetries,
				"model":       c.config.Model,
				"error":       err.Error(),
			})
			
			backoff := time.Duration(1<<i) * time.Second
			jitter := time.Duration(generateJitter(int64(backoff))) * time.Millisecond
			time.Sleep(backoff + jitter)
		}
	}

	if err != nil {
		c.logger.Error("OpenAI streaming request failed after all retries", map[string]any{
			"error":       err.Error(),
			"model":       c.config.Model,
			"max_retries": c.config.MaxRetries,
		})
		return nil, err
	}

	return response, nil
}

func (c *OpenAIClient) makeRequest(ctx context.Context, request openAIRequest) (*domain.CompletionResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code first
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		// Enhanced error logging
		c.logger.Error("OpenAI API returned non-200 status", map[string]any{
			"status_code":   resp.StatusCode,
			"status":        resp.Status,
			"response_body": string(body),
			"model":         c.config.Model,
			"endpoint":      req.URL.String(),
			"content_type":  resp.Header.Get("Content-Type"),
		})
		fmt.Println("[ERROR] OpenAI API non-200 status:", resp.StatusCode, "Body:", string(body), "Model:", c.config.Model)

		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Debug successful response info
	if c.config.DebugMode {
		c.logger.Info("OpenAI API response received", map[string]any{
			"status_code":  resp.StatusCode,
			"content_type": resp.Header.Get("Content-Type"),
		})
	}

	var openAIResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if openAIResp.Error != nil {
		fmt.Println("[ERROR] OpenAI API error response:", openAIResp.Error.Message, "Type:", openAIResp.Error.Type, "Code:", openAIResp.Error.Code, "Param:", openAIResp.Error.Param)
		return nil, fmt.Errorf("OpenAI API error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		fmt.Println("[ERROR] No choices returned from OpenAI API")
		return nil, fmt.Errorf("no choices returned from OpenAI")
	}

	choice := openAIResp.Choices[0]
	msg := choice.Message

	// Handle tool calls (not an error, but we don't support tools for this use case)
	if len(msg.ToolCalls) > 0 {
		fmt.Println("[WARN] Model returned tool calls, but we don't support tools. Treating as error.")
		return nil, fmt.Errorf("model returned tool calls but tools are not supported for this operation")
	}

	// Handle refusal (model refused to respond)
	if strings.TrimSpace(msg.Refusal) != "" {
		fmt.Println("[ERROR] Model refusal:", msg.Refusal)
		return nil, fmt.Errorf("model refusal: %s", msg.Refusal)
	}

	// Only then check for empty content
	if strings.TrimSpace(msg.Content) == "" {
		fmt.Println("[ERROR] Empty content returned from OpenAI API, finish_reason:", choice.FinishReason)
		return nil, fmt.Errorf("empty assistant content (finish_reason=%s)", choice.FinishReason)
	}

	var totalTokens int
	var reasoningTokens int
	var outputTokens int
	if openAIResp.Usage != nil {
		totalTokens = openAIResp.Usage.TotalTokens
		outputTokens = openAIResp.Usage.CompletionTokens
		if openAIResp.Usage.CompletionTokensDetails != nil {
			reasoningTokens = openAIResp.Usage.CompletionTokensDetails.ReasoningTokens
		}
	}

	responseText := msg.Content
	if c.config.DebugMode {
		textPreview := responseText
		if len(textPreview) > 50 {
			textPreview = textPreview[:50] + "..."
		}
		if reasoningTokens > 0 {
			fmt.Printf("[DEBUG] OpenAI response - Text: %s | Total: %d tokens (Reasoning: %d, Output: %d)\n",
				textPreview, totalTokens, reasoningTokens, outputTokens)
		} else {
			fmt.Printf("[DEBUG] OpenAI response - Text: %s | Tokens: %d\n", textPreview, totalTokens)
		}
	}

	return &domain.CompletionResponse{
		Text:       responseText,
		TokensUsed: totalTokens,
		Model:      openAIResp.Model,
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func supportsTemperature(model string) bool {
	// GPT-5 series (gpt-5, gpt-5-mini, gpt-5-nano) don't support custom temperature
	return !strings.HasPrefix(model, "gpt-5")
}

func supportsStop(model string) bool {
	// GPT-5 series don't support `stop` parameter on Chat Completions
	return !strings.HasPrefix(model, "gpt-5")
}

// GenerateEmbedding generates a vector embedding for the given text using OpenAI embeddings API
func (c *OpenAIClient) GenerateEmbedding(ctx context.Context, text string, model string) ([]float64, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	if model == "" {
		model = "text-embedding-3-small" // Default embedding model
	}

	embeddingReq := map[string]any{
		"model": model,
		"input": text,
	}

	jsonData, err := json.Marshal(embeddingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make embedding request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("OpenAI embeddings API returned non-200 status", map[string]any{
			"status_code":   resp.StatusCode,
			"response_body": string(body),
			"model":         model,
		})
		return nil, fmt.Errorf("OpenAI embeddings API error (status %d): %s", resp.StatusCode, string(body))
	}

	var embeddingResp struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
		Error *openAIError `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to decode embedding response: %w", err)
	}

	if embeddingResp.Error != nil {
		return nil, fmt.Errorf("OpenAI embeddings API error: %s", embeddingResp.Error.Message)
	}

	if len(embeddingResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned from OpenAI")
	}

	embedding := embeddingResp.Data[0].Embedding
	if len(embedding) == 0 {
		return nil, fmt.Errorf("empty embedding returned from OpenAI")
	}

	if c.config.DebugMode {
		c.logger.Info("Generated embedding", map[string]any{
			"model":           model,
			"text_length":     len(text),
			"embedding_dims":  len(embedding),
			"tokens_used":     embeddingResp.Usage.TotalTokens,
		})
	}

	return embedding, nil
}

type streamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

func (c *OpenAIClient) makeStreamRequest(ctx context.Context, request openAIRequest, callback func(domain.StreamChunk) error) (*domain.CompletionResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal stream request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create stream request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make stream request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("OpenAI streaming API returned non-200 status", map[string]any{
			"status_code":   resp.StatusCode,
			"response_body": string(body),
		})
		return nil, fmt.Errorf("OpenAI streaming API error (status %d): %s", resp.StatusCode, string(body))
	}

	var fullContent strings.Builder
	var totalTokens int
	var model string
	
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if line == "" || line == "data: [DONE]" {
			continue
		}
		
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		
		jsonStr := strings.TrimPrefix(line, "data: ")
		var streamResp streamResponse
		if err := json.Unmarshal([]byte(jsonStr), &streamResp); err != nil {
			// Skip malformed JSON chunks
			continue
		}
		
		model = streamResp.Model
		
		if len(streamResp.Choices) > 0 {
			choice := streamResp.Choices[0]
			content := choice.Delta.Content
			
			if content != "" {
				fullContent.WriteString(content)
				
				// Call callback with chunk
				if callback != nil {
					if err := callback(domain.StreamChunk{
						Content: content,
						Done:    false,
					}); err != nil {
						return nil, fmt.Errorf("streaming callback error: %w", err)
					}
				}
			}
			
			if choice.FinishReason != nil && *choice.FinishReason != "" {
				// Final chunk
				if callback != nil {
					callback(domain.StreamChunk{
						Content: "",
						Done:    true,
					})
				}
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading stream: %w", err)
	}

	finalContent := fullContent.String()
	if strings.TrimSpace(finalContent) == "" {
		return nil, fmt.Errorf("empty content from streaming response")
	}

	// Estimate token usage (rough approximation)
	totalTokens = len(strings.Fields(finalContent)) + 10 // Add some overhead

	return &domain.CompletionResponse{
		Text:       finalContent,
		TokensUsed: totalTokens,
		Model:      model,
	}, nil
}

// generateJitter creates random jitter for exponential backoff to avoid thundering herd
func generateJitter(maxJitterMs int64) int64 {
	if maxJitterMs <= 0 {
		return 0
	}
	// Generate random number between 0 and maxJitterMs
	n, err := rand.Int(rand.Reader, big.NewInt(maxJitterMs))
	if err != nil {
		return maxJitterMs / 2 // fallback to half the max
	}
	return n.Int64()
}

// isTemporaryError determines if an error is temporary and should be retried
func isTemporaryError(err error) bool {
	errStr := strings.ToLower(err.Error())
	
	// Network-level errors that are typically temporary
	temporaryErrors := []string{
		"connection reset by peer",
		"connection refused",
		"timeout",
		"context deadline exceeded",
		"temporary failure",
		"service unavailable",
		"bad gateway",
		"gateway timeout",
		"too many requests",
		"rate limit",
		"internal server error",
	}
	
	for _, tempErr := range temporaryErrors {
		if strings.Contains(errStr, tempErr) {
			return true
		}
	}
	
	return false
}

// isPermanentError determines if an error is permanent and should not be retried
func isPermanentError(err error) bool {
	errStr := strings.ToLower(err.Error())
	
	// Errors that indicate permanent issues
	permanentErrors := []string{
		"invalid api key",
		"unauthorized",
		"forbidden",
		"not found",
		"bad request",
		"invalid request",
		"model not found",
		"quota exceeded",
		"billing",
	}
	
	for _, permErr := range permanentErrors {
		if strings.Contains(errStr, permErr) {
			return true
		}
	}
	
	return false
}