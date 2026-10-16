# Probe: CLI backend model (2026-10-16, feat/revmodel-001 8ec13935)

Model 호출 없음. throwaway test `TestZZRevmodelProbe3`(실행 뒤 삭제)와 `--help`, v0.50.125 바이너리 load만 썼다.

## 코드 경로

- argv: `pkg/orchestra/subprocess_runner.go:111-127` `buildSubprocessArgs`는 `Config.Args`를 복사하고 schema flag와 prompt만 더한다. `pkg/orchestra/command.go:31`이 실행한다. CLI 경로는 `ProviderConfig.Model`을 쓰지 않는다(grep).
- 설정 → ProviderConfig: `resolveProviders`(`internal/cli/orchestra_config.go:92`)와 `providerConfigFromEntry`(`orchestra_helpers.go:221`). `auto spec review`는 `resolveSpecReviewProviderConfigs` → `resolveProviders(&cfg.Orchestra, "review", names)`(`spec_review_readonly.go:163-167`). 설정 없는 `auto orchestra run`만 `buildProviderConfigsForRuntime`(`orchestra_run.go:118`)으로 기본 항목을 쓴다.
- claude: 기본 args `--print --model claude-fable-5-1 --effort max`(`pkg/config/claude_provider.go:10, 36-47`). 런타임 전역 `--effort`는 `upsertClaudeEffortArg`로 effort만 덮는다(`orchestra_run_runtime.go:21-26`). 등급 model은 `ClaudeModelForTier`(`quality_tier.go:8-11, 104-115`).
- codex: `CodexOrchestraProfile`(`codex_profile.go:93-95`, astra·max)은 `model_policy: quality`일 때만 load migration(`codex_provider.go:200-214`)과 런타임(`orchestra_run_runtime.go:27-35`)이 `ApplyCodexProviderProfile`(`codex_provider.go:24-27`)로 args의 `-m`·effort를 덮는다. `pinned`는 v0.50.66 자동 pin argv와 같을 때만 되돌린다.
- gemini: 설정된 binary는 `agy`, args `--print ""`(`defaults.go:119`). `agy --help`: `--model  Model for the current CLI session`, `models  List available models`(`-m` 짧은 형은 없음). `gemini --help`: `-m, --model  Model [string]`. Autopus에는 agy model 목록 해석기가 없다.
- key: `ProviderEntry`의 `model`·`args`·`model_policy`·`backend`는 v0.50.125 schema에 있다. `validOMPModelSelector`는 `backend: omp`일 때만 `model`을 검사한다(`schema_provider_backend.go:41-46`).

## 실행 결과 (`loadHarnessConfigForFlags` → `applyRuntimeHarnessOverrides(--effort low)` → `assembleSpecReviewProviders`)

```text
pinned claude backend="" binary="claude" args=["--print" "--model" "claude-opus-5-5" "--effort" "low" "--permission-mode" "plan" …]
pinned codex backend="" binary="codex" args=["exec" "--json" "--sandbox" "read-only" "-m" "gpt-6-luna" "-c" "model_reasoning_effort=\"high\"" "--ephemeral" …]
pinned gemini backend="" binary="agy" args=["--print" "" "--mode" "plan" "--sandbox" "--disable-slash-commands"]
quality codex backend="" binary="codex" args=["exec" "--json" "--sandbox" "read-only" "-m" "gpt-6-astra" "-c" "model_reasoning_effort=\"low\"" …]
switch gemini backend="omp" binary="agy" model="google-antigravity/gemini-3.8-flash:high" args=["--print" ""]
```

- pinned codex는 쓴 `-m gpt-6-luna`와 effort `high`가 load·런타임·read-only 투영 뒤에도 남는다. quality codex는 `-m gpt-6-astra`로 덮인다(effort는 `--effort low`).
- claude `--model claude-opus-5-5`는 남고 `--effort`만 전역 flag가 덮는다.
- `backend: omp` + `model`로 바꾼 gemini는 `binary`·`args`가 남아도 OMP backend로 조립되고 read-only 정책을 지난다.

## v0.50.125 load

- `pinned.yaml`(codex `model_policy: pinned` + `-m gpt-6-luna` args, claude `--model claude-opus-5-5` args): `auto config get usage_profile` exit 0(`developer`)
- `switch.yaml`(gemini `backend: omp` + `model`, binary·args 남음): exit 0(`developer`)
