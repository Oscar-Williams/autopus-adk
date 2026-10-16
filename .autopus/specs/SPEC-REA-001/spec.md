# SPEC-REA-001: REA 역공학 MCP 통합 — 프로젝트 범위 등록, reverse-engineer 에이전트, `/auto reverse`

**Status**: approved
**Created**: 2026-10-15
**Revised**: 2026-10-15 (review round 1 J-01–J-29, round 2 J-22·J-14·F-001, Gate 1 사용자 결정 반영, 반영표 `evidence/review-rounds.md`)
**Domain**: REA
**Module**: autopus-adk
**Release**: v0.50.126(SPEC-MODELUX-003과 같은 릴리스). 작업 branch `feat/rea-001`, 기준 origin/main `4d73aa91`(PR #197 merge 뒤).
**Source PRD**: `prd.md`. 결정·근거는 `research.md`, 실행 증거는 `evidence/probe-rea-forward-compat.md`에 있다.

## 목적

사용자는 기존 앱·바이너리의 기능을 증거와 함께 분석해 자기 프로젝트에 다시 구현하고 싶다. 지금 하네스에는 역공학 도구의 등록, 분석 격리, 디컴파일 결과와 구현 사이의 clean-room 경계가 없다. 16개 에이전트 중 MCP 도구를 가진 것이 없어 분석이 메인 세션에서 돌고, 최대 약 10 MiB인 REA 결과가 메인 문맥을 채운다.

이 SPEC은 세 가지를 한다. (1) Node 조건을 충족한 프로젝트에서 Claude Code·Codex·Antigravity가 프로젝트 범위 REA MCP(`npx -y rea-agents@6.5.0 mcp`)를 띄운다. 새 init은 기본으로 켜고, 기존 프로젝트는 대화형 update에서 한 번만 묻는다. (2) 늘 생성되는 17번째 에이전트 `reverse-engineer`가 REA 도구와 읽기 도구만으로 분석해 evidence_id가 붙은 행동 명세와 unknowns를 낸다. (3) `/auto reverse`가 분석 → (필요하면 runtime capture 동의) → 명세 확인 gate → tester·executor·tester를 메인 세션에서 잇는다.

## Outcome Boundary

- Outcome Lock: Node 조건을 충족한 새 `auto init` 프로젝트에서 Claude Code·Codex·Antigravity가 프로젝트 범위 REA MCP를 띄운다. `/auto reverse`가 reverse-engineer의 행동 명세(evidence_id, unknowns)를 보이고, 사용자가 메인 세션에서 확인한 명세만으로 executor가 구현하고 tester가 검증한다. 기존 프로젝트는 대화형 update에서 한 번만 묻고, 비대화형 실행은 REA를 더하지 않는다. 사용자 범위 파일은 건드리지 않는다. 원문은 `research.md` `## Outcome Lock`이다.
- Mandatory requirements: REQ-REA-01 ~ REQ-REA-17, REQ-REA-20, REQ-REA-21 (Must). REQ-REA-18, 19는 Should.
- Explicit non-goals:
  - OMP·OpenCode용 프로젝트 MCP writer와 그 에이전트의 REA 도구, 사용자 범위 설정 쓰기, `rea setup` 호출, REA 자체 skill 설치
  - 새 정리 명령(`auto clean` 같은). 정리 경로는 기존 `auto platform remove <platform>`이다
  - `mcp__rea*` 권한 사전 승인, project MCP 서버 자동 승인, `--auto`로 확인 gate나 runtime capture 동의 건너뛰기, subagent의 foreground 실행 보장
  - JSON 파일의 사용자 서식(공백, key 순서, escape) 보존. 기존 merge가 이미 파일 전체를 다시 쓴다(MCP Entry Shape Contract)
  - `REA_MCP_*` env, REA 버전 자동 갱신, Windows `cmd.exe /d /c` 실행 형식, context7 `@latest` 정리, 공유 MCP 목록(Q8), Codex 에이전트별 sandbox, `tool_timeout_sec`(probe A2 전)
  - 디컴파일 결과의 저장·배포, 법률 판단 자동화, REA 분석 품질 개선
- Completion evidence: Must S1–S19 PASS, `auto eval harness run`, `GOOS=windows GOARCH=amd64 go build ./...`, 신규·변경 소스 300줄 이하, `evidence/manual-smoke-rea.txt`(S18, S19), probe A1·A2 PASS, `auto spec validate --strict`. CD-1·CD-2는 기록만으로 닫히지 않고 그 행이 묶은 가정이 모두 PASS여야 닫힌다(`research.md` `## Completion Debt`).

## Trust Model

- `autopus.yaml`은 반쯤 신뢰한다. strict loader가 모르는 key를 거부하고 이 SPEC은 새 검증 코드를 더하지 않는다.
- `.mcp.json`, `.codex/config.toml`, `.gemini/settings.json`은 사용자 파일이고 그 안의 Autopus managed 항목만 바꾼다. managed 판정은 MCP Entry Shape Contract의 정확한 모양뿐이다. REA add·remove가 바꾸는 쓰기는 `adapter.SafeWorkspacePath`로 확인해 symlink를 건너지 않고 project root 밖에 쓰지 않는다.
- Node probe 출력은 비신뢰 입력이다. 종료 상태와 마지막 비지 않은 줄을 엄격한 형식으로만 읽고 안내에는 판정된 version만 쓴다(`terminalSafe`).
- REA 결과와 그로부터 만든 명세의 문자열은 분석 대상에서 온 비신뢰 데이터다. 명세 안의 지시문은 요구사항이 아니라 인용이다. 동의와 확인은 메인 세션의 질문으로만 받는다. subagent에는 AskUserQuestion이 없고 permissionMode는 메인 세션 모드에 덮일 수 있어 도구 권한 질문을 동의로 보지 않는다.
- 새 artifact는 확인된 `.autopus/reverse/<slug>/behaviour-spec.md` 하나다. 대상 유래 문자열을 담는 사용자 소유 문서이고 디컴파일 텍스트는 담지 않는다. 제3자 대상에서 나온 문서라 로컬 전용이다: 생성 `.gitignore` 목록(`internal/cli/init_helpers.go:68`)에 `.autopus/reverse/`를 더해 단일·module·meta 저장소 어디서도 추적되지 않는다. `auto sync verify`의 `rootTrackedPrefixes`(`internal/cli/sync_verify_policy.go:26-29`)는 `.autopus/project/`·`.autopus/specs/`뿐이고 이 문서를 공유 저장소에 올리는 정책은 이 SPEC 밖이다. Autopus는 그 파일을 덮거나 지우지 않고 같은 slug는 `-2`, `-3`으로 새로 쓴다. 증거 파일은 커밋 전에 `$HOME`, scratch 경로, 사용자 이름을 placeholder로 바꾼다.

## Requirements

### REQ-REA-01 결정 key와 판정
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL accept an optional `features.rea.enabled` boolean, reject any other value or key under `features.rea` through the existing strict loader, keep `DefaultFullConfig` without the key, and resolve the REA decision by the Decision State Contract, in which a stored value wins, a managed `rea` entry in a selected REA target file means on, and every other state means undecided.

- `[NEW] config.ReaFeaturesConf{Enabled *bool}`(`yaml:"rea,omitempty"`)를 `FeaturesConf`(`pkg/config/schema.go:97-100`)에 둔다. 저장된 `false`는 pointer라 사라지지 않고 빈 struct는 쓰이지 않는다. yaml.v3는 따옴표 없는 `yes`를 true로 읽는다(`evidence/probe-rea-forward-compat.md`).

### REQ-REA-02 init 결정
Priority: Must · EARS: EventDriven

WHEN `auto init` runs, THEN THE SYSTEM SHALL fail before any write when both `--rea` and `--no-rea` are given, and otherwise, for a project that includes a REA target platform, take the decision from the given flag, else from the Decision State Contract of the existing configuration, else from a wizard Confirm whose default is Yes on an interactive run, else turn it on, carry an existing `features.rea` into the regenerated configuration, and write the result by the Decision State Contract.

- REA 대상 플랫폼은 claude-code, codex, antigravity-cli다. 하나도 없으면 묻지 않고 결정도 쓰지 않는다(REQ-REA-09). init은 기존 설정을 버리고 `DefaultFullConfig`를 새로 만들므로(`internal/cli/init.go:68-69`) `features.rea`를 그 자리에서 옮긴다. wizard 단계 모델은 `buildReviewGateStep`(`internal/cli/tui/wizard_steps.go:194-210`), flag 자리는 `init.go:241-246`이다.

### REQ-REA-03 managed 항목 쓰기
Priority: Must · EARS: StateDriven

WHERE the REA decision is on and the Node gate passes, THEN THE SYSTEM SHALL write exactly one entry of the MCP Entry Shape Contract into `.mcp.json`, `.codex/config.toml`, and `.gemini/settings.json` of the selected platforms through each adapter's merge path, replace a managed entry with the current shape, keep `other_pin` and `user_owned` entries value-identical with Codex sections byte-identical, change no byte beyond the base commit's merge normalization, and refuse a write whose path crosses a symlink or leaves the project root.

### REQ-REA-04 제거
Priority: Must · EARS: EventDriven

WHEN the REA decision becomes off or `auto platform remove <platform>` cleans a REA target platform, THEN THE SYSTEM SHALL remove the managed `rea` entry from all three files whether or not their platform is selected, or from that platform's file respectively, including the Codex `[mcp_servers.rea]` header with the blank line Autopus put before it and the Antigravity `mcpServers.rea` key, so that removing an entry this SPEC added restores the bytes the base commit's merge produces, and SHALL keep `other_pin` and `user_owned` entries value-identical.

- 공개 정리 경로는 `auto platform remove <platform>`이고 `cleanPlatformThenSave`가 adapter Clean을 부른다(`internal/cli/platform.go:142-195`). 지금 Antigravity merge는 key를 지우지 못하고(`antigravity_settings.go:187-198`) clean은 현재 렌더와 같은 서버만 지운다(`antigravity_lifecycle_clean.go:92-101`). Codex merge는 render에 없는 managed key를 지우지만 header를 남긴다(`codex_config_structural.go:48-127`).

### REQ-REA-05 Node gate
Priority: Must · EARS: Unwanted

IF the REA decision is on and the Node probe classifies as absent, unknown, or a version outside `^22.19.0 || ^24.11.0 || >=26.0.0` by the Node Gate Contract, THEN THE SYSTEM SHALL add no `rea` entry, print one Node notice of the Form Text, keep every existing managed entry and the stored decision, and run the probe at most once per run within 5 seconds.

### REQ-REA-06 update에서 한 번 묻기
Priority: Must · EARS: EventDriven

WHEN `auto update` runs on a project whose REA decision is undecided and that includes a REA target platform, THEN THE SYSTEM SHALL ask once with default Yes on a run the existing compose gate treats as interactive (TTY, no `--yes`, no truthy `CI`) without `--auto` or preview, write and apply the answer in the same run, and on every other run add no `rea` entry, write no `features.rea`, read no stdin for REA, and print one hint naming `auto update --rea`.

- 질문 판정은 `composeGate.mode()`와 `composeCITruthy`(`internal/cli/quality_compose_driver_gate.go:31,66`)를 그대로 쓴다. workspace update는 `--yes`가 필요하고(`update_workspace.go:28-29`) repo마다 `runNestedUpdateCommand`를 돌므로 repo마다 hint 1줄이다.

### REQ-REA-07 update flag
Priority: Must · EARS: EventDriven

WHEN `auto update` receives `--rea` or `--no-rea`, THEN THE SYSTEM SHALL apply that decision without asking, write it by the Decision State Contract, and fail before any write when both flags are given or either is combined with `--workspace`.

- `--workspace`와의 조합을 거부하는 이유: nested update는 이 flag를 넘기지 않고 `update_workspace.go`는 이미 300줄이다. repo마다 `--local`로 실행한다.

### REQ-REA-08 doctor
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL report in `auto doctor` text and `--json` the REA decision and its source, the Node probe state and version, the ownership class of each selected REA target platform, and the OMP and OpenCode guidance, without starting the REA server or making any network call.

### REQ-REA-09 OMP·OpenCode 안내
Priority: Must · EARS: StateDriven

WHERE the project includes OMP or OpenCode and the REA decision is not off, THEN THE SYSTEM SHALL write no REA MCP configuration for them and print, in init, update, and doctor, one guidance line stating that `/auto reverse` analysis runs on Claude Code, Codex, and Antigravity and that `npx -y rea-agents@6.5.0 setup` registers REA for the main session only.

### REQ-REA-10 reverse-engineer 에이전트
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL ship `content/agents/reverse-engineer.md` by the Reverse Engineer Agent Contract, so that every platform generates it whatever the REA decision, its Claude Code and Antigravity tools are the REA server and read tools only, no platform's file grants Write, Edit, or Bash, it returns a Runtime Capture Request instead of calling a runtime-capture tool without recorded consent, and its body demands a behaviour specification with item IDs, evidence_id references, and unknowns, never decompiled pseudocode or assembly, ending with the platform's unavailable sentence when REA tools are missing.

### REQ-REA-11 17번째 에이전트 파급
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL add `reverse-engineer` to the canonical agent list and to every consumer in the Agent Set Contract, so that each platform generates 17 agents equal to `CanonicalAgentNames()`, OMP autopus mode routes the 17 agents and task, native OMP folds it into task, the Preset Contract holds, and, when an OMP balanced profile's catalog lacks the model pinned for the analysis rung, the run leaves only reverse-engineer unrouted so it inherits the session model, prints one named fallback line, and succeeds, while every other rung keeps the existing `omp_profile_candidate_unavailable` blocker.

### REQ-REA-12 MCP 참조 변환 예외
Priority: Must · EARS: EventDriven

WHEN a skill, rule, command, workflow, or agent body is rendered for a platform other than Claude Code, THEN THE SYSTEM SHALL keep every reference whose first `__` segment after `mcp__` is exactly `rea` verbatim and keep rewriting every other `mcp__*` reference as the base commit does.

- 지금 `pkg/content/skill_transformer_replace_mcp.go:22,50`은 남은 모든 `mcp__*`를 `Context7 MCP with WebSearch fallback`으로 바꾸고, `skill_transformer_replace.go:124`가 non-Claude 렌더마다(agent 본문 포함) 부른다. `_`가 `\w`에 들어 있어 예외는 `__`로 나눈 첫 segment를 정확히 비교한다(`mcp__reader__x`는 계속 바뀐다).

### REQ-REA-13 `reverse` 라우트
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL add the `reverse` route to the router, the route registry, and the Antigravity command surface of all five platforms with one detail that follows the Reverse Route Contract.

### REQ-REA-14 clean-room handoff
Priority: Must · EARS: EventDriven

WHEN the operator confirms a behaviour specification and its owned and test paths at the gate of `/auto reverse`, THEN THE SYSTEM SHALL have the main session write only that specification to `.autopus/reverse/<slug>/behaviour-spec.md`, dispatch tester, executor, and tester with the inputs of the Reverse Handoff Contract, and keep REA tool results, evidence bundles, and decompiled text out of every dispatch prompt and out of every project file.

### REQ-REA-15 사전 승인 없음
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL NOT add any `mcp__rea` entry to `permissions.allow` and SHALL NOT write `enableAllProjectMcpServers`, `enabledMcpjsonServers`, Codex `enabled_tools`, or a REA approval-mode key into any generated file.

- 에이전트 frontmatter의 `tools: … mcp__rea`(REQ-REA-10)는 도구 허용 목록이지 권한 사전 승인이 아니다. 지금 사전 승인 MCP는 `mcp__sequential-thinking__sequentialthinking` 하나다(`pkg/content/hooks.go:196`).

### REQ-REA-16 꺼져 있으면 bytes 그대로
Priority: Must · EARS: Ubiquitous

WHILE the REA decision is undecided or off and no managed `rea` entry exists, THE SYSTEM SHALL generate `.mcp.json`, `.codex/config.toml`, and `.gemini/settings.json` byte-identical to the output of the base commit `4d73aa91` for the same configuration.

### REQ-REA-17 옛 바이너리 줄
Priority: Must · EARS: EventDriven

WHEN a write adds `features.rea` to an autopus.yaml that did not hold it, THEN THE SYSTEM SHALL print the old-binary line of the Form Text once in that run and in no other run.

- strict loader(`pkg/config/loader_strict.go:98-125`)가 v0.50.113부터 모르는 key를 거부한다. auto 0.50.123은 `features.rea`가 든 파일에서 `auto update --yes`를 종료 코드 1로 끝냈다(probe A3).

### REQ-REA-18 locale
Priority: Should · EARS: Ubiquitous

THE SYSTEM SHALL render every new prompt, notice, hint, and doctor line in Korean when the `composeLanguage` seam returns ko and in English otherwise, while stored values stay locale-independent.

### REQ-REA-19 문서
Priority: Should · EARS: Ubiquitous

THE SYSTEM SHALL document the REA default, `--rea` and `--no-rea`, the Node range, the lazy `features.rea` write and its old-binary limit, the 17-agent presets that v0.50.125 classifies as customized, the OMP analysis-rung fallback to the session model, the POSIX-only server command on Windows, the clean-room flow, and the runtime-capture consent in init, update, and doctor help, README, and the v0.50.126 CHANGELOG, without removing sentences existing tests check.

### REQ-REA-20 test seam과 수동 smoke
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL run every new path through injectable seams (`[NEW] reaNodeExec` at the exec boundary, `composeTTYProbe`, `composeCIEnv`, `composeLanguage`, a `runInitWizard` result stub in CLI tests, and huh accessible input with an injected reader for the REA step in `internal/cli/tui` tests) with no real `node`, `npx`, or network call in tests, and SHALL record the real-terminal smoke S18 and the consented live run S19 in `evidence/manual-smoke-rea.txt` together with the verdicts of probe rows A1 and A2.

### REQ-REA-21 runtime capture 동의
Priority: Must · EARS: EventDriven

WHEN reverse-engineer returns a Runtime Capture Request, THEN THE SYSTEM SHALL have the main session ask the operator once naming the target, the tool, and what will run, re-dispatch the agent with that consent only after an explicit yes, and on no or no answer continue without any runtime-capture call and record the gap under `## Unknowns`.

- `/auto reverse … --runtime`은 dispatch prompt에 `Runtime capture wanted: yes`를 붙여 에이전트의 첫 결과가 Runtime Capture Request가 되게 한다(동의 경로를 결정적으로 일으키는 입구). `--runtime`이 없으면 에이전트는 정적 증거로 닫히지 않는 항목에서만 요청한다. 어느 쪽이든 허용 전 capture 호출은 0이다.

## Contracts

### Decision State Contract

| 상태 | 판정 | doctor 출처 |
|---|---|---|
| on | `features.rea.enabled: true` | `features.rea` |
| off | `features.rea.enabled: false` | `features.rea` |
| on | key 없음, 선택된 REA 대상 플랫폼의 파일 중 하나 이상에 managed `rea` 항목 | `managed_entry` |
| undecided | 그 밖 전부(사용자 소유·other_pin 항목만 있는 경우 포함) | `none` |

쓰기 표. M은 "이번 실행 뒤 선택된 대상 파일에 managed 항목이 하나 이상 있는가"다.

| 결정 | M | autopus.yaml |
|---|---|---|
| on | 있음 | key가 없으면 쓰지 않는다(항목이 결정을 담는다). `false`면 `true`로 고친다 |
| on | 없음(gate 실패, 대상 파일이 모두 사용자 소유, 사용자가 지운 파일) | `enabled: true`를 쓴다 |
| off | - | `enabled: false`를 쓴다. 이미 같으면 쓰지 않는다 |
| undecided | - | 쓰지 않는다 |

- M은 이번 실행의 platform 쓰기가 끝난 뒤 선택된 대상 파일을 다시 읽어 정하고 autopus.yaml은 그 뒤에 쓴다. preview는 계획된 mapping으로 정한다. manifest에 있던 파일을 사용자가 지워 `ActionSkip`(`pkg/adapter/manifest.go:137`)이 되면 `--rea`도 그 파일을 다시 만들지 않고 그 파일은 M에 들지 않는다.
- key를 지우는 경로는 없다. 재-init은 기존 key를 옮긴다(REQ-REA-02). 사용자가 지운 key는 위 판정을 다시 탄다.
- adapter 입력은 `[NEW] config.RuntimeConf.ReaMCP`(`[NEW] config.ReaMCPAction`, yaml `-`, `pkg/config/runtime.go:13-15` 옆)다. zero 값 `keep`, `add`, `remove`. on + pass → add, on + gate 실패 → keep, off → remove, undecided → keep, `auto platform remove`의 Clean → remove. CLI를 거치지 않는 생성(`DefaultFullConfig` test, eval harness)과 결정을 풀지 않는 `auto platform add`는 keep이라 넣거나 지우지 않는다(다음 update가 더한다).
- action 전달: `cloneHarnessConfig`(`internal/cli/update_preview.go:20-35`)는 YAML round trip 뒤 `Runtime.ReaMCP`를 복사한다(다른 `Runtime` 필드는 지금처럼). update는 preview 분기(`update.go:112`) 전에 flag와 기존 결정으로 action을 정하고 질문 뒤 다시 정한다. 그래서 preview, 기존 플랫폼(`update.go:226`), 새로 감지된 플랫폼(`platform_update.go:64`)이 같은 action을 받는다.
- off sweep: off는 선택되지 않은 대상 플랫폼 파일의 managed 항목도 지운다. adapter가 돌지 않는 파일은 CLI가 `[NEW] pkg/reamcp`의 같은 변환을 `adapter.SafeWorkspacePath`·`adapter.WriteFileIfChanged`(`pkg/adapter/writeutil.go:11`)로 적용한다.
- 고른 근거와 대안 비교는 `research.md` `### Q1`이다.

### MCP Entry Shape Contract

| 플랫폼 | 자리 | managed 모양(pin 6.5.0) |
|---|---|---|
| Claude Code | `.mcp.json` `mcpServers.rea` | `{"$comment": "REA reverse-engineering MCP (rea-agents@6.5.0)", "command": "/bin/bash", "args": ["-l", "-c", "exec npx -y rea-agents@6.5.0 mcp"]}` |
| Antigravity | `.gemini/settings.json` `mcpServers.rea` | `{"command": "/bin/bash", "args": ["-l", "-c", "exec npx -y rea-agents@6.5.0 mcp"]}` |
| Codex | `.codex/config.toml` 끝 | 빈 줄 하나, `[mcp_servers.rea]`, `command = "npx"`, `args = ["-y", "rea-agents@6.5.0", "mcp"]`, `startup_timeout_sec = 30` |

| 분류 | 조건 | add | remove | doctor |
|---|---|---|---|---|
| absent | `rea` 이름이 없다 | 넣는다 | 할 일 없음 | `absent` |
| managed | `[NEW] reamcp.ManagedPins`(`["6.5.0"]`, 이전에 낸 pin 전부) 중 하나의 모양과 정확히 같다. JSON은 deep-equal이고 Claude의 두 pin 자리(`$comment`, args 셋째)가 같은 pin이다. Codex는 rea 서버 table의 빈 줄을 뺀 줄이 세 줄과 순서까지 같고 하위 table이 없다 | 현재 pin 모양으로 바꾼다 | 지운다 | `managed` |
| other_pin | 모양이 같고 pin 자리만 목록 밖(`6.4.0`, `latest`). Claude는 두 자리가 같은 pin | 손대지 않는다 | 손대지 않는다 | `other_pin` |
| user_owned | 그 밖(다른 command·args, 더한 key·`env`, 다른 `$comment`, Claude 두 자리 pin 불일치, Codex 주석 줄, 빠진 key, `[mcp_servers.rea.*]` 하위 table) | 손대지 않는다 | 손대지 않는다 | `user_owned` |

- 실행 형식은 그 파일의 기존 서버와 같다(Claude·Antigravity `/bin/bash -l -c "exec npx …"`, Codex `npx`). Windows에서는 `/bin/bash`가 없어 Claude·Antigravity 서버가 시작되지 않는다. 기존 서버와 같은 한계이고 REQ-REA-19 문서가 알린다.
- 보존: JSON(Claude·Antigravity)은 기존 merge가 파일을 decode한 뒤 `json.MarshalIndent`로 다시 쓴다(`claude_files.go:25-87`, `antigravity_settings.go:18-60`). 그래서 사용자 항목은 값(deep-equal)으로 보존되고 서식은 기준 commit의 merge가 이미 모든 update에서 정규화하는 대로다. 이 SPEC은 그 정규화 밖의 byte를 바꾸지 않는다. Codex는 줄 단위 merge라 rea 서버 table 밖의 줄과 user_owned table은 bytes가 같다.
- Claude: managed 삭제 loop(`claude_files.go:60-68`)는 `rea`를 건너뛰고 rea는 action으로만 바뀐다. clean(`claude_clean.go:176-207`)은 remove와 같은 함수를 부른다.
- Codex: `codexManagedConfigKeys`(key 단위, `codex_config_structural.go:10-27`)에 넣지 않는다. 넣으면 사용자 `[mcp_servers.rea]`의 `command`·`args`를 덮는다. rea 서버 table은 `[mcp_servers.rea]`와 `[mcp_servers.rea.*]` 전부이고 줄 읽기는 기존 `codexTOMLScanState`를 써서 multiline 문자열 안의 `[`를 header로 보지 않는다. remove는 managed table과 그 앞 빈 줄 하나를 지운다. clean(`codex_clean_merge.go:180-213`)은 generic clean 전에 같은 제거를 한다.
- Antigravity: `mergeSettingsMaps` 뒤에 rea를 적용해 key를 지울 수 있다. clean은 현재 렌더와 무관하게 managed 모양이면 지운다.
- 쓰기 경계: REA가 내용을 바꾸는 쓰기는 `adapter.SafeWorkspacePath`(`pkg/adapter/transaction_helpers.go:278`, symlink를 건너는 경로를 거부)로 확인한다. 거부되면 그 플랫폼 쓰기는 경로를 담은 오류로 끝나고 밖 파일은 그대로다. Antigravity Generate(`antigravity.go:130-137`)와 clean(`antigravity_lifecycle_clean.go:69-110`)은 지금 `os.WriteFile`을 바로 쓰므로 REA가 내용을 바꾸는 실행에서만 이 확인을 더한다(keep 실행의 기존 동작은 그대로).
- `tool_timeout_sec`는 쓰지 않는다(Codex 기본 60초). probe A2가 필요를 보이면 scope expansion 규칙으로 더한다.

### Node Gate Contract

| exec 결과 | 판정 |
|---|---|
| `/bin/bash`(Windows는 `node`) 시작 실패, 또는 종료 127이고 stdout이 비어 있음 | absent |
| 종료 0, 마지막 비지 않은 줄(앞뒤 공백 제거)이 `v<MAJOR>.<MINOR>.<PATCH>`이고 22.x이면서 22.19.0 이상, 24.x이면서 24.11.0 이상, 또는 MAJOR 26 이상 | pass |
| 종료 0, 같은 형식이고 위 범위 밖 | unsupported |
| 종료 0이지만 형식 밖(prerelease 꼬리 포함), 그 밖의 0이 아닌 종료(출력이 있어도), 5초 timeout | unknown |

- probe는 POSIX에서 서버와 같은 `/bin/bash -l -c 'node --version'`, Windows에서 PATH의 `node --version`이다. 실행당 1회, 결정이 on이거나 질문에 Yes가 나왔을 때만 돌고 doctor는 늘 1회다. seam은 exec 경계 `[NEW] reaNodeExec`(stdout, 종료 상태, 시작 오류)이고 판정은 그 결과로만 한다.
- gate는 추가만 막는다. managed 항목을 지우는 것은 결정 off와 `auto platform remove`뿐이라 Node가 다른 기계에서 update해도 항목이 오락가락하지 않는다.

### Update Prompt Contract

| 결정 \ 실행 | 대화형(TTY, `--yes` 없음, truthy `CI` 없음, `--auto`·preview 아님) | non-TTY, `--yes`, truthy `CI`, `--auto`, workspace의 repo별 실행 | preview(`--plan`·`--preview`·`--dry-run`) |
|---|---|---|---|
| undecided | REA 질문 1회, 기본 Yes, 답을 이번 실행에 적용 | 질문 없음, REA stdin 읽기 0, rea 항목·`features.rea` 쓰기 없음, hint 1줄 | 쓰기 없음, hint 1줄 |
| on, off | 질문 없음, 결정 적용 | 같음 | 쓰기 없이 미리보기 |
| `--rea`, `--no-rea` | 질문 없음, flag 적용 | 같음 | 쓰기 없이 미리보기(계획된 rea 변경 포함) |
| 두 flag 함께, 또는 flag + `--workspace` | 쓰기 전 오류 | 같음 | 같음 |

- REA 대상 플랫폼이 없으면 묻지 않고 hint도 없다(OMP·OpenCode 안내만).
- 질문 자리는 `promptLanguageSettings` 뒤, compose 단계 앞이다. REA 질문은 SPEC-MODELUX-003 compose 단계의 질문과 별개이고 그 단계는 지금처럼 돈다. 17번째 에이전트와 `reverse` 라우트는 결정과 무관하게 모든 쓰는 실행에서 생성된다.

### Doctor Contract

| check ID | detail 예상 값 | severity |
|---|---|---|
| `doctor.rea.decision` | `on (features.rea)`, `on (managed_entry)`, `off (features.rea)`, `undecided` | info |
| `doctor.rea.node` | `pass v24.11.0`, `unsupported v22.18.0`, `absent`, `unknown` | 결정 on이고 pass가 아니면 warning, 아니면 info |
| `doctor.rea.platform.<claude-code·codex·antigravity-cli>` | `managed`, `absent`, `other_pin`, `user_owned` | 결정 on이고 gate pass인데 `absent`면 warning, 아니면 info |
| `doctor.rea.guidance.<omp·opencode>` | `npx -y rea-agents@6.5.0 setup` | info |

- payload `data.rea`는 `{"decision", "decision_source", "node": {"status", "version"}, "platforms": {<platform>: <class>}, "guidance": [<platform>]}`이다. 선택되지 않은 플랫폼은 key가 없고 결정이 off면 guidance는 빈 배열이다. 조립 자리는 `internal/cli/doctor_json.go:118-137`과 `doctor_text.go:15`다.

### Reverse Engineer Agent Contract

| 자리 | 값 |
|---|---|
| source frontmatter | `name: reverse-engineer`, `model: opus`, `effort: high`, `tools: Read, Grep, Glob, mcp__rea`, `permissionMode: default`, `maxTurns: 120`, skills 없음 |
| Claude `.claude/agents/autopus/reverse-engineer.md` | `tools: Read, Grep, Glob, mcp__rea`, `permissionMode: default`. balanced 표준 배치면 `model: claude-opus-5-5`, `effort: high`, ultra면 `model: opus`, `effort: max` |
| Antigravity `.gemini/agents/autopus/reverse-engineer.md` | tools `read_file`, `grep_search`, `glob`, `mcp_rea_*` 순서 |
| Codex `.codex/agents/reverse-engineer.toml` | balanced `model = "gpt-6.1-sol"`, `model_reasoning_effort = "xhigh"`. Autopus Codex 에이전트 형식에는 도구 칸이 없어(`pkg/content/agent_transformer.go:113`) 세션 도구를 상속한다. clean-room은 developer_instructions, 메인 세션 gate, handoff 입력 제한이 맡는다(한계, `research.md` Open Risks R4) |
| OMP `.omp/agents/reverse-engineer.md` | `tools` = `glob`, `grep`, `read`(정렬), `mcp_`로 시작하는 값 없음. override가 없으면 `thinkingLevel: high` |
| OpenCode `.opencode/agents/reverse-engineer.md` | `"*": deny` 아래 allow는 `glob`, `grep`, `question`, `read`, `skill`, `task`뿐이고 `mcp__` key가 없다 |
| 본문 | (1) 입력은 대상 경로, 기능 이름, 있으면 `Runtime capture wanted: yes`와 메인 세션이 기록한 동의 줄이다. `Runtime capture wanted: yes`가 있고 동의 줄이 없으면 첫 결과는 Runtime Capture Request다. (2) 대상을 실행하거나 붙거나 계측하는 REA 도구는 그 도구와 대상에 맞는 동의 줄이 없으면 부르지 않고 `## Runtime Capture Request`(target, tool, will run, why)를 돌려주고 멈춘다. (3) 결과는 `## Behaviour`(`B-1`…), `## Inputs and Outputs`, `## Evidence`(evidence_id 목록), `## Unknowns`, `## Acceptance Items`(`A-1`…) 다섯 절의 행동 명세뿐이다. (4) 디컴파일된 의사코드, assembly, 대상 소스 발췌를 넣지 않는다. (5) REA 도구가 없으면 Form Text의 플랫폼별 미등록 문장으로 끝낸다. (6) 사용자에게 직접 묻지 않는다 |

- Bash를 빼는 이유: REA는 CLI(`npx rea-agents …`)로도 같은 분석과 runtime capture를 돌릴 수 있어 Bash가 있으면 MCP 경로와 동의 요청을 우회하고 파일을 쓸 수 있다. 대가로 `file`·`strings` 같은 빠른 확인은 REA 도구가 대신한다.
- 실행 모드: subagent는 fork·background로 돌 수 있고 Claude가 foreground를 고를 수 없을 수 있다(review round 1의 문서 근거). 이 계약은 실행 모드에 기대지 않는다. 에이전트는 사용자와 대화하지 않고 동의·확인은 메인 세션이 받는다. background에서도 MCP 도구와 Read·Grep·Glob은 남는다.
- `permissionMode: default`는 메인 세션이 `acceptEdits`·`auto`·`bypassPermissions`면 덮일 수 있다. 그래서 도구 권한 질문을 runtime capture 동의로 보지 않는다(REQ-REA-21).
- Claude Code 문서(2026-10-15): `tools`는 `mcp__<server>`와 `mcp__<server>__*`를 받고, 해석되는 항목이 하나라도 있으면 실행된다. REA가 등록되지 않아도 Read·Grep·Glob이 해석되어 에이전트가 미등록 문장을 낸다.

### Reverse Route Contract

| 단계 | 규칙 |
|---|---|
| 진입 | `/auto reverse <대상> <기능> [--runtime]`(Codex `$auto reverse`, 다른 플랫폼은 그 문법). 인자가 없으면 메인 세션이 묻는다. `--runtime`이면 dispatch prompt에 `Runtime capture wanted: yes`를 붙인다 |
| 1 분석 | 메인 세션이 reverse-engineer를 보낸다. 실행 모드는 정하지 않는다. REA가 없으면 그 에이전트의 미등록 문장으로 끝내고 아무것도 쓰지 않으며 다른 에이전트를 보내지 않는다 |
| 2 동의 | 결과가 `## Runtime Capture Request`면 메인 세션이 Form Text의 동의 질문(허용 / 거절)을 한 번 묻는다. 허용이면 동의 줄 `Runtime capture consent: <tool> on <target>`을 붙여 1을 다시 보내고, 거절이나 무응답이면 동의 줄 없이 정적 분석만 이어 가고 빈 곳을 `## Unknowns`에 남긴다. `--auto`도 묻는다 |
| 3 확인 gate | 메인 세션이 명세와 제안한 소유 경로·테스트 경로를 보이고 `확인 / 수정 / 중단`을 묻는다. 수정은 1로 돌아가고 중단은 아무것도 쓰지 않고 어떤 에이전트도 보내지 않는다. `--auto`도 멈춘다 |
| 4 저장 | 확인 뒤 메인 세션이 명세만 `.autopus/reverse/<slug>/behaviour-spec.md`에 쓴다. slug: 기능 이름을 소문자로 바꾸고 `[a-z0-9]` 밖의 문자 묶음을 하이픈 하나로 바꾼 뒤 양끝 하이픈을 지운다. 비면 `feature`. 그 디렉터리가 있으면 `-2`, `-3`을 붙인다(`Login Flow!` → `login-flow`, `로그인 화면` → `feature`, 두 번째 `Login Flow!` → `login-flow-2`). `.autopus/reverse/`는 생성 `.gitignore`가 무시하는 로컬 artifact다 |
| 5 구현 | Reverse Handoff Contract대로 tester(Test Scaffold) → executor → tester(Testing)를 보낸다 |
| 금지 | dispatch prompt에 REA 결과, evidence bundle, 디컴파일 텍스트를 넣지 않는다. `export_evidence_bundle`을 프로젝트 안 경로로 부르지 않는다. 디컴파일 텍스트를 어떤 파일에도 쓰지 않는다 |

- 계약 문장은 `[NEW] templates/shared/auto-reverse-contract.md.tmpl` 한 곳에 두고 다섯 detail이 렌더한다: Claude `.claude/skills/auto-reverse/SKILL.md`(`claude_workflow_skills.go:22-42` customBody, `[NEW] claude_reverse_route.go`), Antigravity `.gemini/skills/autopus/auto-reverse/SKILL.md`(`[NEW] templates/gemini/skills/auto-reverse/SKILL.md.tmpl`)와 command `.gemini/commands/auto/reverse.toml`(`[NEW] templates/gemini/commands/auto/reverse.toml.tmpl`, 기존 19개 옆), Codex `codex_workflow_specs.go`·`codex_workflow_custom.go`, OMP `omp_specs.go`·`omp_workflow_render.go`, OpenCode `opencode_specs.go`·`opencode_workflow_custom.go`.
- router: Claude `templates/claude/commands/auto-router.md.tmpl:55-75`과 Antigravity root router에 `reverse` 1행(각자 detail 경로), Codex router 설명(`codex_workflow_specs.go:17`) 끝에 `/reverse`. root router는 8192 bytes 이하다(지금 템플릿 6174 bytes).

### Reverse Handoff Contract

| 받는 쪽 | 기존 필수 입력 | reverse 흐름의 값 |
|---|---|---|
| tester (Test Scaffold) | SPEC/Task 식별자, 대상 요구사항(acceptance ID), 테스트 경로, Phase(`content/agents/tester.md:75-77`) | `SPEC: N/A`, `Task: reverse/<slug>`, `Behaviour Spec: <파일 경로>`, `A-1`…, gate에서 확인한 테스트 경로, `Phase: Test Scaffold` |
| executor | SPEC/Task 식별자(없으면 N/A), 관련 요구사항, 수정 대상 파일, 소유·금지 범위, 빠지면 BLOCKED(`content/agents/executor.md:104-106`), Step 0 SPEC 로딩(`:40-45`) | `SPEC: N/A`, `Task: reverse/<slug>`, `Behaviour Spec: <파일 경로>`, `B-1`…와 `A-1`…, gate에서 확인한 소유 경로, 금지 범위 = 분석 대상 경로·`.autopus/reverse/`(읽기만)·REA 결과 |
| tester (Testing) | 위 tester 입력 | 같은 값, `Phase: Testing` |

- executor는 GREEN·REFACTOR만 맡으므로(`content/agents/executor.md:31-38`) tester의 Test Scaffold가 먼저 간다. 사용자 결정(분석 → 확인 → executor → tester)의 순서를 기존 executor 계약에 맞춘 것이다(`spec.md` Open Issues).
- `content/agents/executor.md` Step 0·입력 절과 `content/agents/tester.md` 입력 절에 "`SPEC: N/A`이고 `Behaviour Spec:` 경로가 있으면 그 파일이 요구사항 원천이다"를 한 문장씩 더한다.

### Agent Set Contract

| 소비처 | 바뀌는 점 |
|---|---|
| `pkg/config/quality_tier.go:17-21` `canonicalAgentNames` | `planner` 뒤, `reviewer` 앞에 `reverse-engineer`(알파벳 순서, index 10). `agentNameAliases`(L25-32)에 `reverse_engineer` |
| `pkg/config/role_model_policy_matrix.go:38-55` `capabilityBySourceAgent` | `reverse-engineer: deep_reasoning`. OMP role은 `autopus_reverse_engineer` |
| `role_model_policy_balanced.go:20-68`, `native_balanced.go:33-38` | Preset Contract의 새 rung |
| `defaults.go:136-147` ultra | `reverse-engineer: opus` |
| `quality_preset_history.go:52-104` | HB6, HU3 |
| `omp_native_agents.go:11-25` `OMPNativeAgentForRole` | native 모드에서 `task`로 접힌다 |
| `omp_agent_source.go:50-52` `OMPAutopusAgentNames` | 정규 목록에서 나오므로 17 + task = 18. 주석의 16을 17로 |
| `internal/cli/quality_compose_groups.go:24-37` | 탐색·보조 = explorer, annotator, reverse-engineer. 그룹 선택은 지금처럼 세 멤버 모두에 적용된다 |
| `templates/shared/omp-agent-pipeline.md.tmpl:69-72`, `internal/cli/platform_omp_agent_migration.go:27,40`, `init_help.go:20`, `update_help.go:31`, `quality_help.go:30`, `content/skills/adaptive-quality.md:87` | 숫자 16 → 17, 목록에 reverse-engineer, pipeline 템플릿에 선택 문장 1줄(분석만, 구현 금지) |
| `content/agents/executor.md`, `content/agents/tester.md` | Reverse Handoff Contract의 한 문장씩 |
| 생성 에이전트(Codex TOML, Gemini MD), OMP agent 파일·override key | `make generate-templates`와 정규 목록에서 파생 |

### Preset Contract

| 항목 | 값 |
|---|---|
| balanced 표준 배치 | `[NEW] balancedRungAnalysis`, tier `opus`. anthropic `anthropic/claude-opus-5-5`·thinking `high`, openai `openai-codex/gpt-6.1-sol`·thinking `xhigh`. 손으로 opus를 고른 사다리 값과 같다(`ClaudeAgentEffort("balanced","opus")` = `high`, `codexLadderProfile` opus = `gpt-6.1-sol`·`xhigh`) |
| `defaultBalancedAgentTiers()`, ultra | 둘 다 17개, reverse-engineer `opus` |
| Historical | `HB6` balanced 16개(7 fable + 9 sonnet, commit fbe502c0, v0.50.116–v0.50.125, runtimeStandard), `HU3` ultra 16개(7 fable + 9 opus, commit 001830fd, v0.50.115–v0.50.125) |
| 분류 | v0.50.125 기본값 16개 map은 Historical이고 기존 `u` 갱신이 reverse-engineer `opus`를 더한다. 그 전에도 reverse-engineer는 map에 없으므로 표준 배치를 받는다(`native_balanced.go:20-24`) |
| OMP balanced | 기존 rung은 지금처럼 후보 하나를 pin하고 없는 model은 `omp_profile_candidate_unavailable`(`internal/cli/platform_omp_profile_preview.go:75`)로 막는다(SPEC-MODELUX-003 그대로). 새 analysis rung만 다르다(Gate 1 사용자 결정 "세션 모델로 대체"): 그 후보(`anthropic/claude-opus-5-5` 또는 `openai-codex/gpt-6.1-sol`)가 catalog에 없으면 reverse-engineer의 프로필 route만 버린다. `task.agentModelOverrides`에 reverse-engineer key가 없어 세션 model을 상속하고, `.omp/agents/reverse-engineer.md`는 지금처럼 `model:`이 없고 `thinkingLevel: high`를 갖는다. preview·apply·update 출력은 Form Text의 대체 줄을 1번 내고 실행은 성공한다. operator pin(`role_model_policy.agents.reverse-engineer`)의 model이 없으면 지금처럼 fail closed다. 판정은 `[NEW] config.OMPReaAnalysisFallback`(rung·agent·catalog) 한 곳이고 CLI plan과 adapter가 함께 쓴다 |
| 탐색·보조 표시 | 기본값에서 `현재: 혼합 (opus 1 · sonnet 2)` |

### Form Text (ko / en)

- wizard: `[n/m] REA 역공학 MCP` · 설명 `Claude Code·Codex·Antigravity에 rea-agents@6.5.0을 프로젝트 범위로 등록합니다 (Node ^22.19.0 || ^24.11.0 || >=26.0.0)` / `[n/m] REA reverse-engineering MCP` · `Registers rea-agents@6.5.0 for Claude Code, Codex, and Antigravity in this project (Node ^22.19.0 || ^24.11.0 || >=26.0.0)`
- update 질문: `REA 역공학 MCP(rea-agents@6.5.0)를 이 프로젝트의 Claude Code·Codex·Antigravity에 등록할까요? [Y/n]` / `Register the REA reverse-engineering MCP (rea-agents@6.5.0) for Claude Code, Codex, and Antigravity in this project? [Y/n]`
- 등록: `REA 역공학 MCP를 등록했습니다: <파일 목록> (rea-agents@6.5.0)` / `Registered the REA reverse-engineering MCP: <files> (rea-agents@6.5.0)`
- hint: `REA 역공학 MCP를 아직 정하지 않았습니다. 켜려면 auto update --rea, 끄려면 auto update --no-rea` / `REA reverse-engineering MCP is undecided; run auto update --rea to enable or auto update --no-rea to decline`
- Node: `REA MCP를 등록하지 않았습니다: node <version>은 ^22.19.0 || ^24.11.0 || >=26.0.0 범위가 아닙니다` · `REA MCP를 등록하지 않았습니다: node를 찾지 못했습니다 (^22.19.0 || ^24.11.0 || >=26.0.0 필요)` · `REA MCP를 등록하지 않았습니다: node 버전을 확인하지 못했습니다` / `REA MCP not registered: node <version> is outside ^22.19.0 || ^24.11.0 || >=26.0.0` · `REA MCP not registered: node not found (needs ^22.19.0 || ^24.11.0 || >=26.0.0)` · `REA MCP not registered: node version unreadable`
- 옛 바이너리: `features.rea는 v0.50.126보다 오래된 auto가 읽지 못합니다` / `auto older than v0.50.126 cannot read features.rea`
- OMP 대체: `reverse-engineer: <selector>이 OMP catalog에 없어 세션 모델을 씁니다` / `reverse-engineer: <selector> not in catalog → session model`
- OMP·OpenCode: `OMP·OpenCode에는 Autopus가 REA MCP를 등록하지 않습니다. /auto reverse 분석은 Claude Code·Codex·Antigravity에서 실행하세요. npx -y rea-agents@6.5.0 setup은 REA를 사용자 범위의 메인 세션에만 등록합니다` / `Autopus does not register REA MCP for OMP or OpenCode; run /auto reverse analysis in Claude Code, Codex, or Antigravity. npx -y rea-agents@6.5.0 setup registers REA for the main session only (user scope)`
- flag: `--rea와 --no-rea는 함께 쓸 수 없습니다` · `--rea·--no-rea는 --workspace와 함께 쓸 수 없습니다. repo마다 --local로 실행하세요` / `--rea and --no-rea cannot be used together` · `--rea and --no-rea cannot be combined with --workspace; run them per repository with --local`
- doctor text: `REA: 켜짐 (managed 항목)` · `REA: 켜짐 (autopus.yaml)` · `REA: 꺼짐 (autopus.yaml)` · `REA: 미결정` / `REA: on (managed entry)` · `REA: on (autopus.yaml)` · `REA: off (autopus.yaml)` · `REA: undecided`
- 동의 질문(메인 세션): `REA가 <target>을 실행하거나 붙는 <tool>을 쓰려 합니다: <will run>. 허용할까요? (허용 / 거절)` / `REA wants to use <tool>, which runs or attaches to <target>: <will run>. Allow? (allow / decline)`
- 확인 gate(메인 세션): `이 행동 명세로 구현할까요? 소유 경로: <paths> · 테스트 경로: <path> (확인 / 수정 / 중단)` / `Implement from this behaviour specification? owned: <paths> · tests: <path> (confirm / revise / stop)`
- 에이전트 미등록 문장(본문은 한국어): Claude Code·Codex·Antigravity `REA MCP 도구를 찾지 못했습니다. auto update --rea로 이 프로젝트에 등록하거나 npx -y rea-agents@6.5.0 setup으로 사용자 범위에 등록하세요.` · OMP·OpenCode `이 플랫폼의 reverse-engineer에는 REA 도구가 없습니다. /auto reverse는 Claude Code·Codex·Antigravity에서 실행하세요.`

## Existing Test Changes

0건이 아니다. T2가 파일마다 분류하고(`[NEW] evidence/test-classification.md`) 규칙은 아래뿐이다.

1. 수 고정: 정규 에이전트 16 → 17, OMP 16 + task 17 → 18(`pkg/adapter/integration_test.go:214`, `latest_cli_omp_contract_test.go:33`, `codex/codex_agents_test.go:126`, `omp/omp_agents_test.go:83,118`, `omp_acceptance_surface_test.go:52,169`, `omp_latest_contract_test.go:30`, `omp_native_registry_test.go:32`, `omp_agent_surface_report_test.go:52`, `pkg/config/omp_agent_source_test.go:67`, `role_model_policy_pins_test.go:125`, `internal/cli/` compose·migration test 등). literal만 바꾼다.
2. frozen route 목록(`pkg/adapter/claude/router_budget_test.go:16-19`, `pkg/adapter/antigravity/router_budget_test.go`의 route·command 목록, `codex_surface_test.go`, `evals/harness/tasks/surface/GT-ROUTE-*` 중 `GT-ROUTE-ANTIGRAVITY-COMMANDS` 포함)에 `reverse`를 더한다.
3. compose golden(`internal/cli/testdata/quality_compose/*.golden`): reverse-engineer 행 추가, 탐색·보조 멤버 목록과 `현재:` 표시만 바뀐다. 다른 diff는 회귀다.
4. v0.50.125 기본값 16개를 literal로 쓴 fixture는 HB6·HU3 Historical로 분류된다. 기존 갱신 줄·`u` 선택지가 나타나는 자리만 바뀐다.
5. MCP inventory test(`GT-SETTINGS-MCP-SERVERS.json`, `antigravity_settings_test.go`, `antigravity_clean_ownership_test.go`, `claude_lifecycle_contract_test.go`, `codex_config_test.go`, `codex_latest_v2_test.go`)는 바꾸지 않는다(REQ-REA-16).
6. CLI test 바이너리의 `TestMain`(`internal/cli/pipeline_backend_omp_rpc_test.go:36`)이 `reaNodeExec`를 `v24.11.0`·종료 0으로 고정한다. 새 init 출력이나 `.mcp.json` bytes를 통째로 비교하는 test는 등록 줄과 managed 항목만 더해진다.
7. OMP catalog fixture `ompCLIBalancedCatalogJSON`(`internal/cli/platform_omp_profile_flags_test.go:74`), `ompCLIProfileNativeCatalogJSON`(`:87`), `composeOMPCatalogML`(`quality_compose_omp_autopus_test.go:29`)에 `anthropic/claude-opus-5-5`와 `openai-codex/gpt-6.1-sol` 행을 더해 기존 OMP test가 reverse-engineer key를 갖게 한다. 그 목록을 단언하는 자리(`quality_compose_omp_autopus_test.go:113-115`)는 두 model만 더해진다. 기존 rung model을 뺀 blocker test는 바뀌지 않는다.
8. 대화형 update·init 통합 test에서 REA 질문이 새로 열리는 실행은 입력에 Enter 하나를 더한다(SPEC-MODELUX-003 규칙 3과 같은 방식). 단언은 바꾸지 않는다.
9. 생성 `.gitignore` 정책 test(`internal/cli/init_gitignore_policy_test.go`의 required 목록)와 gitignore 갱신 test에는 `.autopus/reverse/` 한 줄만 더해진다.
10. `internal/cli/quality_compose_line_golden_test.go:36`(`update R enter`): golden은 SPEC-MODELUX-001 줄 출력과 bytes가 같아야 하므로 바꾸지 않는다. 대신 settled R fixture의 REA를 실제 비대화형 `auto update --no-rea`로 off로 미리 정한다(`reaSettledOff`). off는 REA 질문·hint·OMP·OpenCode 안내·옛 바이너리 줄이 모두 없는 유일한 결정이라 출력이 compose 단계의 것만 남는다(on은 안내 줄을 낸다).
11. `internal/cli/json_contract_commands_test.go:56`(`TestJSONContract_PhaseOneCommandsRequireEnvelopeSupport`의 doctor 행): 따로 빌드한 `auto` binary에는 `TestMain`의 seam이 닿지 않아 doctor의 Node probe(`/bin/bash -l`)가 실제로 돈다. 결과가 개발자 기계의 login profile에 기대지 않도록 그 행만 자식 HOME을 `t.TempDir()`로 둔다(REQ-REA-20 결정성). envelope 단언은 그대로이고 Node 상태는 단언하지 않는다.
12. `internal/cli/edit_guard_corpus_test.go:148`(합법 corpus fixture): `README.md`를 실제 프로젝트 파일로 심는다. 없는 상대 경로의 Gemini CLI `replace`는 host가 그 경로로 끝나는 workspace 파일로 바꿔 쓰므로(guard의 suffix search) 없는 `README.md`의 replace는 보호된 `.claude/hooks/README.md`를 고치는 편집이고 합법이 아니다. 기본 실행은 target마다 shape 하나를 index로 돌리는데, 생성 표면이 늘어 `README.md`가 index 1415(`1415 % 12 = 11`, Gemini replace)에 닿았다. `AUTOPUS_EDITGUARD_CORPUS_FULL=1`은 이 SPEC 전 corpus에서도 같은 deny로 실패한다. guard 판정, corpus 범주, target 수는 그대로이고 없는 파일의 search deny는 `pkg/editguard`의 `TestGeminiReplace_RelativePathNamingNoFile_DeniesTheProtectedFileItReaches`가 고정한다.
13. Codex analysis rung 대체(T23 scope expansion, lead 결정 (a), commit `7b444d1b`): Codex catalog에 `gpt-6.1-sol`·`xhigh`가 없는 compose golden(`table-D-CAT-B.golden`)에서는 reverse-engineer의 Codex 칸이 `세션 모델`이고 footer에 대체 줄 1개가 더해진다.

- 규칙 1은 정규 에이전트 집합을 literal로 적은 map·목록(기본 등급, 표준 배치, role 행, test 이름의 수)에 reverse-engineer를 그 기본값으로 더하는 것과, 17번째 에이전트로 하나씩 느는 파생 수(사용자 수정 `N개 다름`, Codex `대체 N칸`, compose `*` 칸 수, OMP 표면 줄 수, catalog를 모를 때 보고하는 profile 수)를 포함한다.
- 규칙 2는 route 수(20 → 21)를 고정한 OMP test와 reverse를 더한 GT-ROUTE task의 digest를 담은 `evals/harness/baseline.json`을 포함한다.
- 규칙 3은 golden 밖 compose test의 같은 표시와, 새 행 때문에 넓어진 열 폭(`table-O1-CAT-A.golden`, 값 변화 없음)을 포함한다. 탐색·보조 그룹에 opus 멤버가 생겨 effort 선택지가 두 모델을 함께 보이고 그 선택이 저장하는 efforts key가 바뀌는 것도 이 규칙이다.
- 규칙 7은 목록에 두 model이 더해져 같은 model을 고르는 입력 번호가 하나씩 밀리는 것을 포함한다.
- 새 기능 test가 쓰는 helper 추가(`internal/cli/tui/export_test.go`의 alias 2개)는 기존 단언을 바꾸지 않는다. `pkg/worker/a2a/control_plane_test.go`는 같은 branch에 들어온 이 SPEC 밖의 commit `33b5bb55`(A2A `redline_instructions.v1` capability)다. 파일별 분류는 `evidence/test-classification.md`다.

## 생성 파일 상세

- `pkg/config`: `schema.go`(+1줄), `runtime.go`(+1줄), `[NEW] rea_feature.go`(+test), `quality_tier.go`, `role_model_policy_balanced.go`, `native_balanced.go`, `defaults.go`, `quality_preset_history.go`, `role_model_policy_matrix.go`, `omp_native_agents.go`, `omp_agent_source.go`(주석).
- `[NEW] pkg/reamcp`: `shape.go`(+test, 모양·pin·분류·JSON과 Codex 변환), `node.go`(+test, exec 결과 판정). Go 의존성은 늘지 않는다.
- adapter: `[NEW] pkg/adapter/claude/claude_rea_mcp.go`(+test), `claude_files.go`, `claude_clean.go`, `[NEW] claude_reverse_route.go`, `claude_workflow_skills.go`; `[NEW] pkg/adapter/codex/codex_rea_mcp.go`(+test), `codex_settings.go`, `codex_clean_merge.go`, `codex_workflow_specs.go`, `codex_workflow_custom.go`; `[NEW] pkg/adapter/antigravity/antigravity_rea_mcp.go`(+test), `antigravity.go`, `antigravity_settings.go`, `antigravity_lifecycle_clean.go`; `pkg/adapter/omp/omp_specs.go`, `omp_workflow_render.go`; `pkg/adapter/opencode/opencode_specs.go`, `opencode_workflow_custom.go`.
- content·templates: `[NEW] content/agents/reverse-engineer.md`, `content/agents/executor.md`, `content/agents/tester.md`, `content/skills/adaptive-quality.md`, `pkg/content/agent_transformer.go`, `agent_transformer_omp_render.go`, `agent_transformer_opencode.go`, `skill_transformer_replace_mcp.go`, `[NEW] templates/shared/auto-reverse-contract.md.tmpl`, `[NEW] templates/gemini/skills/auto-reverse/SKILL.md.tmpl`, `[NEW] templates/gemini/commands/auto/reverse.toml.tmpl`, `templates/shared/omp-agent-pipeline.md.tmpl`, `templates/claude/commands/auto-router.md.tmpl`, Antigravity root router 템플릿.
- OMP 대체: `[NEW] pkg/config/omp_rea_fallback.go`(+test, `OMPReaAnalysisFallback`), `[NEW] internal/cli/platform_omp_rea_fallback.go`(+test, `platform_omp_profile_preview.go`(276줄)의 blocker loop L94-115에서 호출), `[NEW] pkg/adapter/omp/omp_model_rea_fallback.go`(+test, route compile·활성화 경로 `omp_model_routing_compile.go`·`omp_model_integration_degraded.go:21-37`에서 호출).
- `internal/cli`: `[NEW] rea_decision.go`, `[NEW] rea_node_probe.go`, `[NEW] init_rea.go`, `[NEW] update_rea.go`, `[NEW] doctor_rea.go`, `[NEW] tui/wizard_rea_step.go`(모두 +test), `init.go`, `init_helpers.go`(`gitignorePatterns`에 `.autopus/reverse/` 1줄), `update.go`(292줄 → 300줄 이하), `update_preview.go`(clone이 `Runtime.ReaMCP`를 옮김), `tui/wizard_steps.go`, `doctor_json.go`, `doctor_text.go`, `quality_compose_groups.go`, `quality_help.go`, `init_help.go`, `update_help.go`, `platform_omp_agent_migration.go`, `quality_compose_form_edit.go`(주석).
- 문서: README, CHANGELOG(v0.50.126), help 상수. 증거: `evidence/probe-rea-forward-compat.md`, `evidence/review-rounds.md`(있음), `[NEW] evidence/test-classification.md`, `[NEW] evidence/manual-smoke-rea.txt`. 신규·변경 소스는 300줄 이하이고 `update.go`·`update_workspace.go`(300)·`doctor_json_checks.go`(293)·`wizard_steps.go`(268)에는 호출 몇 줄만 더하거나 손대지 않는다.

## Related SPECs

- SPEC-MODELUX-003(같은 릴리스, 미발행): 16을 17로만 넓힌다. OMP Agent File Contract는 17개 파일, Agent Set Contract key는 17 + task, Migration Contract 미리보기는 17줄과 `OMP는 이제 Autopus 에이전트 17개를 씁니다.`, Role Group Contract 탐색·보조는 3명, OMP Autopus Tab Contract 목록은 17개다. 옛 바이너리 경고 방식과 compose 단계는 그대로 쓴다.
- SPEC-MODELUX-001: Historical Preset Registry에 HB6·HU3를 더한다. 분류 규칙(정확한 map 일치)은 그대로다.
- SPEC-MODELUX-002: Role Group Contract의 탐색·보조 멤버만 바뀐다.
- Sibling: 없다(`research.md` `## Sibling SPEC Decision`).

## Sibling SPEC Decision

없다. Primary SPEC 하나가 Outcome Lock을 닫는다. 근거와 후보 검토는 `research.md` `## Sibling SPEC Decision`이다.

## Completion Debt

- CD-1(S18, probe A2의 시작 시간)과 CD-2(S19, probe A1, Antigravity 도구 해석, A2의 도구 호출 시간, 다섯 route 경로)가 PASS로 기록되기 전에는 완료가 아니다. 표는 `research.md` `## Completion Debt`다.

## Evolution Ideas

- 완료를 막지 않는 선택 개선만 `research.md` `## Evolution Ideas`에 둔다. 번호나 후속 문서로 승격하지 않는다.

## Open Issues

- Gate 1(2026-10-15, 사용자 결정): Q1·J-03·J-14·`--runtime`·J-07은 그대로 유지, J-01은 fail closed 대신 "세션 모델로 대체"로 바꿨다(Preset Contract OMP balanced 행). 아래 항목은 모두 resolved다.
- Q1 | resolved(Gate 1 유지) | `features.rea` lazy 쓰기 | attempt 1 | ledger의 "features flag로 저장"을 "managed 항목이 결정을 담지 못할 때만 key를 쓴다"로 좁혔다(probe A3). "늘 쓴다"를 고르면 Decision State Contract 쓰기 표의 on 두 행을 `enabled: true`로 바꾸고 REQ-REA-17이 모든 새 init에서 한 줄을 낸다.
- J-03 | resolved(Gate 1 유지) | 사용자 항목 보존 범위 | attempt 2 | PRD Story 2의 "그 항목의 bytes가 그대로다"를 JSON은 값 보존 + 기존 정규화, Codex는 bytes 보존으로 좁혔다. bytes 보존을 고르면 JSON 원문 편집기를 새로 두는 task가 하나 더 생긴다.
- J-01 | resolved(Gate 1 변경) | OMP analysis rung | attempt 3 | 사용자가 "세션 모델로 대체"를 골랐다. 새 analysis rung의 model이 catalog에 없으면 reverse-engineer만 세션 model을 상속하고 대체 줄 1개를 내며 update는 성공한다. 기존 rung의 blocker는 그대로다(REQ-REA-11, S11, T23).
- J-14 | resolved(Gate 1 유지) | behaviour-spec 범위 | attempt 3 | 로컬 전용(`.gitignore`)으로 정했다. 공유 저장소에 올리려면 사용자가 ignore 줄을 지우거나 별도 정책 결정이 필요하다.
- J-22 | resolved(Gate 1 유지) | `--runtime` route 인자 | attempt 3 | 동의 허용 경로를 S19에서 결정적으로 일으키려고 더했다. 동의 자체는 그대로 메인 세션 질문이다.
- J-07 | resolved(Gate 1 유지) | handoff 순서 | attempt 2 | 사용자 흐름(확인 → executor → tester) 앞에 tester Test Scaffold를 둔다. executor가 GREEN만 맡는 기존 계약 때문이다.

## Traceability Matrix

| Requirement | Plan Task | Acceptance Scenario | Semantic Invariant |
|-------------|-----------|---------------------|--------------------|
| REQ-REA-01 | T3, T9 | S1, S2 | INV-REA-01 |
| REQ-REA-02 | T9, T10, T11 | S3 | INV-REA-01 |
| REQ-REA-03 | T4, T5, T6, T7, T9 | S4, S5, S16 | INV-REA-03, INV-REA-11 |
| REQ-REA-04 | T4, T5, T6, T7, T9 | S2, S5 | INV-REA-04 |
| REQ-REA-05 | T8, T9 | S6 | INV-REA-05 |
| REQ-REA-06 | T12 | S7 | INV-REA-01, INV-REA-02 |
| REQ-REA-07 | T12 | S7 | INV-REA-02 |
| REQ-REA-08 | T4, T13 | S8 | INV-REA-03 |
| REQ-REA-09 | T9, T13, T14 | S3, S8, S9 | - |
| REQ-REA-10 | T1, T14 | S9, S19 | INV-REA-06, INV-REA-12 |
| REQ-REA-11 | T15, T16, T23 | S9, S10, S11 | INV-REA-06, INV-REA-07, INV-REA-13 |
| REQ-REA-12 | T17 | S12 | INV-REA-08 |
| REQ-REA-13 | T18 | S13 | INV-REA-09 |
| REQ-REA-14 | T18, T20 | S3, S13, S19 | INV-REA-09 |
| REQ-REA-15 | T19 | S14 | INV-REA-11 |
| REQ-REA-16 | T3, T5, T6, T7 | S15 | INV-REA-10 |
| REQ-REA-17 | T9, T11, T12 | S3, S7 | INV-REA-01 |
| REQ-REA-18 | T21 | S20 | - |
| REQ-REA-19 | T21 | S21 | INV-REA-13 |
| REQ-REA-20 | T2, T8, T10, T22 | S17, S18, S19 | INV-REA-06 |
| REQ-REA-21 | T14, T18 | S13, S19 | INV-REA-12 |
