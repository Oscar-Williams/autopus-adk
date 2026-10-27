# SPEC-REVMODEL-001 acceptance fixtures

`acceptance.md`의 시나리오가 쓰는 fixture다. 값은 그 파일의 oracle과 같이 읽는다.

- R3: `platforms: [claude-code, codex, omp]`. `orchestra.judge: claude`. `orchestra.providers`의 claude·codex·gemini는 모두 `backend: omp`, `subprocess.timeout: 900`이고 model은 차례로 `anthropic/claude-fable-5-1:max`, `openai-codex/gpt-6-astra:max`, `google-antigravity/gemini-3.1-pro:high`다. `orchestra.commands`의 brainstorm·plan·review·secure는 모두 `providers: [claude, codex, gemini]`다. `spec.review_gate`는 `enabled: true`, `strategy: debate`, `providers: [claude, codex, gemini]`, `judge: claude`, `max_revisions: 2`다. 이 branch의 `auto update --yes`로 한 번 맞춘 bytes를 B0라 한다.
- RH: R3와 같은 값을 손으로 쓴 파일. 들여쓰기 2칸, `spec:` 바로 위에 주석 줄 `# review settings`, 구역 사이 빈 줄, `providers: [claude, codex, gemini] # reviewers`. retired key는 없다.
- RL: R3에서 `orchestra.providers`만 이 branch의 기본 항목(`DefaultClaudeProviderEntry`, `DefaultCodexProviderEntry`, gemini `binary: agy`)으로 바꾼 것. 셋 다 `backend`가 없다. RL을 이 branch의 `auto update --yes`로 한 번 맞춘 bytes를 B1이라 한다. CATC: Codex catalog(`codex debug models` JSON). `gpt-6-astra`·`gpt-6-luna`는 `low`–`max` 다섯 단계, `gpt-6.1-sol`은 `low`·`medium`·`high`·`xhigh`(SPEC-MODELUX-003 CAT-E와 같다). RB: R3의 `spec.review_gate.providers: [claude]`. RN: R3에서 `spec.review_gate.judge` 줄을 지운 것. RE: R3의 그 줄이 `judge: ""`인 것. RD: R3의 `enabled: false`.
- CATG: strict catalog. 보고 thinking은 `anthropic/claude-fable-5-1` [low, medium, high, max], `anthropic/claude-opus-5-5` [high, max], `openai-codex/gpt-6-astra` [low, medium, high, xhigh, max], `openai-codex/gpt-6-luna` [low, medium, high], `google-antigravity/gemini-3.1-pro` [low, high], `google-antigravity/gemini-3.8-flash` [low, medium, high], `example/foo-1` [high]. 정규화 뒤 순서는 정렬 순서다(예: luna는 `high`, `low`, `medium`).
- CATN: CATG에서 `google-antigravity/gemini-3.1-pro`를 뺀 것. CATE: CATG에 `thinking: []`인 `example/bar-1`을 더한 것(`catalog_invalid`). CATX: `omp models`가 오류를 내는 runner.
- FB 응답: claude·codex와 `google-antigravity/gemini-3.8-flash:high`인 gemini는 output `AUTOPUS_PROVIDER_SMOKE_OK`, `google-antigravity/gemini-3.1-pro:high`인 gemini는 error `HTTP 404 Not Found (Cloud Code Assist)`. FB는 요청마다 Provider, Prompt, Timeout, `Config.Model`을 기록한다.
- AGY: `agyModelsRunner`가 내는 관측 sample(agy 1.4.0, 2026-10-16) 다섯 줄. `\t`는 탭 문자다.

```text
Fetching available models...
gemini-3.8-flash-high\tGemini 3.8 Flash (High)
gemini-4-argon-medium\tGemini 4 Argon (Medium)
claude-opus-5-5-low\tClaude Opus 5.5 (Low)
gpt-oss-120b-medium\tGPT-OSS 120B (Medium)
```
