# SPEC-REA-001: REA 역공학 MCP 통합 — 프로젝트 범위 등록, reverse-engineer 에이전트, `/auto reverse`

**Status**: draft
**Created**: 2026-10-15
**Domain**: REA
**Module**: autopus-adk
**Release**: v0.50.126(SPEC-MODELUX-003과 같은 릴리스). 작업 branch `feat/rea-001`, 기준 origin/main `4d73aa91`(PR #197 merge 뒤).
**Source PRD**: `prd.md`. 결정·근거는 `research.md`, 실행 증거는 `evidence/probe-rea-forward-compat.md`에 있다.

## 목적

사용자는 기존 앱·바이너리의 기능을 증거와 함께 분석해 자기 프로젝트에 다시 구현하고 싶다. 지금 하네스에는 역공학 도구의 등록, 분석 격리, 디컴파일 결과와 구현 사이의 clean-room 경계가 없다. 16개 에이전트 중 MCP 도구를 가진 것이 없어 분석이 메인 세션에서 돌고, 최대 약 10 MiB인 REA 결과가 메인 문맥을 채운다.

이 SPEC은 세 가지를 한다. (1) Node 조건을 충족한 프로젝트에서 Claude Code·Codex·Antigravity가 프로젝트 범위 REA MCP(`npx -y rea-agents@6.5.0 mcp`)를 띄운다. 새 init은 기본으로 켜고, 기존 프로젝트는 대화형 update에서 한 번만 묻는다. (2) 늘 생성되는 17번째 에이전트 `reverse-engineer`가 REA 도구와 읽기 도구만으로 분석해 evidence_id가 붙은 행동 명세와 unknowns를 낸다. (3) `/auto reverse`가 분석 → 명세 확인 gate → executor 구현 → tester 검증을 잇는다.

## Outcome Boundary

- Outcome Lock: Node 조건을 충족한 새 `auto init` 프로젝트에서 Claude Code·Codex·Antigravity가 프로젝트 범위 REA MCP를 띄운다. `/auto reverse`가 reverse-engineer의 행동 명세(evidence_id, unknowns)를 보이고, 사용자가 확인한 명세만으로 executor가 구현하고 tester가 검증한다. 기존 프로젝트는 대화형 update에서 한 번만 묻고, 비대화형 실행은 REA를 더하지 않는다. 사용자 범위 파일은 건드리지 않는다. 원문은 `research.md` `## Outcome Lock`이다.
- Mandatory requirements: REQ-REA-01 ~ REQ-REA-17, REQ-REA-20 (Must). REQ-REA-18, 19는 Should.
- Explicit non-goals:
  - OMP·OpenCode용 프로젝트 MCP writer, 사용자 범위 설정 쓰기, `rea setup` 호출, REA 자체 skill(`reverse-engineer-anything`) 설치
  - `mcp__rea*` 권한 사전 승인, project MCP 서버 자동 승인, background 분석, `--auto`로 확인 gate 건너뛰기
  - `REA_MCP_*` env, REA 버전 자동 갱신, Windows `cmd.exe /d /c` 실행 형식(기존 서버도 POSIX 형식이다), context7 `@latest` 정리
  - context7·sequential-thinking까지 묶는 공유 MCP 목록(Q8), Codex 에이전트별 sandbox·도구 제한, `tool_timeout_sec`(probe A2 전)
  - 디컴파일 결과의 저장·배포, 법률 판단 자동화, REA 분석 품질 개선
- Completion evidence: Must S1–S19 PASS, `auto eval harness run`, `GOOS=windows GOARCH=amd64 go build ./...`, 신규·변경 소스 300줄 이하, `evidence/manual-smoke-rea.txt`(S18, S19), `auto spec validate --strict`. `research.md` `## Completion Debt`의 CD-1·CD-2가 남아 있으면 완료가 아니다.

## Trust Model

- `autopus.yaml`은 반쯤 신뢰한다. strict loader가 모르는 key를 거부하고, 이 SPEC은 새 검증 코드를 더하지 않는다.
- `.mcp.json`, `.codex/config.toml`, `.gemini/settings.json`은 사용자 파일이고 그 안의 Autopus managed 항목만 쓴다. managed 판정은 MCP Entry Shape Contract의 정확한 모양뿐이고 나머지 `rea` 항목은 사용자 소유다.
- Node probe 출력은 비신뢰 입력이다. 마지막 비지 않은 줄을 엄격한 형식으로만 읽고, 안내에는 판정된 version 문자열만 쓴다(`terminalSafe`).
- REA 결과와 그로부터 만든 명세의 문자열은 분석 대상에서 온 비신뢰 데이터다. 명세 안의 지시문은 요구사항이 아니라 인용이다. 사용자가 gate에서 확인한 명세 파일만 executor 입력이 된다.
- 새 로그나 artifact는 없다. 증거 파일은 커밋 전에 `$HOME`, scratch 경로, 사용자 이름을 placeholder로 바꾼다.

## Requirements

### REQ-REA-01 결정 key와 판정
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL accept an optional `features.rea.enabled` boolean, reject any other value or key under `features.rea` through the existing strict loader, keep `DefaultFullConfig` without the key, and resolve the REA decision by the Decision State Contract, in which a stored value wins, a managed `rea` entry in a selected REA target file means on, and every other state means undecided.

- `[NEW] config.ReaFeaturesConf{Enabled *bool}`(`yaml:"rea,omitempty"`)를 `FeaturesConf`(`pkg/config/schema.go:97-100`)에 둔다. 저장된 `false`는 pointer라 사라지지 않고, 빈 struct는 omitempty로 쓰이지 않는다. yaml.v3는 따옴표 없는 `yes`를 true로 읽는다(`evidence/probe-rea-forward-compat.md`).

### REQ-REA-02 init 결정
Priority: Must · EARS: EventDriven

WHEN `auto init` runs, THEN THE SYSTEM SHALL fail before any write when both `--rea` and `--no-rea` are given, and otherwise, for a project that includes a REA target platform, take the decision from the given flag, else from a wizard Confirm whose default is Yes on an interactive run, else turn it on, and write it by the Decision State Contract.

- REA 대상 플랫폼은 claude-code, codex, antigravity-cli다. 하나도 없으면 묻지 않고 결정도 쓰지 않는다(OMP·OpenCode 안내만, REQ-REA-09). wizard 단계 모델은 `buildReviewGateStep`(`internal/cli/tui/wizard_steps.go:194-210`), flag 자리는 `internal/cli/init.go:241-246`이다.

### REQ-REA-03 managed 항목 쓰기
Priority: Must · EARS: StateDriven

WHERE the REA decision is on and the Node gate passes, THEN THE SYSTEM SHALL write exactly one entry of the MCP Entry Shape Contract into `.mcp.json`, `.codex/config.toml`, and `.gemini/settings.json` of the selected platforms through each adapter's merge path, replace a managed entry with the current shape, leave an `other_pin` or `user_owned` entry byte-identical, and write nothing outside the project root.

### REQ-REA-04 제거와 clean
Priority: Must · EARS: EventDriven

WHEN the REA decision becomes off or `auto clean` runs, THEN THE SYSTEM SHALL remove every managed `rea` entry from `.mcp.json`, `.codex/config.toml`, and `.gemini/settings.json`, including the Codex `[mcp_servers.rea]` header with the blank line Autopus put before it and the Antigravity `mcpServers.rea` key, so that removing an entry this SPEC added restores the previous bytes, and SHALL leave `other_pin` and `user_owned` entries byte-identical.

- 지금 Antigravity merge는 key를 지우지 못하고(`pkg/adapter/antigravity/antigravity_settings.go:187-198`) clean은 현재 렌더와 같은 서버만 지운다(`antigravity_lifecycle_clean.go:92-101`). Codex merge는 render에 없는 managed key를 지우지만 header 줄은 남긴다(`pkg/adapter/codex/codex_config_structural.go:48-127`). 고치는 방법은 MCP Entry Shape Contract의 쓰기 규칙이다.

### REQ-REA-05 Node gate
Priority: Must · EARS: Unwanted

IF the REA decision is on and the Node probe reports absent, unknown, or a version outside `^22.19.0 || ^24.11.0 || >=26.0.0`, THEN THE SYSTEM SHALL add no `rea` entry, print one Node notice of the Form Text, keep every existing managed entry and the stored decision, and run the probe at most once per run within 5 seconds.

### REQ-REA-06 update에서 한 번 묻기
Priority: Must · EARS: EventDriven

WHEN `auto update` runs on a project whose REA decision is undecided and that includes a REA target platform, THEN THE SYSTEM SHALL ask once with default Yes on an interactive TTY run without `--yes`, `--auto`, or preview, write and apply the answer in the same run, and on every other run add nothing, write nothing, read no stdin, and print one hint naming `auto update --rea`.

- 기존 질문 패턴은 `promptLanguageSettings`(`internal/cli/prompts.go:66-93`, 호출 `update.go:185-188`)다. workspace update는 `--yes`가 필요하고(`update_workspace.go:28-29`) repo마다 `runNestedUpdateCommand`를 돌므로 repo마다 hint 1줄이다.

### REQ-REA-07 update flag
Priority: Must · EARS: EventDriven

WHEN `auto update` receives `--rea` or `--no-rea`, THEN THE SYSTEM SHALL apply that decision without asking, write it by the Decision State Contract, and fail before any write when both flags are given or either is combined with `--workspace`.

- `--workspace`와의 조합을 거부하는 이유: nested update는 이 flag를 넘기지 않고 `update_workspace.go`는 이미 300줄이다. repo마다 `--local`로 실행한다.

### REQ-REA-08 doctor
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL report in `auto doctor` text and `--json` the REA decision and its source, the Node probe state and version, the ownership class of each selected REA target platform, and the `rea setup` guidance for OMP and OpenCode, without starting the REA server or making any network call.

### REQ-REA-09 OMP·OpenCode 안내
Priority: Must · EARS: StateDriven

WHERE the project includes OMP or OpenCode and the REA decision is not off, THEN THE SYSTEM SHALL write no REA MCP configuration for them and print one guidance line naming `npx -y rea-agents@6.5.0 setup` in init, update, and doctor.

### REQ-REA-10 reverse-engineer 에이전트
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL ship `content/agents/reverse-engineer.md` by the Reverse Engineer Agent Contract, so that every platform generates it whatever the REA decision, its Claude Code and Antigravity tools are the REA server and read tools only, no platform's file grants Write, Edit, or Bash, and its body demands a behaviour specification with evidence_id references and unknowns, never decompiled pseudocode or assembly, and ends with the unavailable-REA sentence when REA tools are missing.

### REQ-REA-11 17번째 에이전트 파급
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL add `reverse-engineer` to the canonical agent list and to every consumer in the Agent Set Contract, so that each platform generates 17 agents equal to `CanonicalAgentNames()`, OMP autopus mode routes the 17 agents and task, native OMP folds it into task, and the Preset Contract holds.

### REQ-REA-12 MCP 참조 변환 예외
Priority: Must · EARS: EventDriven

WHEN a skill, rule, command, workflow, or agent body is rendered for a platform other than Claude Code, THEN THE SYSTEM SHALL keep every `mcp__rea` and `mcp__rea__<tool>` reference verbatim and keep rewriting every other `mcp__*` reference as the base commit does.

- 지금 `pkg/content/skill_transformer_replace_mcp.go:22,50`은 남은 모든 `mcp__*`를 `Context7 MCP with WebSearch fallback`으로 바꾸고, `skill_transformer_replace.go:124`가 non-Claude 렌더마다(agent 본문 포함) 부른다.

### REQ-REA-13 `reverse` 라우트
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL add the `reverse` route to the router and route registry of all five platforms with one detail that follows the Reverse Route Contract.

### REQ-REA-14 clean-room handoff
Priority: Must · EARS: EventDriven

WHEN the operator confirms a behaviour specification at the gate of `/auto reverse`, THEN THE SYSTEM SHALL have the main session write only that specification to `.autopus/reverse/<slug>/behaviour-spec.md`, hand the executor only that file and the user's project context, and keep REA tool results, evidence bundles, and decompiled text out of the executor input and out of every project file.

### REQ-REA-15 사전 승인 없음
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL NOT write any `mcp__rea` permission, `enableAllProjectMcpServers`, `enabledMcpjsonServers`, Codex `enabled_tools`, or REA approval-mode key into any generated file.

- 지금 사전 승인 MCP는 `mcp__sequential-thinking__sequentialthinking` 하나다(`pkg/content/hooks.go:196`). project 서버 자동 승인 key는 어디에도 없어 Claude Code가 처음 세션에서 rea 서버 승인을 묻는다.

### REQ-REA-16 꺼져 있으면 bytes 그대로
Priority: Must · EARS: Ubiquitous

WHILE the REA decision is undecided or off and no managed `rea` entry exists, THE SYSTEM SHALL generate `.mcp.json`, `.codex/config.toml`, and `.gemini/settings.json` byte-identical to the output of the base commit `4d73aa91` for the same configuration.

### REQ-REA-17 옛 바이너리 줄
Priority: Must · EARS: EventDriven

WHEN a write adds `features.rea` to an autopus.yaml that did not hold it, THEN THE SYSTEM SHALL print the old-binary line of the Form Text once in that run and in no other run.

- strict loader(`pkg/config/loader_strict.go:98-125`)가 v0.50.113부터 모르는 key를 거부한다. auto 0.50.123은 `features.rea`가 든 파일에서 `auto update --yes`를 종료 코드 1로 끝냈다(probe A3).

### REQ-REA-18 locale
Priority: Should · EARS: Ubiquitous

THE SYSTEM SHALL render every new prompt, notice, hint, and doctor line in Korean when tui.DetectSystemLang returns ko and in English otherwise, while stored values stay locale-independent.

### REQ-REA-19 문서
Priority: Should · EARS: Ubiquitous

THE SYSTEM SHALL document the REA default, `--rea` and `--no-rea`, the Node range, the lazy `features.rea` write and its old-binary limit, the clean-room flow, and the runtime-capture caution in init, update, and doctor help, README, and the v0.50.126 CHANGELOG, without removing sentences existing tests check.

### REQ-REA-20 test seam과 수동 smoke
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL run every new path through injectable seams (`[NEW] reaNodeProbe`, the existing TTY probe `composeTTYProbe`, `composeFormsProbe`, `composeRunForm`) with no real `node`, `npx`, or network call in tests, and SHALL record the real-terminal smoke S18 and the consented live run S19 in `evidence/manual-smoke-rea.txt`.

## Contracts

### Decision State Contract

| 상태 | 판정 | doctor 출처 |
|---|---|---|
| on | `features.rea.enabled: true` | `features.rea` |
| off | `features.rea.enabled: false` | `features.rea` |
| on | key 없음, 선택된 REA 대상 플랫폼의 파일 중 하나 이상에 managed `rea` 항목 | `managed_entry` |
| undecided | 그 밖 전부(사용자 소유·other_pin 항목만 있는 경우 포함) | `none` |

쓰기 표. M은 "이번 실행 뒤 managed 항목이 하나 이상 남는가"이고 CLI가 쓰기 전에 정한다: action이 add이고 대상 파일 중 absent·managed가 하나 이상이면 있음, action이 keep이면 지금 managed 항목이 있으면 있음.

| 결정 | M | autopus.yaml |
|---|---|---|
| on | 있음 | key가 없으면 쓰지 않는다(항목이 결정을 담는다). `false`면 `true`로 고친다 |
| on | 없음(gate 실패, 대상 파일이 모두 사용자 소유) | `enabled: true`를 쓴다 |
| off | - | `enabled: false`를 쓴다. 이미 같으면 쓰지 않는다 |
| undecided | - | 쓰지 않는다 |

- key를 지우는 경로는 없다. 사용자가 지운 key는 위 판정을 다시 탄다.
- adapter 입력은 `[NEW] config.RuntimeConf.ReaMCP`(`[NEW] config.ReaMCPAction`, yaml `-`, `pkg/config/runtime.go:13-15` 옆)다. zero 값 `keep`, `add`, `remove`. CLI가 결정과 gate로 정한다: on + pass → add, on + gate 실패 → keep, off → remove, undecided → keep. `auto clean`은 늘 remove다. CLI를 거치지 않는 생성(`DefaultFullConfig` test, eval harness)은 keep이라 아무것도 넣거나 지우지 않는다.
- 이 표를 고른 근거와 대안 비교는 `research.md` `### Q1`이다.

### MCP Entry Shape Contract

| 플랫폼 | 자리 | managed 모양(pin 6.5.0) |
|---|---|---|
| Claude Code | `.mcp.json` `mcpServers.rea` | `{"$comment": "REA reverse-engineering MCP (rea-agents@6.5.0)", "command": "/bin/bash", "args": ["-l", "-c", "exec npx -y rea-agents@6.5.0 mcp"]}` |
| Antigravity | `.gemini/settings.json` `mcpServers.rea` | `{"command": "/bin/bash", "args": ["-l", "-c", "exec npx -y rea-agents@6.5.0 mcp"]}` |
| Codex | `.codex/config.toml` 끝 | 빈 줄 하나, `[mcp_servers.rea]`, `command = "npx"`, `args = ["-y", "rea-agents@6.5.0", "mcp"]`, `startup_timeout_sec = 30` |

| 분류 | 조건 | add | remove·clean | doctor |
|---|---|---|---|---|
| absent | `rea` 이름이 없다 | 넣는다 | 할 일 없음 | `absent` |
| managed | `[NEW] reamcp.ManagedPins`(`["6.5.0"]`, 이전에 낸 pin 전부) 중 하나의 모양과 정확히 같다. JSON은 deep-equal, Codex는 header 아래 빈 줄을 뺀 줄이 세 줄과 순서까지 같다 | 현재 pin 모양으로 바꾼다 | 지운다 | `managed` |
| other_pin | 모양은 같고 pin 자리만 목록 밖(`6.4.0`, `latest`) | 손대지 않는다 | 손대지 않는다 | `other_pin` |
| user_owned | 그 밖(다른 command·args, 더한 key나 `env`, 다른 `$comment`, 주석 줄, 빠진 key) | 손대지 않는다 | 손대지 않는다 | `user_owned` |

- 실행 형식은 그 파일의 기존 서버와 같다(Claude·Antigravity `/bin/bash -l -c "exec npx …"`, Codex `npx`).
- Claude: merge(`pkg/adapter/claude/claude_files.go:25-87`)의 managed 삭제 loop는 `rea`를 건너뛰고 rea는 action으로만 바뀐다. clean(`claude_clean.go:176-207`)은 remove와 같은 함수를 부른다.
- Codex: `codexManagedConfigKeys`(key 단위, `codex_config_structural.go:10-27`)에 넣지 않는다. 넣으면 사용자 `[mcp_servers.rea]`의 `command`·`args`를 덮는다. section 단위로 merge 뒤에 적용한다. remove는 header부터 다음 header 전까지와 header 바로 앞 빈 줄 하나를 지운다. clean(`codex_clean_merge.go:180-213`)은 generic clean 전에 같은 제거를 한다.
- Antigravity: `mergeSettingsMaps` 뒤에 rea를 적용해 key를 지울 수 있다. clean은 현재 렌더와 무관하게 managed 모양이면 지운다.
- `tool_timeout_sec`는 쓰지 않는다(Codex 기본 60초). probe A2가 필요를 보이면 scope expansion 규칙으로 더한다.

### Node Gate Contract

| probe 결과(stdout의 마지막 비지 않은 줄, 앞뒤 공백 제거) | 판정 |
|---|---|
| `v<MAJOR>.<MINOR>.<PATCH>`이고 22.x이면서 22.19.0 이상, 24.x이면서 24.11.0 이상, 또는 MAJOR 26 이상 | pass |
| 같은 형식이고 위 범위 밖 | unsupported |
| 실행 파일 없음 | absent |
| 그 밖(형식 밖, prerelease 꼬리, 0이 아닌 종료, 5초 timeout) | unknown |

- probe는 POSIX에서 서버와 같은 `/bin/bash -l -c 'node --version'`, Windows에서 PATH의 `node --version`이다. 실행당 1회, 결정이 on이거나 질문에 Yes가 나왔을 때만 돌고 doctor는 늘 1회다. seam은 `[NEW] reaNodeProbe`다.
- gate는 추가만 막는다. managed 항목을 지우는 것은 결정 off와 clean뿐이라 Node가 다른 기계에서 update해도 항목이 오락가락하지 않는다.

### Update Prompt Contract

| 결정 \ 실행 | TTY 대화형(`--yes`·`--auto`·preview 아님) | non-TTY, `--yes`, `--auto`, workspace의 repo별 실행 | preview(`--plan`·`--preview`·`--dry-run`) |
|---|---|---|---|
| undecided | 질문 1회, 기본 Yes, 답을 이번 실행에 적용 | 질문 없음, stdin 읽기 0, 쓰기 없음, hint 1줄 | 쓰기 없음, hint 1줄 |
| on, off | 질문 없음, 결정 적용 | 같음 | 쓰기 없이 미리보기 |
| `--rea`, `--no-rea` | 질문 없음, flag 적용 | 같음 | 쓰기 없이 미리보기 |
| 두 flag 함께, 또는 flag + `--workspace` | 쓰기 전 오류 | 같음 | 같음 |

- REA 대상 플랫폼이 없으면 묻지 않고 hint도 없다(OMP·OpenCode 안내만).
- 질문 자리는 `promptLanguageSettings` 뒤, compose 단계 앞이다. 17번째 에이전트와 `reverse` 라우트는 결정과 무관하게 모든 쓰는 실행에서 생성된다.

### Doctor Contract

| check ID | detail 예상 값 | severity |
|---|---|---|
| `doctor.rea.decision` | `on (features.rea)`, `on (managed_entry)`, `off (features.rea)`, `undecided` | info |
| `doctor.rea.node` | `pass v24.11.0`, `unsupported v22.18.0`, `absent`, `unknown` | 결정 on이고 pass가 아니면 warning, 아니면 info |
| `doctor.rea.platform.<claude-code·codex·antigravity-cli>` | `managed`, `absent`, `other_pin`, `user_owned` | 결정 on이고 gate pass인데 `absent`면 warning, 아니면 info |
| `doctor.rea.guidance.<omp·opencode>` | `npx -y rea-agents@6.5.0 setup` | info |

- payload `data.rea`는 `{"decision", "decision_source", "node": {"status", "version"}, "platforms": {<platform>: <class>}, "guidance": [<platform>]}`이다. 선택되지 않은 플랫폼은 key가 없고, 결정이 off면 guidance는 빈 배열이다. 조립 자리는 `internal/cli/doctor_json.go:118-137`과 `doctor_text.go:15`다.

### Reverse Engineer Agent Contract

| 자리 | 값 |
|---|---|
| source frontmatter | `name: reverse-engineer`, `model: opus`, `effort: high`, `tools: Read, Grep, Glob, mcp__rea`, `permissionMode: default`, `maxTurns: 120`, skills 없음 |
| Claude `.claude/agents/autopus/reverse-engineer.md` | `tools: Read, Grep, Glob, mcp__rea`, `permissionMode: default`, `background` 없음. balanced 표준 배치면 `model: claude-opus-5-5`, `effort: high`, ultra면 `model: opus`, `effort: max` |
| Antigravity `.gemini/agents/autopus/reverse-engineer.md` | tools `read_file`, `grep_search`, `glob`, `mcp_rea_*` 순서 |
| Codex `.codex/agents/reverse-engineer.toml` | balanced `model = "gpt-6.1-sol"`, `model_reasoning_effort = "xhigh"`. Autopus Codex 에이전트 형식에는 도구 칸이 없어(`pkg/content/agent_transformer.go:113`) 세션 도구를 상속한다. Codex의 clean-room은 developer_instructions와 기존 approval policy가 맡는다(한계, `research.md` Open Risks R4) |
| OMP `.omp/agents/reverse-engineer.md` | `tools` = `glob`, `grep`, `read`(정렬), `mcp_`로 시작하는 값 없음. override가 없으면 `thinkingLevel: high` |
| OpenCode `.opencode/agents/reverse-engineer.md` | `"*": deny` 아래 allow는 `glob`, `grep`, `question`, `read`, `skill`, `task`뿐이고 `mcp__` key가 없다 |
| 본문 | (1) 입력은 대상 경로와 기능 이름이다. (2) runtime capture 계열 도구 전에 사용자 확인을 받는다. (3) 결과는 `## Behaviour`, `## Inputs and Outputs`, `## Evidence`(evidence_id 목록), `## Unknowns`, `## Acceptance Items` 다섯 절의 행동 명세뿐이다. (4) 디컴파일된 의사코드, assembly, 대상 소스 발췌를 명세에 넣지 않는다. (5) REA 도구가 없으면 Form Text의 미등록 문장을 쓰고 끝낸다 |

- Bash를 빼는 이유: REA는 CLI(`npx rea-agents …`)로도 같은 분석과 runtime capture를 돌릴 수 있어 Bash가 있으면 MCP 권한 질문을 우회하고 파일을 쓸 수 있다. 대가로 `file`·`strings` 같은 빠른 확인은 REA 도구가 대신한다.
- `permissionMode: plan`을 쓰지 않는다. plan mode는 읽기 전용 탐색이라 MCP 호출이 막힐 수 있다(probe A1). `default`는 매번 묻는다.
- Claude Code 문서(2026-10-15): `tools`는 `mcp__<server>`와 `mcp__<server>__*`를 받고, 해석되는 항목이 하나라도 있으면 실행된다. background subagent는 MCP 도구를 모두 받지만 built-in 도구 대부분을 잃으므로 foreground만 쓴다. REA가 등록되지 않아도 Read·Grep·Glob이 해석되므로 에이전트는 떠서 미등록 문장을 낸다.
- OMP·OpenCode에는 MCP 이름을 쓰지 않는다(Q11). Antigravity는 Gemini CLI 문서의 `mcp_<server>_*` wildcard를 쓴다.

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
| 생성 에이전트(Codex TOML, Gemini MD), OMP agent 파일·override key | `make generate-templates`와 정규 목록에서 파생 |

### Preset Contract

| 항목 | 값 |
|---|---|
| balanced 표준 배치 | `[NEW] balancedRungAnalysis`, tier `opus`. anthropic `anthropic/claude-opus-5-5`·thinking `high`, openai `openai-codex/gpt-6.1-sol`·thinking `xhigh`. 손으로 opus를 고른 사다리 값과 같다(`ClaudeAgentEffort("balanced","opus")` = `high`, `codexLadderProfile` opus = `gpt-6.1-sol`·`xhigh`) |
| `defaultBalancedAgentTiers()`, ultra | 둘 다 17개, reverse-engineer `opus` |
| Historical | `HB6` balanced 16개(7 fable + 9 sonnet, commit fbe502c0, v0.50.116–v0.50.125, runtimeStandard), `HU3` ultra 16개(7 fable + 9 opus, commit 001830fd, v0.50.115–v0.50.125) |
| 분류 | v0.50.125 기본값 16개 map은 Historical이고 기존 `u` 갱신이 reverse-engineer `opus`를 더한다. 그 전에도 reverse-engineer는 map에 없으므로 표준 배치를 받는다(`native_balanced.go:20-24`) |
| 탐색·보조 표시 | 기본값에서 `현재: 혼합 (opus 1 · sonnet 2)` |

### Reverse Route Contract

| 단계 | 규칙 |
|---|---|
| 진입 | `/auto reverse <대상> <기능>`(Codex `$auto reverse`, 다른 플랫폼은 그 문법). 인자가 없으면 묻고 끝낸다 |
| 1 분석 | reverse-engineer를 foreground로 보낸다(Claude `Agent(subagent_type="reverse-engineer", …)`, `run_in_background` 없음). REA가 없으면 그 에이전트의 미등록 문장으로 끝낸다 |
| 2 확인 gate | 명세를 보이고 `확인 / 수정 / 중단`을 묻는다. `--auto`도 멈춘다. 수정은 1로 돌아가고 중단은 아무것도 쓰지 않는다 |
| 3 저장 | 확인 뒤 메인 세션이 명세만 `.autopus/reverse/<slug>/behaviour-spec.md`에 쓴다. slug는 기능 이름을 소문자 `[a-z0-9-]`로 바꾼 것이고 겹치면 `-2`, `-3` |
| 4 구현 | executor 입력은 그 파일 경로와 사용자 프로젝트 문맥뿐이다. REA 결과, evidence bundle, 디컴파일 텍스트는 넣지 않는다 |
| 5 검증 | tester가 명세의 `## Acceptance Items`로 검증한다 |
| 금지 | `export_evidence_bundle`을 프로젝트 안 경로로 부르지 않는다. 디컴파일 텍스트를 어떤 파일에도 쓰지 않는다 |

- 계약 문장은 `[NEW] templates/shared/auto-reverse-contract.md.tmpl` 한 곳에 두고 다섯 detail이 렌더한다: Claude `claude_workflow_skills.go:22-42`의 customBody(`[NEW] claude_reverse_route.go`), Antigravity `[NEW] templates/gemini/skills/auto-reverse/SKILL.md.tmpl`, Codex `codex_workflow_specs.go`·`codex_workflow_custom.go`, OMP `omp_specs.go`·`omp_workflow_render.go`, OpenCode `opencode_specs.go`·`opencode_workflow_custom.go`.
- router: Claude `templates/claude/commands/auto-router.md.tmpl:55-75`과 Antigravity root router에 `reverse` → `.claude/skills/auto-reverse/SKILL.md`(Antigravity는 그 경로) 1행, Codex router 설명(`codex_workflow_specs.go:17`) 끝에 `/reverse`. root router는 8192 bytes 이하다(지금 템플릿 6174 bytes).

### Form Text (ko / en)

- wizard: `[n/m] REA 역공학 MCP` · 설명 `Claude Code·Codex·Antigravity에 rea-agents@6.5.0을 프로젝트 범위로 등록합니다 (Node ^22.19.0 || ^24.11.0 || >=26.0.0)` / `[n/m] REA reverse-engineering MCP` · `Registers rea-agents@6.5.0 for Claude Code, Codex, and Antigravity in this project (Node ^22.19.0 || ^24.11.0 || >=26.0.0)`
- update 질문: `REA 역공학 MCP(rea-agents@6.5.0)를 이 프로젝트의 Claude Code·Codex·Antigravity에 등록할까요? [Y/n]` / `Register the REA reverse-engineering MCP (rea-agents@6.5.0) for Claude Code, Codex, and Antigravity in this project? [Y/n]`
- 등록: `REA 역공학 MCP를 등록했습니다: <파일 목록> (rea-agents@6.5.0)` / `Registered the REA reverse-engineering MCP: <files> (rea-agents@6.5.0)`
- hint: `REA 역공학 MCP를 아직 정하지 않았습니다. 켜려면 auto update --rea, 끄려면 auto update --no-rea` / `REA reverse-engineering MCP is undecided; run auto update --rea to enable or auto update --no-rea to decline`
- Node: `REA MCP를 등록하지 않았습니다: node <version>은 ^22.19.0 || ^24.11.0 || >=26.0.0 범위가 아닙니다` · `REA MCP를 등록하지 않았습니다: node를 찾지 못했습니다 (^22.19.0 || ^24.11.0 || >=26.0.0 필요)` · `REA MCP를 등록하지 않았습니다: node 버전을 확인하지 못했습니다` / `REA MCP not registered: node <version> is outside ^22.19.0 || ^24.11.0 || >=26.0.0` · `REA MCP not registered: node not found (needs ^22.19.0 || ^24.11.0 || >=26.0.0)` · `REA MCP not registered: node version unreadable`
- 옛 바이너리: `features.rea는 v0.50.126보다 오래된 auto가 읽지 못합니다` / `auto older than v0.50.126 cannot read features.rea`
- OMP·OpenCode: `OMP·OpenCode에는 Autopus가 REA MCP를 등록하지 않습니다. 쓰려면 npx -y rea-agents@6.5.0 setup을 직접 실행하세요(사용자 범위)` / `Autopus does not register REA MCP for OMP or OpenCode; run npx -y rea-agents@6.5.0 setup yourself (user scope)`
- flag: `--rea와 --no-rea는 함께 쓸 수 없습니다` · `--rea·--no-rea는 --workspace와 함께 쓸 수 없습니다. repo마다 --local로 실행하세요` / `--rea and --no-rea cannot be used together` · `--rea and --no-rea cannot be combined with --workspace; run them per repository with --local`
- doctor text: `REA: 켜짐 (managed 항목)` · `REA: 켜짐 (autopus.yaml)` · `REA: 꺼짐 (autopus.yaml)` · `REA: 미결정` / `REA: on (managed entry)` · `REA: on (autopus.yaml)` · `REA: off (autopus.yaml)` · `REA: undecided`
- 에이전트 미등록 문장(본문은 한국어): `REA MCP 도구를 찾지 못했습니다. auto update --rea로 이 프로젝트에 등록하거나 npx -y rea-agents@6.5.0 setup으로 사용자 범위에 등록하세요.`

## Existing Test Changes

0건이 아니다. T2가 파일마다 분류하고(`[NEW] evidence/test-classification.md`) 규칙은 아래뿐이다.

1. 수 고정: 정규 에이전트 16 → 17, OMP 16 + task 17 → 18(`pkg/adapter/integration_test.go:214`, `latest_cli_omp_contract_test.go:33`, `codex/codex_agents_test.go:126`, `omp/omp_agents_test.go:83,118`, `omp_acceptance_surface_test.go:52,169`, `omp_latest_contract_test.go:30`, `omp_native_registry_test.go:32`, `omp_agent_surface_report_test.go:52`, `pkg/config/omp_agent_source_test.go:67`, `role_model_policy_pins_test.go:125`, `internal/cli/` compose·migration test 등). literal만 바꾼다.
2. frozen route 목록(`pkg/adapter/claude/router_budget_test.go:16-19`, antigravity 같은 test, `codex_surface_test.go`, `evals/harness/tasks/surface/GT-ROUTE-*`)에 `reverse`를 더한다.
3. compose golden(`internal/cli/testdata/quality_compose/*.golden`): reverse-engineer 행 추가, 탐색·보조 멤버 목록과 `현재:` 표시만 바뀐다. 다른 diff는 회귀다.
4. v0.50.125 기본값 16개를 literal로 쓴 fixture는 HB6·HU3 Historical로 분류된다. 기존 갱신 줄·`u` 선택지가 나타나는 자리만 바뀐다.
5. MCP inventory test(`GT-SETTINGS-MCP-SERVERS.json`, `antigravity_settings_test.go`, `antigravity_clean_ownership_test.go`, `claude_lifecycle_contract_test.go`, `codex_config_test.go`, `codex_latest_v2_test.go`)는 바꾸지 않는다(REQ-REA-16).
6. CLI test 바이너리의 `TestMain`(`internal/cli/pipeline_backend_omp_rpc_test.go:36`)이 `reaNodeProbe`를 `v24.11.0`으로 고정한다. 새 init 출력이나 `.mcp.json` bytes를 통째로 비교하는 test는 등록 줄과 managed 항목만 더해진다.

## 생성 파일 상세

- `pkg/config`: `schema.go`(+1줄), `runtime.go`(+1줄), `[NEW] rea_feature.go`(+test), `quality_tier.go`, `role_model_policy_balanced.go`, `native_balanced.go`, `defaults.go`, `quality_preset_history.go`, `role_model_policy_matrix.go`, `omp_native_agents.go`, `omp_agent_source.go`(주석).
- `[NEW] pkg/reamcp`: `shape.go`(+test, 모양·pin·분류), `node.go`(+test, 버전 판정). Go 의존성은 늘지 않는다.
- adapter: `[NEW] pkg/adapter/claude/claude_rea_mcp.go`(+test), `claude_files.go`, `claude_clean.go`, `[NEW] claude_reverse_route.go`, `claude_workflow_skills.go`; `[NEW] pkg/adapter/codex/codex_rea_mcp.go`(+test), `codex_settings.go`, `codex_clean_merge.go`, `codex_workflow_specs.go`, `codex_workflow_custom.go`; `[NEW] pkg/adapter/antigravity/antigravity_rea_mcp.go`(+test), `antigravity_settings.go`, `antigravity_lifecycle_clean.go`; `pkg/adapter/omp/omp_specs.go`, `omp_workflow_render.go`; `pkg/adapter/opencode/opencode_specs.go`, `opencode_workflow_custom.go`.
- content: `[NEW] content/agents/reverse-engineer.md`, `pkg/content/agent_transformer.go`, `agent_transformer_omp_render.go`, `agent_transformer_opencode.go`, `skill_transformer_replace_mcp.go`, `content/skills/adaptive-quality.md`, `[NEW] templates/shared/auto-reverse-contract.md.tmpl`, `[NEW] templates/gemini/skills/auto-reverse/SKILL.md.tmpl`, `templates/shared/omp-agent-pipeline.md.tmpl`, `templates/claude/commands/auto-router.md.tmpl`, Antigravity root router 템플릿.
- `internal/cli`: `[NEW] rea_decision.go`, `[NEW] rea_node_probe.go`, `[NEW] init_rea.go`, `[NEW] update_rea.go`, `[NEW] doctor_rea.go`, `[NEW] tui/wizard_rea_step.go`(모두 +test), `init.go`, `update.go`(292줄 → 300줄 이하), `tui/wizard_steps.go`, `doctor_json.go`, `doctor_text.go`, `quality_compose_groups.go`, `quality_help.go`, `init_help.go`, `update_help.go`, `platform_omp_agent_migration.go`, `quality_compose_form_edit.go`(주석).
- 문서: README, CHANGELOG(v0.50.126), help 상수. 증거: `evidence/probe-rea-forward-compat.md`(있음), `[NEW] evidence/test-classification.md`, `[NEW] evidence/manual-smoke-rea.txt`. 신규·변경 소스는 300줄 이하이고 `update.go`·`update_workspace.go`(300)·`doctor_json_checks.go`(293)·`wizard_steps.go`(268)에는 호출 몇 줄만 더하거나 손대지 않는다.

## Related SPECs

- SPEC-MODELUX-003(같은 릴리스, 미발행): 16을 17로만 넓힌다. OMP Agent File Contract는 17개 파일, Agent Set Contract key는 17 + task, Migration Contract 미리보기는 17줄과 `OMP는 이제 Autopus 에이전트 17개를 씁니다.`, Role Group Contract 탐색·보조는 3명, OMP Autopus Tab Contract 목록은 17개다. 옛 바이너리 경고 방식을 그대로 쓴다. 나머지 계약은 그대로다.
- SPEC-MODELUX-001: Historical Preset Registry에 HB6·HU3를 더한다. 분류 규칙(정확한 map 일치)은 그대로다.
- SPEC-MODELUX-002: Role Group Contract의 탐색·보조 멤버만 바뀐다.
- Sibling: 없다(`research.md` `## Sibling SPEC Decision`).

## Sibling SPEC Decision

없다. Primary SPEC 하나가 Outcome Lock을 닫는다. 근거와 후보 검토는 `research.md` `## Sibling SPEC Decision`이다.

## Completion Debt

- CD-1 실제 터미널 smoke(S18)와 CD-2 동의 뒤 live `/auto reverse`(S19)가 기록되기 전에는 완료가 아니다. 표는 `research.md` `## Completion Debt`다.

## Evolution Ideas

- 완료를 막지 않는 선택 개선만 `research.md` `## Evolution Ideas`에 둔다. 번호나 후속 문서로 승격하지 않는다.

## Open Issues

- Q1 | decision, Gate 1 확인 대상 | `features.rea` lazy 쓰기 | attempt 1 | ledger의 "features flag로 저장"을 "managed 항목이 결정을 담지 못할 때만 key를 쓴다"로 좁혔다(probe A3). 사용자가 "늘 쓴다"를 고르면 Decision State Contract 쓰기 표의 on 두 행을 `enabled: true`로 바꾸고 REQ-REA-17이 모든 새 init에서 한 줄을 낸다.

## Traceability Matrix

| Requirement | Plan Task | Acceptance Scenario | Semantic Invariant |
|-------------|-----------|---------------------|--------------------|
| REQ-REA-01 | T3, T9 | S1, S2 | INV-REA-01 |
| REQ-REA-02 | T10, T11 | S3 | INV-REA-01 |
| REQ-REA-03 | T4, T5, T6, T7, T9 | S4, S5, S16 | INV-REA-03, INV-REA-11 |
| REQ-REA-04 | T4, T5, T6, T7 | S5 | INV-REA-04 |
| REQ-REA-05 | T8, T9 | S6 | INV-REA-05 |
| REQ-REA-06 | T12 | S7 | INV-REA-01, INV-REA-02 |
| REQ-REA-07 | T12 | S7 | INV-REA-02 |
| REQ-REA-08 | T4, T13 | S8 | INV-REA-03 |
| REQ-REA-09 | T9, T13 | S3, S8 | - |
| REQ-REA-10 | T1, T14 | S9, S18 | INV-REA-06 |
| REQ-REA-11 | T15, T16 | S9, S10, S11 | INV-REA-06, INV-REA-07 |
| REQ-REA-12 | T17 | S12 | INV-REA-08 |
| REQ-REA-13 | T18 | S13 | INV-REA-09 |
| REQ-REA-14 | T18 | S13, S19 | INV-REA-09 |
| REQ-REA-15 | T19 | S14 | INV-REA-11 |
| REQ-REA-16 | T3, T5, T6, T7 | S15 | INV-REA-10 |
| REQ-REA-17 | T9, T11, T12 | S3, S7 | INV-REA-01 |
| REQ-REA-18 | T20 | S20 | - |
| REQ-REA-19 | T20 | S21 | - |
| REQ-REA-20 | T2, T8, T21 | S17, S18, S19 | INV-REA-06 |
