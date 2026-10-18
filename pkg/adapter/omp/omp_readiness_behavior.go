package omp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
)

const (
	ompReadinessNegotiateID   = "readiness-negotiate"
	ompReadinessStateID       = "readiness-state"
	ompReadinessCommandsID    = "readiness-commands"
	ompReadinessSubscribeID   = "readiness-subscribe"
	ompReadinessUnsubscribeID = "readiness-unsubscribe"
	// @AX:NOTE [AUTO]: OMP resolves this selector in its built-in catalog before RPC starts and exits on an unknown one, so it must name a model every supported release ships (omp/17.2.7 has no gpt-6.x); the readiness path never invokes its provider.
	ompReadinessBootstrapModel = "openai-codex/gpt-5.6-sol"
)

// ompReadinessProbeTools are the only tools the provider-free probe enables,
// and get_state must dump each of them. OMP exits on an unknown --tools name
// before RPC starts, so the list holds only tools every supported release
// ships: omp/18.3.0 replaced hub with wait, and earlier releases reject wait.
var ompReadinessProbeTools = []string{"task", "todo"}

func ompProviderFreeRPCInput() []byte {
	frames := []map[string]any{
		{"id": ompReadinessNegotiateID, "type": "negotiate_protocol", "protocolVersion": 2},
		{"id": ompReadinessStateID, "type": "get_state"},
		{"id": ompReadinessCommandsID, "type": "get_available_commands"},
		{"id": ompReadinessSubscribeID, "type": "set_subagent_subscription", "level": "progress"},
		{"id": ompReadinessUnsubscribeID, "type": "set_subagent_subscription", "level": "off"},
	}
	var input bytes.Buffer
	encoder := json.NewEncoder(&input)
	for _, frame := range frames {
		_ = encoder.Encode(frame)
	}
	return input.Bytes()
}

func ompProviderFreeRPCArgs(root string) []string {
	return []string{
		"--mode", "rpc",
		"--no-session",
		"--cwd", root,
		"--model", ompReadinessBootstrapModel,
		"--tools", strings.Join(ompReadinessProbeTools, ","),
		"--no-extensions",
		"--no-rules",
		"--no-lsp",
		"--no-pty",
	}
}

func runOMPReadinessProviderFreeProbe(
	parent context.Context,
	opts OMPReadinessOptions,
	runner commandOMPProbeRunner,
) ompProbeResult {
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, opts.Timeout)
	defer cancel()
	input := ompProviderFreeRPCInput()
	// @AX:NOTE [AUTO]: 2 KiB bounds the deterministic five-frame readiness request before process launch.
	if len(input) > 2048 {
		return ompProbeResult{reason: "output_invalid"}
	}
	output, err := runner.runRPC(
		ctx, opts.Executable, opts.Root, input, ompProviderFreeRPCArgs(opts.Root)...,
	)
	if err != nil {
		return ompProbeResult{output: output, reason: classifyOMPProbeError(ctx, err)}
	}
	if len(output) > opts.MaxOutput {
		return ompProbeResult{reason: "output_oversized"}
	}
	return ompProbeResult{output: output}
}
