# SPEC-REA-001 구현 계획

## Tasks

기준은 origin/main `4d73aa91`이고 branch는 `feat/rea-001`이다. 순서: T1 → T2 → (T3, T4, T8) → T5·T6·T7(T3, T4 뒤) → T9(T4, T8 뒤) → T10 → T11·T12·T13(T9 뒤) → T15 → T14·T16(T15 뒤) → T23(T15 뒤) → T17 → T18·T20 → T19 → T21 → T22. `update.go`(292줄), `update_workspace.go`(300), `doctor_json_checks.go`(293), `wizard_steps.go`(268)에는 호출 몇 줄만 더하거나 손대지 않는다. round 1 반영표는 `evidence/review-rounds.md`다.

- [ ] T1: risk-first probe. A1·A2를 T5–T7·T14 전에 돌릴 수 있으면 돌려 아래 표와 `evidence/`에 남긴다. 돌리지 못하면 not-run으로 두고 문서 기본값으로 진행하되 CD-1·CD-2는 S18·S19에서 그 행이 PASS가 될 때만 닫힌다. 소스 변경 없음 (REQ-REA-03, 10, 20).
- [ ] T2: 기존 test 분류. 수 고정, frozen route·command 목록, compose golden, literal 16개 fixture, MCP inventory, CLI init 출력 비교, OMP catalog fixture, TTY 입력 Enter를 spec.md Existing Test Changes 규칙 1–8로 파일마다 정해 `[NEW] evidence/test-classification.md`에 남긴다. 소스 변경 없음 (REQ-REA-20).
- [ ] T3: 결정 key. 소유 `pkg/config/schema.go`(+1줄), `pkg/config/runtime.go`(+1줄), `[NEW] pkg/config/rea_feature.go`(+test). `ReaFeaturesConf{Enabled *bool}`, `ReaMCPAction`(zero `keep`, `add`, `remove`), `RuntimeConf.ReaMCP`(yaml `-`). 거부는 기존 strict loader가 한다 (REQ-REA-01, 16).
- [ ] T4: REA 모양·분류·변환. 소유 `[NEW] pkg/reamcp/shape.go`(+test): `Pin`, `ManagedPins`, JSON 모양(Claude 두 pin 자리 일치), Codex 서버 table(`[mcp_servers.rea]`와 하위 table, `codexTOMLScanState`와 같은 줄 규칙), 분류, JSON map·Codex text의 add·remove 변환 (REQ-REA-03, 04, 08).
- [ ] T5: Claude. 소유 `[NEW] pkg/adapter/claude/claude_rea_mcp.go`(+test), `claude_files.go`(managed 삭제 loop에서 rea 제외, action 적용), `claude_clean.go`(Clean = remove). REA가 내용을 바꾸는 쓰기는 `adapter.SafeWorkspacePath`를 거친다 (REQ-REA-03, 04, 16).
- [ ] T6: Codex. 소유 `[NEW] pkg/adapter/codex/codex_rea_mcp.go`(+test), `codex_settings.go`(merge 뒤 table 적용), `codex_clean_merge.go`(generic clean 전에 managed table과 앞 빈 줄 제거). `codexManagedConfigKeys`에는 넣지 않는다. `SafeWorkspacePath` 확인 (REQ-REA-03, 04, 16).
- [ ] T7: Antigravity. 소유 `[NEW] pkg/adapter/antigravity/antigravity_rea_mcp.go`(+test), `antigravity_settings.go`(merge 뒤 적용, key 삭제), `antigravity.go`(Generate의 settings 쓰기 L130-137), `antigravity_lifecycle_clean.go`(현재 렌더와 무관하게 managed 제거). REA가 내용을 바꾸는 실행에서 settings 쓰기 전에 `SafeWorkspacePath`로 확인하고 거부되면 오류로 끝낸다 (REQ-REA-03, 04, 16).
- [ ] T8: Node gate. 소유 `[NEW] pkg/reamcp/node.go`(+test: exec 결과 판정, 127 규칙), `[NEW] internal/cli/rea_node_probe.go`(+test: seam `reaNodeExec`, POSIX login shell·Windows PATH, 5초, 실행당 1회 memo), `internal/cli/pipeline_backend_omp_rpc_test.go`의 `TestMain`이 seam을 `v24.11.0`·종료 0으로 고정 (REQ-REA-05, 20).
- [ ] T9: 결정 판정. 소유 `[NEW] internal/cli/rea_decision.go`(+test): `resolveReaDecision`(key > managed 항목 > undecided, 선택된 대상 파일만), action, 쓰기 뒤 다시 읽는 M(preview는 계획된 mapping, `ActionSkip` 제외), 쓰기 표, off sweep(선택되지 않은 파일은 `pkg/reamcp` 변환 + `SafeWorkspacePath` + `WriteFileIfChanged`), 재-init용 key 이전 helper, 등록·옛 바이너리·Node·OMP·OpenCode 줄, Form Text ko/en (`composeLanguage`) (REQ-REA-01–05, 09, 17).
- [ ] T10: wizard. 소유 `[NEW] internal/cli/tui/wizard_rea_step.go`(+test: huh accessible mode와 주입 reader·writer로 REA form을 직접 돌린다), `wizard_steps.go`(`InitWizardOpts`·`InitWizardResult` 1줄씩, reviewgate 뒤 entry, `defaultResult` 기본 Yes). flag·기존 결정이 있거나 대상 플랫폼이 없으면 단계를 넣지 않는다 (REQ-REA-02, 20).
- [ ] T11: init. 소유 `[NEW] internal/cli/init_rea.go`(+test, `runInitWizard` 결과 stub), `init.go`(flag 2줄, `DefaultFullConfig` 자리(L68-69)에서 기존 `features.rea` 이전, wizard 결과·비대화형 기본 자리 호출). 두 flag는 Step 1 전에 거부한다 (REQ-REA-02, 17).
- [ ] T12: update. 소유 `[NEW] internal/cli/update_rea.go`(+test), `update.go`(flag 등록 helper 1줄, preview 분기(L112) 전 flag·기존 결정 해석, workspace 분기 전 거부, `promptLanguageSettings` 뒤 질문; 300줄 이하), `update_preview.go`(`cloneHarnessConfig`가 `Runtime.ReaMCP`를 옮김). 질문 판정은 `composeGate.mode()`(CI 포함). Update Prompt Contract 전부 (REQ-REA-06, 07, 17).
- [ ] T13: doctor. 소유 `[NEW] internal/cli/doctor_rea.go`(+test), `doctor_json.go`(+1 호출), `doctor_text.go`(+1 호출). Doctor Contract (REQ-REA-08, 09).
- [ ] T14: 에이전트 본문과 도구 변환. 소유 `[NEW] content/agents/reverse-engineer.md`(다섯 절, ID, Runtime Capture Request, 플랫폼별 미등록 문장, 사용자에게 묻지 않음), `pkg/content/agent_transformer.go`(Gemini 허용 목록 `mcp__rea` → `mcp_rea_*`), `agent_transformer_omp_render.go`·`agent_transformer_opencode.go`(`mcp__rea` 제외), `make generate-templates` 산출물 (REQ-REA-09, 10, 21).
- [ ] T15: config 파급. 소유 `pkg/config/quality_tier.go`, `role_model_policy_balanced.go`, `native_balanced.go`, `defaults.go`, `quality_preset_history.go`, `role_model_policy_matrix.go`, `omp_native_agents.go`, `omp_agent_source.go`(주석). Agent Set·Preset Contract (REQ-REA-11).
- [ ] T16: CLI·문장·fixture 파급. 소유 `internal/cli/quality_compose_groups.go`, `quality_help.go`, `init_help.go`, `update_help.go`, `platform_omp_agent_migration.go`, `quality_compose_form_edit.go`(주석), `templates/shared/omp-agent-pipeline.md.tmpl`, `content/skills/adaptive-quality.md`, compose golden 재생성, OMP catalog fixture 세 개(규칙 7) (REQ-REA-11).
- [ ] T17: MCP 참조 예외. 소유 `pkg/content/skill_transformer_replace_mcp.go`(+test, `mcp__reader__x` 부정 사례 포함). `__`로 나눈 첫 segment가 정확히 `rea`인 참조만 건너뛴다 (REQ-REA-12).
- [ ] T18: `/auto reverse`. 소유 `[NEW] templates/shared/auto-reverse-contract.md.tmpl`(`--runtime`, 동의 loop, gate, slug, handoff, 금지), `internal/cli/init_helpers.go`(`gitignorePatterns`에 `.autopus/reverse/`), `[NEW] pkg/adapter/claude/claude_reverse_route.go`, `claude_workflow_skills.go`(route 1행), `templates/claude/commands/auto-router.md.tmpl`(1행), Antigravity root router 템플릿(1행), `[NEW] templates/gemini/skills/auto-reverse/SKILL.md.tmpl`, `[NEW] templates/gemini/commands/auto/reverse.toml.tmpl`, `pkg/adapter/codex/codex_workflow_specs.go`(spec, router 설명), `codex_workflow_custom.go`, `pkg/adapter/omp/omp_specs.go`, `omp_workflow_render.go`, `pkg/adapter/opencode/opencode_specs.go`, `opencode_workflow_custom.go` (REQ-REA-13, 14, 21).
- [ ] T19: 사전 승인 없음. 소유 `[NEW] pkg/content/hooks_rea_test.go`와 S14 생성 test. `hooks.go`는 바꾸지 않는다 (REQ-REA-15).
- [ ] T20: handoff 입력. 소유 `content/agents/executor.md`(Step 0과 입력 절), `content/agents/tester.md`(입력 절)에 Reverse Handoff Contract 문장 한 줄씩, 생성 산출물 (REQ-REA-14).
- [ ] T21: Should. Form Text ko/en, init·update·doctor help, README, CHANGELOG v0.50.126(REQ-REA-19 목록). 생성 표면은 `content/`·`templates/` 수정 뒤 `make generate-templates` (REQ-REA-18, 19).
- [ ] T23: OMP analysis rung 대체(Gate 1 사용자 결정). 소유 `[NEW] pkg/config/omp_rea_fallback.go`(+test: `OMPReaAnalysisFallback`, built-in balanced의 새 rung이고 agent가 reverse-engineer이며 후보가 catalog에 없을 때만 참), `[NEW] internal/cli/platform_omp_rea_fallback.go`(+test, `platform_omp_profile_preview.go` blocker loop에서 호출: 그 행은 blocker 대신 대체 줄), `[NEW] pkg/adapter/omp/omp_model_rea_fallback.go`(+test, compile·활성화에서 reverse-engineer route 제외). 기존 rung과 operator pin의 blocker는 그대로다 (REQ-REA-11, 19).
- [ ] T22: 회귀와 smoke. `go test ./...` 필터 없이, test diff와 T2 분류 대조, `auto eval harness run`, `GOOS=windows GOARCH=amd64 go build ./...`, 300줄, golangci-lint, generate 무 diff. S18·S19를 `evidence/manual-smoke-rea.txt`에 치환해 남기고 A1·A2 행을 그 결과로 갱신한다. 하나라도 빠지거나 FAIL이면 CD-1·CD-2가 남는다 (REQ-REA-20).

## Implementation Strategy

- 판단은 한 곳씩이다. 결정 판정·action·M은 CLI `rea_decision.go`, 모양·분류·변환은 `pkg/reamcp`, adapter는 `RuntimeConf.ReaMCP`만 받는다. zero 값이 keep이라 CLI를 거치지 않는 생성은 아무것도 넣거나 지우지 않는다. clone은 그 값을 옮긴다.
- 기존 merge 경로를 넓힌다. Claude는 managed 삭제 loop에서 rea를 빼고 action을 적용하고, Codex는 merge 뒤 서버 table 단위로, Antigravity는 merge 뒤 map에서 적용한다. Clean은 remove와 같은 함수다. context7·sequential-thinking 템플릿은 바꾸지 않아 REQ-REA-16이 구조로 성립한다. JSON 서식은 기존 merge의 정규화를 그대로 받는다.
- 동의와 확인은 메인 세션 skill이 받는다. 에이전트는 묻지 않고 요청을 돌려준다. 그래서 subagent의 실행 모드나 permissionMode에 기대지 않는다.
- 결정 key는 managed 항목이 결정을 담지 못할 때만 쓴다(A3 실행 증거). 17번째 에이전트는 정규 목록에서 파생되는 소비처를 그대로 두고 고정된 곳만 고친다.
- 새 Go 의존성은 없다. Node 범위는 `strconv`로 직접 비교한다. rea-agents는 사용자의 AI 도구가 띄울 때 npx가 받는 runtime 의존성이다.

## Visual Planning Brief

init·update·doctor command-flow.

```mermaid
flowchart TD
  A["auto init / auto update"] --> F{"--rea / --no-rea"}
  F -->|"둘 다, 또는 flag + --workspace"| X["오류, 아무것도 쓰지 않음"]
  F -->|"하나"| DEC["그 결정"]
  F -->|"없음"| S{"Decision State: features.rea > managed 항목 > undecided"}
  S -->|"on 또는 off(재-init은 key를 옮김)"| DEC
  S -->|"undecided, init"| W{"대화형?"}
  W -->|"예"| Q["wizard Confirm 기본 Yes"] --> DEC
  W -->|"아니오"| ON["on"] --> DEC
  S -->|"undecided, update"| U{"compose gate가 대화형이고 --auto, preview 아님?"}
  U -->|"예"| P["한 번 질문, 기본 Yes"] --> DEC
  U -->|"아니오"| H["rea 항목과 features.rea 쓰기 없음, hint 1줄"]
  DEC --> G{"결정"}
  G -->|"off"| R["action remove + 선택 안 된 파일 sweep, enabled: false"]
  G -->|"on"| N{"Node probe 1회"}
  N -->|"pass"| M["action add, 쓰기 뒤 M을 다시 읽음"]
  N -->|"fail"| K["action keep, 안내 1줄"]
  M --> KEY{"M 있음?"}
  K --> KEY
  KEY -->|"아니오"| T1["enabled: true"]
  KEY -->|"예"| T2["key 그대로"]
  DR["auto doctor"] --> DS["결정과 출처, Node, 플랫폼별 분류, OMP·OpenCode 안내, 쓰기 없음"]
```

`/auto reverse` sequence. 동의·확인은 메인 세션이 받고 clean-room 경계는 저장 단계다.

```mermaid
sequenceDiagram
  participant U as 사용자
  participant M as 메인 세션 skill
  participant R as reverse-engineer (REA MCP + Read, Grep, Glob)
  participant T as tester
  participant E as executor
  U->>M: /auto reverse <대상> <기능>
  M->>R: 분석 위임 (실행 모드 무관)
  R-->>M: Runtime Capture Request (필요할 때만)
  M->>U: 허용 / 거절 (--auto도 묻는다)
  U-->>M: 허용이면 동의 줄을 붙여 다시 분석, 거절이면 정적 분석만
  R-->>M: 행동 명세 B-n, A-n, evidence_id, unknowns
  M->>U: 확인 / 수정 / 중단 + 소유·테스트 경로
  U-->>M: 확인
  Note over M: clean-room 경계: 확인된 명세만 .autopus/reverse/<slug>/behaviour-spec.md
  M->>T: SPEC N/A, Task reverse/<slug>, A-n, 테스트 경로, Phase Test Scaffold
  M->>E: Behaviour Spec 경로, B-n·A-n, 소유·금지 범위
  M->>T: Phase Testing
  T-->>U: A-n별 판정
```

Low-fi wireframe(`wireframe intent: assumed`, 이번 세션에서 사용자가 확인하지 않았다). 요구사항으로 묶는 것은 Update Prompt Contract의 순서·기본값과 Form Text뿐이다.

```text
auto init (wizard)                                auto update (대화형, 미결정)
┃ [5/6] REA 역공학 MCP                              REA 역공학 MCP(rea-agents@6.5.0)를 이 프로젝트의
┃ Claude Code·Codex·Antigravity에 rea-agents@6.5.0을  Claude Code·Codex·Antigravity에 등록할까요? [Y/n]
┃ 프로젝트 범위로 등록합니다 (Node ...)
┃ > Yes   No
```

## Feature Completion Scope

- 이 SPEC 하나가 Outcome Lock을 닫는다. T3–T13은 결정·등록·제거·doctor, T14–T16은 에이전트와 파급, T17·T18·T20은 변환·라우트·handoff, T19는 권한, T21은 Should, T1·T2·T22는 probe·회귀·smoke다.
- 승인된 sibling 의존성은 없다. task 23개(25개 이하), 소스 파일은 test 포함 약 60개다. Completion Debt CD-1·CD-2는 T22의 S18·S19 기록과 A1·A2 PASS로만 닫힌다.

## Risk-First Integration Probe

| assumption_id | class | risk | boundary | input | oracle | isolation | status | reason | evidence |
|---------------|-------|------|----------|-------|--------|-----------|--------|--------|----------|
| A1 | implementation_assumption | high | subagent 도구 허용 목록과 project MCP 서버. Claude Code `tools: Read, Grep, Glob, mcp__rea`, Antigravity `mcp_rea_*` | S18 project, rea 서버 승인 1회, 동의 뒤 두 도구에서 reverse-engineer dispatch 1회씩(S19) | 두 transcript에서 reverse-engineer가 REA 도구를 1번 이상 부르고 Write, Edit, Bash 호출이 없다. 실행 모드가 background여도 같다 | 실제 터미널, scratch HOME, 모델 호출은 dispatch에만 | not-run | 실제 터미널과 모델 호출이 필요하다. 문서(2026-10-15)는 Claude의 server-level 이름과 Gemini CLI의 `mcp_server_*`를 확인했지만 project 서버 승인 뒤의 해석과 Antigravity CLI의 같은 규칙 여부는 실행 전이다. 다르면 T14를 고치고 CD-2가 남는다 | - |
| A2 | implementation_assumption | medium | Codex CLI, npx 첫 실행, REA stdio 서버. `startup_timeout_sec = 30`, `tool_timeout_sec` 없음(기본 60) | 다른 도구보다 먼저, 새 scratch HOME·CODEX_HOME과 Codex 전용 빈 `npm_config_cache`로 trusted scratch project에서 메시지 없는 세션 시작(S18), 동의 뒤 작은 정적 JS 대상 도구 호출 1회(S19) | 시작 전 두 cache 위치에 rea-agents가 없고 시작 뒤 Codex가 쓴 cache에 `rea-agents@6.5.0`이 생기며(cold 증명), 30초 안에 rea 도구가 보이고 도구 호출이 60초 안에 끝난다. cache가 비어 있으면 cold를 증명하지 못한 것이라 PASS가 아니다 | scratch HOME, CODEX_HOME, npm cache. 시작에는 모델 호출 없음 | not-run | 제3자 npm 코드를 받아 실행하므로 maintainer 동의가 필요하다. Codex 문서 기본값은 10초와 60초이고 30초는 REA `rea setup` 값(계획 문맥, 재확인 전)이다. 넘으면 `tool_timeout_sec`를 scope expansion으로 올리고 CD-1·CD-2가 남는다 | - |
| A3 | verified_fact | high | auto v0.50.113–v0.50.125 strict loader와 MCP merge가 이 SPEC이 쓰는 파일을 읽는가 | auto 0.50.123(c447badc), scratch HOME과 project: autopus.yaml + `features.rea.enabled: true`, autopus.yaml + `future_extension.rea`, 세 파일의 managed rea 모양 | features.rea는 doctor ERROR와 update 종료 1, future_extension은 읽히지만 update가 block을 지움, rea 항목은 update 뒤에도 같은 값 | scratch HOME과 project, 네트워크 없음 | PASS | 세 입력 모두 예측대로 관측했다. 그래서 결정은 managed 항목이 담고 key는 항목이 결정을 담지 못할 때만 쓴다(Decision State Contract) | `evidence/probe-rea-forward-compat.md` |

## Plan Statement Classification

| Statement | Class | Evidence |
|-----------|-------|----------|
| 결정 판정은 key > managed 항목 > undecided이고 질문은 undecided 대화형 update에서만 한 번이다 | requirement_invariant | REQ-REA-01, 06, S2, S7 |
| 비대화형 undecided 실행은 rea 항목과 `features.rea`를 쓰지 않고 REA stdin을 읽지 않는다 | requirement_invariant | REQ-REA-06, S7 |
| 결정이 없거나 꺼져 있으면 세 MCP 파일 bytes가 기준 commit과 같다 | requirement_invariant | REQ-REA-16, S15 |
| runtime capture는 메인 세션의 명시적 허용 뒤에만 부른다 | requirement_invariant | REQ-REA-21, S13, S19 |
| catalog에 analysis rung model이 없으면 reverse-engineer만 세션 model, 기존 rung은 blocker | requirement_invariant | REQ-REA-11, S11, Gate 1 사용자 결정 |
| `features.rea`가 든 파일에서 auto 0.50.123의 update가 종료 1로 끝난다 | verified_fact | A3 V1 |
| `future_extension` block은 옛 update의 typed Save가 지운다 | verified_fact | A3 V2, `loader.go:130-153` Read |
| managed 모양의 rea 항목은 옛 update 뒤에도 같은 값으로 남는다 | verified_fact | A3 P3 |
| Claude·Antigravity JSON merge는 파일을 decode한 뒤 다시 marshal한다 | verified_fact | `claude_files.go:25-87`, `antigravity_settings.go:18-60` Read |
| `cloneHarnessConfig`는 YAML round trip이라 `yaml:"-"` 필드를 잃는다 | verified_fact | `update_preview.go:20-35` Read |
| `runInitWizard`는 결과만 바꾸는 seam이고 `RunInitWizard`는 `form.Run()`을 직접 부른다 | verified_fact | `quality_compose_driver_gate.go:24-26`, `wizard_steps.go:41-66` Read |
| 공개 정리 명령은 `auto platform remove`이고 `auto clean`은 없다 | verified_fact | `internal/cli/platform.go:142-195`, root 명령 목록 `rg` |
| executor·tester는 SPEC/Task 식별자·요구사항·경로를 요구한다 | verified_fact | `content/agents/executor.md:104-106`, `tester.md:75-77` Read |
| subagent에는 AskUserQuestion이 없고 fork mode에서 foreground를 보장할 수 없다 | implementation_assumption | review round 1이 인용한 Claude Code 문서, 설계는 이것에 기대지 않는다 |
| project 서버 승인 뒤 reverse-engineer가 두 도구에서 REA 도구를 받는다 | implementation_assumption | A1, S19, CD-2 |
| npx 첫 실행이 30초, 도구 호출이 60초 안에 끝난다 | implementation_assumption | A2, S18, S19 |
| 대상 유래 문자열이 확인 gate를 지난 명세에서 executor를 오도하지 않는다 | implementation_assumption | Trust Model, S19 |

## Scope Expansion Rule

- 구현자가 요구사항보다 넓은 제약을 넣으면 scope expansion으로 표시하고 fan-out 전에 이 표에 probe 행을 더해 기존 런타임에서 확인한다. 요구사항으로 올리지 않는다. 예: `tool_timeout_sec`, Codex 에이전트별 sandbox, `enableAllProjectMcpServers`, REA 도구 사전 승인이나 이름 나열, runtime capture 도구 목록 고정, Windows 실행 형식, OpenCode `task` permission 제거, JSON 원문 편집기, 기존 OMP rung이나 operator pin의 blocker 완화, Node 범위 변경, 공유 MCP 목록.

## Gate Applicability

- `{SPEC_DIR}/gate-applicability.json`은 아직 없다. 구현 handoff는 `auto spec gates`가 쓴 값만 쓴다. 사용자 MCP 파일과 autopus.yaml 쓰기, subprocess(`node`), 공급망 pin, 대상 실행 동의를 다루므로 classifier가 `security_or_data`로 볼 것으로 예상한다. 예상일 뿐 판정이 아니다.
- security, validation, data_loss, deterministic_oracle gate는 `not_applicable`이 될 수 없다. TUI wizard 경로가 있어 accessibility와 ux_verification은 classifier가 정한다.
