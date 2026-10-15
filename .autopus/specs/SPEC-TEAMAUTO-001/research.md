# SPEC-TEAMAUTO-001 리서치

## 기존 코드 분석

모두 origin/main `33b5bb55` worktree에서 Read·grep으로 확인했다(2026-10-15).

- `templates/claude/commands/auto-workflows.md.tmpl`: `## go` Step 0(997–1044)이 `--team`·`--workflow`·`--solo`를 읽고 `TOPOLOGY_FLAG_COUNT > 1`이면 `[TOPOLOGY ERROR] …`(1026)로 멈추고 `Mode:`를 보인다(1040). Route A(1163–1190)는 "Delegation is risk-based, not size-based"와 독립 slice 규칙, Route B(1192–1246)는 preflight 4개(Claude Code 2.1.246+, `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`, 대화형, schema)와 "If preflight fails, report the blocker and suggest Route A", named `builder-1`·`tester`·`guardian`, shutdown·ack, Route C(1248), Workflow(1253–1293). 생성은 `pkg/adapter/claude/claude_workflow_skills.go`가 `.claude/skills/auto-go/SKILL.md`로 한다. 이 worktree에는 생성 `.claude/skills/auto-go/`가 없다.
- `content/skills/agent-teams.md`(129줄, platform claude 전용): Preflight(23–36, "Settings hooks do not inject it"), Failure Policy(115–123), Separation from Workflows(125–129).
- `content/skills/references/agent-pipeline/phases.md`: Phase 1 배정 표 `Unit | Execution | Owned paths | Forbidden paths | Depends on`(53–78), Phase 2(152–175, fan-out 중 worker는 project 전체 build를 돌리지 않는다). 공유 content라 모든 플랫폼이 쓴다.
- `internal/cli/workflow_triage.go`(105줄): `singleTriageFlag`, `newWorkflowTriageCmd`, `readWorkflowTriageFacts`(Lstat, regular file, `os.SameFile`, 1 MiB). 보고서는 `caller_declared`, `not_independently_verified`, `execution_performed: false`다.
- `pkg/taskroute`: `decide.go`(195줄) `Decide`는 parallel을 `Route == "planned"`일 때만 낸다. planned는 `large_file_scope`(path 6개 이상)·`large_change_estimate` 같은 크기 신호로도 켜진다. `ownershipOverlap`(같은 path 또는 `/` prefix), `workerScopeContained`. `decode.go` `Decode`·`uniqueValue`(중복 key, 깊이 16, trailing), `DisallowUnknownFields`. `validate.go` `cleanPath`(빈 값, `*?[]\:`, 절대 경로, `..`, `.` 거부).
- `pkg/config`: `schema_workflow.go` `WorkflowConf{CoverageThreshold}`, `schema.go:92` `Workflow`. `loader_strict.go:37-41`은 `workflow.team_default`를 `removedConfigKeys`로, `:46-52`는 `future_extension`·`operator_extension`을 reserved로 두고 나머지 모르는 key는 `KnownFields(true)`로 거부한다(`:98-126`). `loader.go:16-19` `Load`는 `loadConfig(dir, true)`라 platform 이름을 정규화하면 `os.WriteFile`로 다시 쓰고(`:90-95`), `:22-25` `LoadPreview`는 `loadConfig(dir, false)`로 쓰지 않는다. 파일이 없으면 둘 다 `DefaultFullConfig`다(`:63`).
- `internal/cli/root.go:203-211` `runRootCommand`가 `Error: %s\n`을 쓰고 `exitCodeForError`(`:173-179`)의 일반 값은 1이다. `content/skills/references/agent-pipeline/delegation.md:111`은 "Literal roots and trailing `/**` scopes are accepted"다.
- `pkg/processprobe/output.go:55` `OutputLimited`(기존 `doctor_agents_inventory.go:31`이 `--version` probe에 쓴다). `internal/cli/orchestra_invoker.go:21`이 이미 `os.Getenv("CLAUDECODE")`를 읽는다.
- `pkg/adapter/claude/claude_topology_flags_test.go:17` `TestAutoGoTopologyFlags_FailClosedBeforeRouteSelection`이 생성 auto-go skill을 읽는다.
- 이전 SPEC: SPEC-HARNESS-WORKFLOW-TEAM-001은 `implemented`(그 뒤 cutover로 `--team` → Workflow와 `workflow.team_default`가 빠졌다), SPEC-TEAMPANE-001은 `completed`. 둘 다 이 SPEC과 충돌하지 않는다.
- SPEC-REA-001: `git diff --stat origin/main...feat/rea-001 -- templates content pkg/adapter/claude`가 2026-10-15에 아무것도 출력하지 않았다. 계획상 router 행만 바꾼다.

## Probe 결과 (`evidence/probe-teamauto-capability.md`)

- A1(PASS): Bash tool child에서 `CLAUDECODE=1`, `CLAUDE_CODE_ENTRYPOINT=cli`, `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`, `CLAUDE_CODE_SESSION_ATTENDED=1`, `CLAUDE_CODE_CHILD_SESSION=1`(subagent 안). EXECPATH `2.1.288 (Claude Code)`, PATH `claude` `2.1.299`. TTY 아님.
- A3(PASS): v0.50.125는 `workflow.auto_team`을 거부하고(update 종료 1), key 없는 파일은 그대로 읽고, `--topology`를 모른다(`unknown flag`). strict decode(`KnownFields(true)`)는 v0.50.113부터 있다(그 commit을 담은 첫 tag).
- F5(review, 2026-10-15): tag v0.50.112를 offline 빌드해 `auto_team: false`를 둔 project에서 `auto update --yes`를 돌렸다. 종료 0이고 그 줄이 지워졌다.
- Y(review round 1): yaml.v3 v3.0.1 `*bool`은 `yes`·`"off"`·`y`·`no`를 bool로, `null`·`~`·빈 값을 nil로 오류 없이 받고 `"false"`·`1`만 거부한다. R: root 오류는 `Error: <msg>\n`, 종료 1이다.

## 생성 surface 영향 (T9, 2026-10-15)

- T7·T8 뒤 `go test ./pkg/adapter/claude/... ./pkg/content/... -count=1`이 바꾸지 않은 test 그대로 통과했다. 깨진 Claude golden은 없고 갱신한 golden도 없다(목록: 없음).
- `make generate-templates`는 `templates/**`를 바꾸지 않았다. `agent-teams`는 codex·gemini mirror에서 빠지고 `templates/claude/commands/auto-workflows.md.tmpl`은 생성 대상이 아니기 때문이다.
- 기준 생성 결과의 SHA-256(auto-go `fd9c9ec2…`, agent-teams `124a1fdf…`, Step 0·Route B·Route C·Workflow 절)은 `pkg/adapter/claude/claude_auto_team_surface_test.go`가 oracle로 고정한다. 새 절을 뺀 branch 생성 결과가 기준과 bytes가 같다.

## 설계 결정

- D1 판단 위치: `pkg/taskroute`의 새 순수 함수 + `auto workflow triage --topology`. 기존 `Decide`의 parallel을 그대로 쓰지 않는다. 그 gate는 planned route(파일 수·변경 줄 수로도 켜짐)에 묶여 있어 "File count … never decide this"(Route A 문구)와 어긋난다. prose만으로 두면 조합을 test할 수 없다.
- D2 감지: env(CLAUDECODE, AGENT_TEAMS, ENTRYPOINT + ATTENDED, CHILD_SESSION) + EXECPATH `--version`. PATH `claude`는 세션 버전이 아니고(A1), TTY는 Bash tool 안에서 늘 거짓이다. 대화형은 ENTRYPOINT와 ATTENDED를 둘 다 요구한다. 거짓 음성은 현재 동작(subagent)이라 안전하고, 거짓 양성(headless에서 팀)은 위험해서 보수적으로 고른다. schema 로드는 세션만 알 수 있어 세션 쪽 fallback(REQ-TA-07)이다.
- D3 설정: `workflow.auto_team`(`*bool`, `omitempty`, 없음 = on). 버린 대안: (a) `future_extension.auto_team`은 옛 typed Save가 지워(SPEC-REA-001 probe V2) 끈 설정이 말없이 사라진다. (b) `workflow.team_default` 재사용은 옛 의미(`--team` → Workflow)가 있고 `pkg/config/workflow_cutover_test.go`의 무시 계약을 깨며 옛 Save가 지운다. (c) env 변수만은 오타가 말없이 무시돼 strict loader 원칙과 맞지 않고 사용자가 고른 "설정"도 아니다. 남는 비용: key를 쓴 사용자(값이 `true`여도)가 auto 0.50.113–0.50.125를 쓰면 load가 막히고(A3), 0.50.112 이하를 쓰면 key가 무시되고 그 `auto update`가 key를 지워 끈 설정이 사라진다(F5). 바이너리가 key를 쓰지 않으니 key를 쓰지 않은 프로젝트에는 비용이 없다. `*bool`만으로는 bool이 아닌 값을 막지 못해(Y) raw node `!!bool` 검사를 `decodeStrict`에 둔다.
- D4 규칙 순서: 명시 flag → 설정 → slice → 능력. 설정과 plan이 팀을 막으면 실험 env를 권하지 않고, 유일한 실행인 버전 probe는 마지막이다.
- D5 범위: Phase 2의 준비된 wave 하나. 다른 phase는 의존 관계라 subagent 그대로라는 사용자 결정이다. 남은 unit은 Route A다.
- D6 비 Claude 플랫폼: 공유 content를 건드리지 않아 bytes가 같다(가정 AS-1).
- D7 실패 줄: CLI가 실패하면 문구를 만들 수 없어 생성 skill의 `sh` snippet이 고정 영어 한 줄을 직접 출력한다. 그래서 그 분기를 snippet 실행으로 판정한다(S14). 기존 `[workflow] fallback-class=…` 줄도 영어다.
- D8 읽기 경로(round 1): `config.Load`는 쓰므로 `LoadPreview` + `Lstat`. 관측과 판단의 경계는 snapshot이다. CLI가 관측하고, 순수 `NeedsVersionProbe`가 probe 필요 여부를, 순수 `DecideTeam`이 결과를 정한다.
- D9 경로 모양(round 1): Phase 1 표와 delegation 계약은 `pkg/foo/**`를 쓴다. 끝 `/**` 하나만 떼고 나머지 glob은 거부한다.
- D10 구현 뒤 보안 점검과 리뷰(2026-10-15): 계약보다 엄격해진 네 가지를 계약에 올렸다. (a) 겹침 비교는 NFC + Unicode case folding(`cases.Fold`) 값이다. 대소문자·정규화만 다른 표기는 기본 macOS·Windows 파일 시스템에서 같은 파일이다. (b) facts key는 `^[a-z_]+$`다. encoding/json의 대소문자 무시 매칭과 뒤 key 우선 규칙이 겹침을 숨길 수 있었다. (c) owned path의 shell 특수 문자를 거부하고 facts는 한 Bash 호출의 quoted heredoc으로 쓴다. Bash 호출끼리는 변수를 나누지 않는다. (d) Windows는 drive 경로의 `claude.exe`만 probe한다. 그리고 잘못된 facts를 오류(종료 1) 대신 `invalid_facts` decision(종료 0)으로 바꿨다. snippet이 stderr를 버려 원인이 보이지 않았기 때문이다. snippet bytes와 S13·S14 실행 방식은 그대로다.

## Plan Intent Ledger

direct `/auto plan`의 pre-plan 대화에서 온 ledger다(BS 파일 없음). 셀은 근거로만 요약했다.

| Field | Status | Source | Confidence | Decision / Assumption | If Wrong | Plan Handoff | SPEC 반영 |
|---|---|---|---|---|---|---|---|
| goal | answered | user | high | 세션이 native team을 실행할 수 있고 겹치지 않는 owned paths의 준비된 slice가 2개 이상일 때만 `/auto go`가 그 phase를 Agent Team으로 | — | requirements | REQ-TA-01–06 |
| scope_boundary | answered | user | high | 의존 단계는 subagent. `/auto go` routing만. plan·review·비 Claude 플랫폼은 이유 줄 말고는 변경 없음. Route B 계약 밖 새 기능 없음 | creep | non-goals | Outcome Boundary non-goals, AS-1 |
| constraints | answered | user | high | 전환·유지 모두 이유 한 줄. `--solo`나 설정으로 끔. 명시 flag 의미 유지, 상호 배타. 새 secret·승인 없음. 300줄, `make generate-templates` | — | constraints | REQ-TA-01, 09, 10, 12, T9 |
| done_evidence | assumed | planner | medium | 판단 함수 결정적 test, 모든 플랫폼 생성 문구, 이유 문자열, 실제 터미널 전환·유지 run 각 1회(못 하면 CD) | weaker | acceptance + CD | S1–S15, CD-1, CD-2 |
| brownfield_impact | assumed | planner | medium | 조건이 맞으면 기본 on. 팀이 가능한 Claude Code 세션만 영향, 나머지는 지금과 같음 | surprise cost | migration notes | REQ-TA-09, 14, Reviewer Brief |

## Question Audit

- question_transport: AskUserQuestion (pre-plan conversation) · question_count: 0 (사용자가 제안문을 그대로 승인) · unresolved_fields: done_evidence, brownfield_impact (assumed).

## Outcome Lock

- User-visible outcome: Claude Code 대화형 최상위 세션에 팀 기능이 켜져 있고 plan에 준비된 disjoint slice가 2개 이상이면 `/auto go`(flag 없음)가 Phase 2를 native Agent Team으로 돌리고, 그 밖에는 subagent 그대로다. 어느 쪽이든 이유 한 줄이 나온다. `workflow.auto_team: false`나 `--solo`로 끈다.
- Mandatory requirements: REQ-TA-01–13, 15 (Must). REQ-TA-14는 Should.
- Explicit non-goals: `spec.md` `## Outcome Boundary`.
- Completion evidence: Must S1–S15 PASS, coverage 85% 이상, 300줄, `GOOS=windows GOARCH=amd64 go build ./...`, `make generate-templates` drift 없음, `evidence/manual-smoke-teamauto.txt`(CD-1–3), `auto spec validate .autopus/specs/SPEC-TEAMAUTO-001 --strict`.

## Visual Planning Brief

판단 흐름은 `spec.md` `## Visual Planning Brief`(flowchart), Phase 2 command flow는 `plan.md` `## Visual Planning Brief`(sequence)다. UI 화면은 없다. 사용자가 보는 것은 이유 한 줄이다: `Agent Team: 전환 — 준비된 slice 2개의 owned_paths가 겹치지 않아 native Agent Team으로 실행한다` 또는 `Agent Team: 유지 — …`.

## Minimality Decision Matrix

| Ladder step | Evidence | Decision | Receipt item |
|-------------|----------|----------|--------------|
| actual need | ledger goal·constraints: 조건부 자동 팀과 이유 한 줄, 끄는 설정 | proceed | 판단, 한 줄, key |
| existing code/helper/pattern | `auto workflow triage`(파일 guard, 형식 flag), `uniqueValue`, `cleanPath`, `ownershipOverlap`, Route B lifecycle 문구, `processprobe.OutputLimited`, `config.Load` | reuse | 새 subcommand 없이 mode flag, Route B 재정의 없음 |
| stdlib/native | `os.LookupEnv`, `os/exec`, `context.WithTimeout`, `regexp`, `strconv` | use | env 읽기, 버전 probe·해석 |
| existing dependency | cobra, yaml.v3 | reuse | flag, key decode |
| new dependency or new abstraction | 의존성 0. 새 abstraction: `DecideTeam`과 facts·decision 타입. 기존 `Decide`는 크기 신호에 묶여 쓸 수 없다(D1) | accepted | REQ-TA-04, 05, 08 |
| minimum sufficient verification | 순수 함수 표 test, 실제 `NewRootCmd`, fake EXECPATH script, 생성 bytes·SHA-256 비교, v0.50.125 대조, 실제 터미널 2 run | required checks | security(EXECPATH guard)·data-loss(설정 왕복)·deterministic-oracle gate 유지 |

## Semantic Invariant Inventory

| ID | source clause | invariant type | affected outputs | acceptance IDs |
|----|---------------|----------------|------------------|----------------|
| INV-TA-01 | "Print ONE reason line both when switching and when not switching" | ordering (first-match 규칙과 출력 하나) | decision `route`·`reason`·`line`, snippet stdout, Phase 2 진입 transcript | S1, S2, S3, S4, S12, S14, S15 |
| INV-TA-02 | "≥2 ready slices that can run concurrently with disjoint owned paths" | paired matching (unit 쌍마다 path 겹침), grouping (ready 집합, follow-up 집합) | `ready_units`, `team_units`, `followup_units`, 겹침 쌍 `{a}, {b}` | S2, S3, S4, S4b, S15 |
| INV-TA-03 | "the session can run native teams" (Route B preflight: 2.1.246+, env, 대화형) | numeric formula (정수 버전 비교), cross-entity comparison (EXECPATH ≠ PATH) | reason `claude_version_*`, `not_*`, probe 실행 여부 | S4, S5, S15 |
| INV-TA-04 | "Users can disable via … a config setting" | parser (strict bool, 없음 = on, 쓰지 않음) | `autopus.yaml` bytes, `config_off`, 옛 바이너리 종료 코드 | S8 |
| INV-TA-05 | "Explicit --team / --workflow / --solo keep current semantics"; 비 Claude 플랫폼 변경 없음 | deduplication (생성 bytes 동일성) | 생성 skill 절 bytes, 네 플랫폼 SHA-256 | S6, S7, S9 |
| INV-TA-06 | 이유 줄 ko/en 문자열 | parser/report row (언어별 정확한 문자열) | `line`, `tools_fallback_line` | S10 |
| INV-TA-07 | "Sequential (dependent) phases stay subagent" | grouping (follow-up = team_units 밖이고 done 아님), ordering (shutdown·ack → follow-up → Gate 2) | `followup_units`, transcript 순서 | S4b, S15 |

## Feature Coverage Map

| Outcome slice | Covered by | Status |
|---------------|------------|--------|
| 판단 함수와 규칙 순서 | REQ-TA-04, 05, 10, T1 | covered |
| 능력 관측(env, EXECPATH) | REQ-TA-06, T4 | covered (A2는 CD-2) |
| CLI mode와 facts 검증 | REQ-TA-08, T2, T5 | covered |
| 설정 key와 옛 바이너리 | REQ-TA-09, T6 | covered |
| 생성 문구(auto-go, agent-teams)와 한 줄 규칙 | REQ-TA-01–03, 07, 13, T7, T8 | covered |
| 비 Claude 무변경 | REQ-TA-11, T9 | covered |
| 팀 wave 뒤 남은 unit 인계 | REQ-TA-02, T7, T8 | covered (실행 증거는 CD-1) |
| CLI 실패 snippet | REQ-TA-13, T7, T9 | covered (S14) |
| 실제 터미널 전환·유지·fallback·resume | REQ-TA-15, T11 | completion-debt (CD-1, CD-2, CD-3) |

## Completion Debt

| Item | Blocks | Required resolution |
|------|--------|---------------------|
| CD-1 실제 터미널 자동 전환 run(남은 unit 포함) | REQ-TA-02, 15 | T11이 S15(a)를 기록 |
| CD-2 실제 터미널 유지 run과 A2 환경 값(npm 포함) | REQ-TA-03, 06, 15 | T11이 S15(b)(e)를 기록, A2를 PASS/FAIL로 |
| CD-3 실제 터미널 fallback·resume | REQ-TA-01, 07, 15 | T11이 S15(c)(d)를 기록 |

## Evolution Ideas

These are optional improvements and do not block sync completion.

| Idea | Why not required now | Promotion trigger |
|------|----------------------|-------------------|
| 두 번째 이후 wave 재판단 | 첫 wave로 사용자 조건이 닫힌다 | 사용자가 wave별 팀을 요청 |
| 겹칠 때 disjoint 부분집합만 팀으로 | 재분할은 planner 몫이고 보수적 유지가 현재 동작이다 | 겹침으로 팀을 놓친 사례 반복 |
| run 단위 `--no-team` | topology flag를 늘리지 않기로 했다 | 사용자가 명시 요청 |
| 비 Claude 플랫폼의 이유 줄 | AS-1. 그 플랫폼에는 native team이 없다 | 사용자가 그 줄을 원한다고 확인 |

## Sibling SPEC Decision

| Decision | Reason | Sibling SPEC IDs |
|----------|--------|------------------|
| none | 한 결과(자동 팀 전환)를 판단·CLI·설정·문구가 함께 닫는다. task 11개, 파일 약 19개로 예외 기준(25개 초과 task이면서 40개 초과 파일) 미만 | None |

## Reference Discipline

| Reference | Type | Verification |
|-----------|------|--------------|
| 위 `templates/claude/…`, `content/skills/…`, `internal/cli/workflow_triage.go`, `pkg/taskroute/*`, `pkg/config/*`, `pkg/processprobe/output.go`, `pkg/adapter/claude/*` 경로·줄·symbol | existing | Read, grep(2026-10-15, `33b5bb55`) |
| `auto 0.50.125` 동작, Claude Code 2.1.288 env·`--version`·`--disallowedTools`, yaml.v3 v3.0.1 decode, SPEC-REA-001 probe | existing (binary, runtime, library, history) | 실행(evidence C·F·Y·R), `claude --help`, `git show feat/rea-001:…` |
| `.claude/skills/auto-go/SKILL.md`, `.claude/skills/agent-teams/SKILL.md` | generated surface (원천은 위 template·content) | `claude_workflow_skills.go`, topology test |
| `TeamInput`, `NeedsVersionProbe`, `DecideTeam`, `firstOverlapPair`, `DecodeTeamFacts`, `TeamDecision`, `team*.go`, `readWorkflowTriageFile`, `workflow_triage_topology.go`, `workflow_triage_capability.go`, `WorkflowConf.AutoTeam`, `AutoTeamEnabled`, `schema_workflow_auto_team.go`, `--topology`, `## Automatic Entry from Route A`, `#### Route A Phase 2 Entry: Agent Team Auto-Selection`, `InvalidFactsDecision`, `invalid_facts`, `facts_error`, `claudeVersionProbePathAllowed`, 새 test 파일, `evidence/manual-smoke-teamauto.txt` | [NEW] planned addition | 부재 확인(grep) |

## Reviewer Brief

- Intended scope: Phase 2 진입의 자동 팀 판단(함수, CLI mode, 능력 관측), `workflow.auto_team`, auto-go·agent-teams 문구, 비 Claude·명시 flag 무변경 증명.
- Explicit non-goals: 다른 phase·plan·review·`auto pipeline run` 변경, 새 topology flag, Route B 밖 팀 기능, 비 Claude 문구, Workflow 결합.
- Self-verified: A1·A3 probe, Route A/B 문구 줄 번호, 재사용 helper 존재, Traceability Matrix, Semantic Invariant Inventory ↔ acceptance, [NEW] 표시.
- Reviewer should focus on: 규칙 순서와 거짓 양성(headless에서 팀) 가능성, EXECPATH 실행 guard, 한 줄 규칙의 세 경로와 snippet, 남은 unit 인계 순서, `!!bool` 검사, `LoadPreview` 읽기 경로, 그리고 Completion Debt only. round 1 처리는 `evidence/review-rounds.md`다.

## Open Risks

- AS-1(가정): ledger "beyond printing the reason line"을 Claude 표면의 줄로 읽었다. 비 Claude 플랫폼에도 줄을 원한다면 네 플랫폼 go 문구에 고정 한 줄을 더해야 하고 S9 oracle이 바뀐다.
- R1: A2 not-run. print mode가 (`cli`, `1`)을 보고하면 headless run이 팀을 고를 수 있다. T11 전에는 릴리스하지 않는다.
- R2: brownfield_impact assumed. 팀 환경을 켠 사용자는 기본으로 팀 run(토큰 증가)을 보게 된다. 한 줄과 문서(REQ-TA-14), 설정으로 끈다.
- R3: VS Code 등 다른 entrypoint는 `not_interactive`로 subagent다(거짓 음성, 안전한 쪽).
- R4: 실험 기능이라 Claude Code가 env 이름이나 팀 동작을 바꿀 수 있다. 바뀌면 판단은 subagent로 떨어진다.
- R5: S15(c)는 `--disallowedTools SendMessage`가 schema 로드를 실패시킨다는 가정이다. 아니면 CD-3은 다른 방법을 찾을 때까지 남는다.

## Self-Verify Summary

- Q-CORR-01 | status: PASS | attempt: 2 | files: spec.md, research.md, plan.md | reason: round 1 F-001·F-010으로 `config.Load`를 `LoadPreview` + `Lstat`로 고쳤고 loader 줄 번호를 다시 읽었다
- Q-CORR-02 | status: PASS | attempt: 1 | files: spec.md, plan.md, research.md | reason: 새 타입·함수·파일·절은 모두 [NEW]다
- Q-CORR-03 | status: PASS | attempt: 2 | files: spec.md, acceptance.md, evidence | reason: yaml.v3 decode와 root 오류 출력을 실행해 관측 문자열(evidence Y·R)로 oracle을 고쳤다
- Q-CORR-04 | status: PASS | attempt: 1 | files: research.md | reason: Reference Discipline이 existing, binary·runtime·library·history, generated surface, [NEW]를 나눈다
- Q-COMP-01 | status: PASS | attempt: 1 | files: 5개 문서 | reason: 계약, 계획, oracle, 근거, 요약이 나뉜다
- Q-COMP-02 | status: PASS | attempt: 2 | files: spec.md, plan.md, acceptance.md | reason: 남은 unit 인계를 규칙 5와 S13·S15에, fallback·resume을 S14·S15(c)(d)에 넣었다
- Q-COMP-03 | status: PASS | attempt: 1 | files: spec.md | reason: REQ마다 조건과 관측 대상이 있다
- Q-COMP-04 | status: PASS | attempt: 2 | files: spec.md, research.md | reason: ledger의 migration notes를 `## Migration Notes`로 옮겼고 실패 경로 검증 공백을 S14·CD-3으로 닫았다
- Q-COMP-05 | status: PASS | attempt: 3 | files: research.md, acceptance.md | reason: 실행 invariant를 실행 oracle(S14 snippet, S15 transcript)에 묶고 S12 무변경 대상을 `<tmp>/proj`로 좁혔다
- Q-COMP-06 | status: PASS | attempt: 1 | files: spec.md, research.md | reason: Traceability Matrix가 REQ 15개를, Reviewer Brief가 범위를 묶는다
- Q-COMP-07 | status: PASS | attempt: 2 | files: research.md, spec.md | reason: 실제 터미널 run을 CD-1–3으로 남겼다
- Q-COMP-08 | status: PASS | attempt: 1 | files: plan.md | reason: A1·A3 PASS는 실행 증거, A2는 이유 있는 not-run, 진술은 세 class다
- Q-FEAS-01 | status: PASS | attempt: 2 | files: spec.md, plan.md | reason: 쓰지 않는 load로 바꾸고 관측(CLI)·판단(순수 함수 둘) 경계를 snapshot으로 맞췄다
- Q-FEAS-02 | status: PASS | attempt: 1 | files: spec.md | reason: 생성 원천 둘과 생성 surface를 나눴다
- Q-FEAS-03 | status: PASS | attempt: 2 | files: acceptance.md, spec.md | reason: S12 fixture를 나눠 충족 가능하게 하고 stderr를 바이너리 기준 bytes로, validate 명령에 SPEC 경로를 넣었다
- Q-STYLE-01 | status: PASS | attempt: 1 | files: spec.md | reason: REQ는 SHALL 단정문이다
- Q-STYLE-02 | status: PASS | attempt: 1 | files: spec.md, acceptance.md | reason: Priority와 EARS type이 별도 축이다
- Q-STYLE-03 | status: PASS | attempt: 1 | files: acceptance.md | reason: bare Given/When/Then/And다
- Q-SEC-01 | status: PASS | attempt: 1 | files: spec.md | reason: facts·env 입력 검증과 정규화한 값만 출력한다
- Q-SEC-02 | status: PASS | attempt: 1 | files: spec.md | reason: EXECPATH guard, PATH 미사용, 비밀 env 미사용
- Q-SEC-03 | status: PASS | attempt: 1 | files: spec.md, acceptance.md | reason: 영구 artifact는 경로를 치환한 evidence뿐이다
- Q-COH-01 | status: PASS | attempt: 1 | files: spec.md | reason: 한 변경 이야기로 모인다
- Q-COH-02 | status: PASS | attempt: 1 | files: research.md | reason: 필수 확인은 debt, 선택 개선은 Evolution Ideas다
- Q-COH-03 | status: PASS | attempt: 1 | files: research.md | reason: Sibling 없음과 근거가 있다
