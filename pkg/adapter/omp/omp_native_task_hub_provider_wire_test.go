package omp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func ompNativeRequestHasAuth(header http.Header) bool {
	for name, values := range header {
		lower := strings.ToLower(name)
		if lower == "authorization" || strings.Contains(lower, "api-key") {
			for _, value := range values {
				if strings.TrimSpace(value) != "" {
					return true
				}
			}
		}
	}
	return false
}

// ompNativeLastToolContent returns the text of the newest tool result in the
// conversation as the model reads it: a plain string, or the text parts of a
// content array joined, so callers can match omp's rendering verbatim.
func ompNativeLastToolContent(messages []json.RawMessage) string {
	for index := len(messages) - 1; index >= 0; index-- {
		var message struct {
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
		}
		if json.Unmarshal(messages[index], &message) != nil || message.Role != "tool" {
			continue
		}
		var text string
		if json.Unmarshal(message.Content, &text) == nil {
			return text
		}
		var parts []struct {
			Text string `json:"text"`
		}
		if json.Unmarshal(message.Content, &parts) == nil {
			texts := make([]string, 0, len(parts))
			for _, part := range parts {
				if part.Text != "" {
					texts = append(texts, part.Text)
				}
			}
			return strings.Join(texts, "\n")
		}
		return string(message.Content)
	}
	return ""
}

func ompNativeRequireTokens(value string, tokens ...string) error {
	for _, token := range tokens {
		if !strings.Contains(value, token) {
			return fmt.Errorf("missing %q", token)
		}
	}
	return nil
}

func writeOMPNativeToolSSE(w http.ResponseWriter, sequence int, name string, args any) {
	arguments, _ := json.Marshal(args)
	chunk := map[string]any{
		"id": "chatcmpl-native", "object": "chat.completion.chunk", "created": sequence, "model": ompLiveModel,
		"choices": []any{map[string]any{
			"index": 0,
			"delta": map[string]any{"role": "assistant", "tool_calls": []any{map[string]any{
				"index": 0, "id": fmt.Sprintf("native-call-%d", sequence), "type": "function",
				"function": map[string]any{"name": name, "arguments": string(arguments)},
			}}},
		}},
	}
	terminal := map[string]any{
		"id": "chatcmpl-native", "object": "chat.completion.chunk", "created": sequence, "model": ompLiveModel,
		"choices": []any{map[string]any{"index": 0, "delta": map[string]any{}, "finish_reason": "tool_calls"}},
		"usage":   map[string]int{"prompt_tokens": 10, "completion_tokens": 2, "total_tokens": 12},
	}
	writeOMPSSE(w, chunk, terminal)
}

func writeOMPNativeFinalSSE(w http.ResponseWriter, sequence int, content string) {
	chunk := map[string]any{
		"id": "chatcmpl-native", "object": "chat.completion.chunk", "created": sequence, "model": ompLiveModel,
		"choices": []any{map[string]any{
			"index": 0, "delta": map[string]any{"role": "assistant", "content": content},
		}},
	}
	terminal := map[string]any{
		"id": "chatcmpl-native", "object": "chat.completion.chunk", "created": sequence, "model": ompLiveModel,
		"choices": []any{map[string]any{"index": 0, "delta": map[string]any{}, "finish_reason": "stop"}},
		"usage":   map[string]int{"prompt_tokens": 10, "completion_tokens": 2, "total_tokens": 12},
	}
	writeOMPSSE(w, chunk, terminal)
}
