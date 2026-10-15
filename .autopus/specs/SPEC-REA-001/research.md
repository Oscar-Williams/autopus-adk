# SPEC-REA-001 리서치

branch `feat/rea-001`, 기준 origin/main `4d73aa91`(PR #197 merge 뒤)에서 2026-10-15에 읽었다. 근거는 worktree Read·`rg`, `git show <tag>`, auto 0.50.123 실행과 문서 확인(`evidence/probe-rea-forward-compat.md`), npm registry다. 확인하지 못한 것은 implementation_assumption이나 not-run probe로 적었다. 인용한 요청·ledger 문구는 비신뢰 입력 증거로만 요약했다.

## 기존 코드 분석
- MCP: Claude 템플릿 `templates/claude/mcp.json.tmpl:1-15`(`/bin/bash -l -c "exec npx …"`, `$comment`), merge `pkg/adapter/claude/claude_files.go:25-87`(managed 서버를 지우고 템플릿 서버를 다시 넣는다), 판정 `claude_clean.go:209-229`(이름 + package 부분 문자열), clean `:176-207`. Codex `templates/codex/config.toml.tmpl:30-32`, merge `codex_config_structural.go:48-127`(key 단위 관리 `:10-27`, header는 남김), clean `codex_clean_merge.go:180-213`(`isAutopusOnlyCodexSection`만 header 삭제), 호출 `codex_settings.go:24-52`. Antigravity `templates/gemini/settings/settings.json.tmpl:1-12`, merge `antigravity_settings.go:187-198`(덮기만 함), clean `antigravity_lifecycle_clean.go:92-101`(현재 렌더와 DeepEqual인 서버만). OMP·OpenCode는 MCP를 쓰지 않는다. 사전 승인 MCP는 `mcp__sequential-thinking__sequentialthinking` 하나(`pkg/content/hooks.go:196`)이고 `enableAllProjectMcpServers`는 없다(`rg`).
- 설정: `FeaturesConf{CC21}`(`pkg/config/schema.go:97-100`), `RuntimeConf`(`runtime.go:13-15`, yaml `-`), strict decode `loader_strict.go:98-125`, 허용 key는 `removedConfigKeys`·`reservedConfigKeys`(`future_extension`, `operator_extension`). typed `Save`(`loader.go:130-153`)는 struct를 다시 marshal한다. Node 버전 검사와 semver 의존성은 없다.
- CLI: wizard `internal/cli/tui/wizard_steps.go`(268줄, 결과 L28-37, 단계 L87-132, Confirm 모델 L194-210), init `init.go:101-137`·flag `:241-246`(258줄), update `update.go`(292줄, flag L276-290, 언어 질문 L185-188 → `prompts.go:66-93`), workspace `update_workspace.go:28-29`(300줄, repo마다 `runNestedUpdateCommand`), `--auto`는 `global_flags.go:17,32`, TTY `prompts.go:18`·seam `quality_compose_driver_gate.go:21`, doctor 조립 `doctor_json.go:118-137`·`doctor_text.go:15`, CLI `TestMain` `pipeline_backend_omp_rpc_test.go:36`.
- 에이전트·프리셋: 정규 16개 `quality_tier.go:17-21`(알파벳 순서), balanced rung `role_model_policy_balanced.go:20-68`(3 rung, "every canonical agent exactly once"), tier `native_balanced.go:33-38`, 표준 판정 `:10-31`(map에 없는 에이전트는 표준 배치), ultra `defaults.go:136-147`, registry `quality_preset_history.go:52-104`, 분류 `:109-133`(정확한 map 일치), capability `role_model_policy_matrix.go:38-55`, opus 사다리 Claude `high`·ultra `max`(`quality_effective.go:53-70`), Codex `gpt-6.1-sol`·`xhigh`(`codex_profile.go:118-135,182-191`). native 접힘 `omp_native_agents.go:11-25`, `OMPAutopusAgentNames` `omp_agent_source.go:50-52`, 그룹 `internal/cli/quality_compose_groups.go:24-37`. 16이 박힌 문장은 `templates/shared/omp-agent-pipeline.md.tmpl:69-72`, `platform_omp_agent_migration.go:27,40`, `init_help.go:20`, `update_help.go:31`, `quality_help.go:30`이고 16·17 수 고정 test는 약 20건이다(`rg`).
- 도구 변환: Claude는 model·effort만 고친다(`claude_agent_content.go:67-104`). Gemini 허용 목록 `agent_transformer.go:159-178`은 모르는 도구를 버린다. OMP `agent_transformer_omp_render.go:135-155`는 `mcp__` 이름을 그대로 둔다. OpenCode `agent_transformer_opencode.go:60-80`은 모르는 도구를 permission key로 넣고 question·skill·task를 늘 더한다. Codex agent TOML에는 도구 칸이 없다(`agent_transformer.go:113`). MCP 참조 변환 `skill_transformer_replace_mcp.go:22,50`, 호출 `skill_transformer_replace.go:124`. `FilterPlatformReferences`(`skill_transformer_filter.go:19`)는 production 호출자가 없다.
- 라우트: Claude `claude_workflow_skills.go:22-42`(customBody), router 표 `templates/claude/commands/auto-router.md.tmpl:55-75`(템플릿 6174 bytes), frozen 목록 `pkg/adapter/claude/router_budget_test.go:16-19`(8192 bytes 한도, antigravity 같은 test), Codex `codex_workflow_specs.go:14-17`, `codex_workflow_custom.go:26`, OMP `omp_specs.go:25`, OpenCode `opencode_specs.go:28`, `opencode_workflow_custom.go:29`, Gemini `templates/gemini/skills/auto-*/SKILL.md.tmpl`.

## Outcome Lock
- User-visible outcome: Node 조건을 충족한 새 `auto init` 프로젝트에서 Claude Code·Codex·Antigravity가 프로젝트 범위 REA MCP(`npx -y rea-agents@6.5.0 mcp`)를 띄운다. `/auto reverse`로 reverse-engineer가 evidence_id가 붙은 행동 명세와 unknowns를 내고, 사용자가 확인한 명세만으로 executor가 구현하고 tester가 검증한다. 기존 프로젝트는 대화형 update에서 한 번만 묻고 비대화형 실행은 REA를 더하지 않는다. 사용자 범위 파일은 건드리지 않는다.
- Mandatory requirements: REQ-REA-01–17, 20 (Must). REQ-REA-18, 19는 Should.
- Explicit non-goals: `spec.md` `## Outcome Boundary`.
- Completion evidence: Must S1–S19 PASS, `auto eval harness run`, Windows cross-build, 300줄, `evidence/manual-smoke-rea.txt`(S18, S19), `auto spec validate --strict`.

## Plan Intent Ledger
| Field | Status | Source | Confidence | Decision / Assumption | If Wrong | SPEC 반영 |
|---|---|---|---|---|---|---|
| goal | answered | user | high | REA MCP 등록, 전용 `reverse-engineer` 에이전트, `/auto reverse`로 분석·증거·재구현 | — | REQ-REA-03, 10, 13, 14 |
| scope_boundary | answered | user | high | MCP는 Claude Code·Codex·Antigravity만. OMP·OpenCode는 `rea setup` 안내만. #197 뒤 별도 PR, v0.50.126 | OMP·OpenCode writer로 범위 확대 | non-goal, REQ-REA-09 |
| constraints | answered | user + repo | high | 새 init 기본 on(wizard 기본 Yes, `--rea`/`--no-rea`, features flag), `rea-agents@6.5.0` 정확한 pin, Node 범위 gate, 새 secret·승인 없음, 300줄, 생성 root는 generate로만 | — | REQ-REA-01–05, Q1 결정 |
| done_evidence | assumed | planner | medium | 표면마다 결정적 test, 실제 터미널 smoke, 동의 뒤 live 실행 | 증거 약화 | S1–S19, CD-1, CD-2 |
| brownfield_impact | answered | user | high | 기존 프로젝트는 대화형 update에서 한 번 묻고 저장, 비대화형은 추가 없이 hint. 17번째 에이전트와 skill은 늘 생긴다. 옛 auto가 새 파일을 읽어야 하고 깨지면 MODELUX-003처럼 알린다 | 업그레이드가 네트워크 MCP를 몰래 더함 | REQ-REA-06, 07, 17, Q1 |

## Question Audit
- question_transport: AskUserQuestion · question_count: 1(ledger gate의 brownfield_impact). plan 전 대화에서 결정 5건(통합 형태, PR 범위, MCP 대상 도구, 기본 on, 전용 에이전트)을 물었다.
- unresolved_fields: done_evidence(assumed → Completion Debt CD-1·CD-2).

## Open Question Decisions
| Q | 결정 | 근거 |
|---|---|---|
| Q1 | `features.rea.enabled`(*bool)를 두되 managed 항목이 결정을 담지 못할 때만 쓴다 | A3: key는 옛 auto의 모든 명령을 막고 managed 항목은 옛 update 뒤에도 남는다 |
| Q2 | balanced에 opus rung `balancedRungAnalysis`를 더한다 | matrix 불변식, 도구 간 같은 model, preset `opus` = rung tier |
| Q3 | Codex는 `startup_timeout_sec = 30`만 쓰고 `tool_timeout_sec`는 기본 60초 | Codex 문서 기본 10초·60초, A2 not-run |
| Q4 | `tools: Read, Grep, Glob, mcp__rea`, `permissionMode: default`, foreground만 | Claude 문서가 `mcp__<server>`를 받는다. runtime은 A1·CD-1 |
| Q5 | CLI가 실행당 1회 probe(POSIX login shell, Windows PATH, 5초), adapter에는 action만 | 서버 실행 형식과 같은 shell, adapter 순수 |
| Q6 | `--auto`도 확인 gate에서 멈춘다 | PRD Anti-goal "자동 승인으로 끝까지", 사용자 결정 2 |
| Q7 | 정확한 모양 + 알려진 pin 목록 `[6.5.0]`, pin만 다르면 `other_pin`(손대지 않음) | Claude merge가 managed를 지운다(R6) |
| Q8 | 공유 MCP 목록은 만들지 않는다. REA 모양만 `pkg/reamcp` 한 곳에 두고 세 adapter와 doctor가 쓴다 | 템플릿 불변으로 REQ-REA-16, 소비처 4곳 |
| Q9 | `.autopus/reverse/<slug>/behaviour-spec.md`를 메인 세션이 확인 뒤 쓴다 | 에이전트에 쓰기 도구가 없다 |
| Q10 | HB6(balanced 16, v0.50.116–v0.50.125, runtimeStandard), HU3(ultra 16, v0.50.115–v0.50.125) | `git show v0.50.115/116/125` 비교 |
| Q11 | OMP·OpenCode 에이전트 파일에 MCP 이름을 쓰지 않는다. Antigravity는 `mcp_rea_*` | OMP의 MCP 도구 이름 미확인(R11), Gemini CLI 문서 |
| Bash | reverse-engineer에서 뺀다 | REA CLI로 MCP 권한 질문을 우회하고 파일을 쓸 수 있다 |
| workspace | `--rea`·`--no-rea` + `--workspace`는 쓰기 전 거부 | nested update가 flag를 넘기지 않고 `update_workspace.go`는 300줄 |

### Q1 forward-compat 비교 (A3 실행 근거)
| 안 | v0.50.113–v0.50.125 | undecided 구분 | 대가 |
|---|---|---|---|
| A 늘 `features.rea` | 새 init 프로젝트 전부에서 모든 명령 실패(update 종료 1) | key 하나, 단순 | MODELUX-003보다 넓다(그쪽은 opt-in일 때만) |
| B `future_extension.rea` | 읽힌다 | 옛 update가 block을 지워 결정이 사라진다 | reserved 이름 남용, loader 변경 |
| D 별도 상태 파일 | 영향 없음 | 된다 | ledger의 features flag와 다르고 새 소유 파일·공유 규칙이 생긴다 |
| F(선택) 항목이 담지 못할 때만 key | 기본 경로(init·update Yes + Node 충족)에서 단절 없음. opt-out과 gate에 막힌 opt-in만 단절 | key > managed 항목 > undecided | 판정이 대상 파일을 읽는다. managed 항목을 손으로 모두 지우면 다시 한 번 묻는다 |
- ledger는 "옛 auto가 새 파일을 읽어야 한다"를 먼저 두고 깨지면 알리라고 했다. F는 기본 경로에서 앞을 지키고 나머지에서 뒤(REQ-REA-17 줄)를 따른다. Gate 1에서 사용자가 A를 고르면 쓰기 표의 on 두 행을 바꾸면 된다(`spec.md` Open Issues).

### Q2 balanced opus 비교
| 안 | matrix 불변식 | 도구 간 model | 대가 |
|---|---|---|---|
| 새 opus rung(선택) | 지킨다 | Claude Opus 5.5·Codex Sol·OMP 같은 후보 | rung 4개, OMP balanced 프로필이 opus·Sol을 새로 요구(R6) |
| top rung | 지킨다 | Fable·Astra | 사용자 결정 `opus`와 다르다. preset을 opus로 쓰면 표준 배치에서 빠진다 |
| matrix 밖 | 깬다 | OMP만 deep_reasoning route(Fable)라 도구마다 다르다 | 정규 집합 일치 test에 예외가 생긴다 |
- 16개 기존 preset은 HB6·HU3 Historical이라 customized로 오인되지 않는다(Q10). rung thinking은 손으로 opus를 고른 사다리 값과 같아 표준·custom 경로 결과가 같다.

## Pinned Version
| 항목 | 값 | source | checked_at |
|---|---|---|---|
| REA | `rea-agents@6.5.0`(MIT, `dist-tags.latest` 6.5.0, bin `rea`·`rea-agents`) | `npm view rea-agents@6.5.0`, REA README | 2026-10-15 |
| Node 범위 | `^22.19.0`, `^24.11.0`, `>=26.0.0` 중 하나 | 같은 결과의 `engines.node` | 2026-10-15 |
- greenfield가 아니라 Technology Stack Decision은 두지 않는다. rea-agents는 Go 의존성이 아니라 사용자의 AI 도구가 서버를 띄울 때 npx가 받는 runtime 의존성이다. auto 자신은 네트워크를 쓰지 않는다.

## Visual Planning Brief
흐름도(init·update·doctor command-flow, `/auto reverse` sequence와 clean-room 경계)와 wireframe은 `plan.md`이고 `wireframe intent: assumed`다.

## Minimality Decision Matrix
| Ladder step | Evidence | Decision | Receipt item |
|-------------|----------|----------|--------------|
| actual need | Outcome Lock, ledger goal·constraints, PRD P0 17개 | proceed | 등록·결정·에이전트·라우트·clean-room |
| existing code/helper/pattern | 세 adapter merge·clean, wizard Confirm, `promptLanguageSettings`, Historical registry, `customBody` route, MODELUX-003 옛 바이너리 줄 | reuse | 판단 밖 경로는 기존 함수 |
| stdlib/native | `strconv`·`strings`(버전 비교), `encoding/json`, `os/exec`(probe), `context`(5초) | use | semver 라이브러리 없음 |
| existing dependency | yaml.v3(`*bool`), huh(Confirm) | reuse | 새 Go 의존성 0 |
| new dependency or new abstraction | 새 Go 의존성 없음. rea-agents는 runtime npx 의존성. 새 abstraction은 `pkg/reamcp`(모양·분류·버전), `ReaMCPAction`, `balancedRungAnalysis` | accepted | 소비처 4곳이 같은 모양을 써야 해서 한 곳 |
| minimum sufficient verification | 실제 adapter·명령 test, seam, round trip SHA, 수동 smoke 2건 | required checks | security·data-loss·deterministic-oracle gate 유지 |

## Semantic Invariant Inventory
| ID | source clause | invariant type | affected outputs | acceptance IDs |
|----|---------------|----------------|------------------|----------------|
| INV-REA-01 | "persists the answer so it is never asked again" | state transition: key > managed 항목 > undecided, 질문은 undecided 대화형에서만 | 질문 수, autopus.yaml, MCP 파일 | S2, S3, S7 |
| INV-REA-02 | "non-TTY never adds rea and prints a one-line hint" | 비대화형 undecided: 추가·쓰기·stdin 0, hint 1 | MCP 파일, autopus.yaml, stdout | S7 |
| INV-REA-03 | "exactly one Autopus-managed rea entry, user entries untouched" | classification: 정확한 모양 + pin 목록 | 분류 값, entry 값 | S4, S8 |
| INV-REA-04 | "remove only the managed entry, no residue" | round trip: remove(add(F)) = F | 세 파일 SHA-256 | S5 |
| INV-REA-05 | "Node ^22.19.0 or ^24.11.0 or >=26.0.0" | ordering: version triple 비교 | pass·unsupported·absent·unknown | S6 |
| INV-REA-06 | "agent inventory 17 across all generated surfaces" | set equality with CanonicalAgentNames, OMP 17 + task | 에이전트 파일 이름, override key | S9, S11 |
| INV-REA-07 | "existing 16-agent presets must not be misread as customisations" | classification: 정확한 map → Historical row | class, Row, RuntimeStandard | S10 |
| INV-REA-08 | "mcp__rea__x preserved, mcp__other__y still rewritten" | parser rewrite per reference | 렌더된 줄 | S12 |
| INV-REA-09 | "executor implements from the confirmed spec only" | data-flow boundary | route detail 문장, executor 입력 | S13, S19 |
| INV-REA-10 | "MCP files unchanged while REA is not enabled" | bytes equality | 세 파일 bytes | S15 |
| INV-REA-11 | "never touch user-level files, no pre-approved permissions" | absence over write set | HOME sandbox, permissions.allow | S14, S16 |

## Feature Coverage Map
| Outcome slice | Covered by | Status |
|---------------|------------|--------|
| Happy path: 새 init → 세 파일에 rea, 17 에이전트, `/auto reverse` | REQ-REA-02, 03, 10, 11, 13 / S3, S9, S13, S18, S19 | covered |
| Error/recovery: Node 없음·구버전, flag 충돌, opt-out 제거, clean, 사용자 항목 보존, REA 미등록 안내, gate 수정·중단 | REQ-REA-04–07, 10, 13 / S3–S7, S9, S13 | covered |
| Integration boundary: 세 writer·clean, OMP·OpenCode 안내, 사용자 범위 무접촉, 5개 route registry, OMP 17 + task, preset·history, MCP 참조 변환, 옛 바이너리 | REQ-REA-03, 09, 11, 12, 13, 16, 17 / S4, S5, S10–S12, S15, S16, A3 | covered |
| CLI surface: wizard, flag, update 1회 질문, hint, doctor text·JSON, 옛 바이너리 줄, ko/en | REQ-REA-02, 06–08, 17, 18 / S3, S7, S8, S20 | covered |
| Verification: 결정적 test, 실제 터미널, 동의한 live, Windows build, eval | REQ-REA-20 / S17–S19 | covered(S18·S19는 CD-1·CD-2) |
| Docs/ops: README, CHANGELOG, help, pin, 사전 승인 없음 | REQ-REA-15, 19 / S14, S21 | covered |

## Completion Debt
| Item | Blocks | Required resolution |
|------|--------|---------------------|
| CD-1 실제 터미널: Claude Code `/mcp`의 rea connected와 도구 목록, `/agents`의 reverse-engineer, Codex·Antigravity MCP 목록의 rea(모델 호출 없음) | REQ-REA-03, 10의 런타임 증거, probe A1 | S18 PASS 기록 |
| CD-2 동의 뒤 live `/auto reverse` 1회(작은 정적 대상), executor 입력에 명세 파일만 | REQ-REA-14의 런타임 증거 | S19 PASS 기록 |

## Evolution Ideas
These are optional improvements and do not block sync completion.

| Idea | Why not required now | Promotion trigger |
|------|----------------------|-------------------|
| OMP·OpenCode 프로젝트 MCP writer와 그 에이전트의 REA 도구 이름 | 사용자 범위 경계, OMP 이름 미확인 | OMP·OpenCode 사용자 요청 |
| 모든 MCP 서버의 Windows 실행 형식 | 기존 서버도 POSIX 형식이다 | Windows 사용자 보고 |
| doctor의 REA 새 버전 안내와 pin 갱신 도우미 | 수동 pin 갱신으로 충분 | REA 새 버전 |
| 읽기 전용 REA 도구 일부 사전 승인 | 도구 감사 전 | 도구 감사 완료 |
| workspace 묶은 hint와 `--workspace --rea` 전달 | 비대화형이라 묻지 않는다 | workspace 사용자 요청 |
| Codex 에이전트별 read-only sandbox | Codex 역할 설정 지원 미확인 | probe 뒤 사용자 요청 |

## Sibling SPEC Decision
| Decision | Reason | Sibling SPEC IDs |
|----------|--------|------------------|
| none | Outcome Lock은 등록 → 분석 → 확인 → 구현의 한 흐름이다. "MCP 등록만"은 에이전트 안내가 `--rea` flag에 기대고, "16→17 파급"은 배포·migration 순서 경계가 없다. 같은 module·같은 릴리스라 나누면 순서 의존만 생긴다. task 21개, 소스 파일 test 포함 약 50개로 기준(25개 초과 task이면서 40개 초과 파일)을 동시에 넘지 않는다 | None |

## Reference Discipline
| Reference | Type | Verification |
|-----------|------|--------------|
| 위 `pkg/config`, `pkg/content`, `pkg/adapter/{claude,codex,antigravity,omp,opencode}`, `internal/cli`, `templates/{claude,codex,gemini,shared}`, `content/agents`, `pkg/spec` 경로와 줄 | existing | Read·`rg`(2026-10-15, 4d73aa91) |
| `git show v0.50.112·113·115·116·125`, auto 0.50.123, npm registry, Claude Code·Codex·Gemini CLI 문서 | existing (history, binary, registry, docs) | `git show`, 실행, `npm view`, WebFetch |
| `ReaFeaturesConf`, `ReaMCPAction`, `RuntimeConf.ReaMCP`, `pkg/reamcp`(`shape.go`, `node.go`, `ManagedPins`, `NodeVersionStatus`), `balancedRungAnalysis`, HB6, HU3, `reaNodeProbe`, `resolveReaDecision`, `rea_decision.go`, `rea_node_probe.go`, `init_rea.go`, `update_rea.go`, `doctor_rea.go`, `wizard_rea_step.go`, `claude_rea_mcp.go`, `codex_rea_mcp.go`, `antigravity_rea_mcp.go`, `claude_reverse_route.go`, `hooks_rea_test.go`, `content/agents/reverse-engineer.md`, `templates/shared/auto-reverse-contract.md.tmpl`, `templates/gemini/skills/auto-reverse/SKILL.md.tmpl`, `evidence/test-classification.md`, `evidence/manual-smoke-rea.txt` | [NEW] planned addition | 부재 확인(`rg`, `ls`) |

## Reviewer Brief
- Intended scope: REA 결정(lazy key)·등록·제거, Node gate, init·update·doctor, reverse-engineer 에이전트와 17개 파급, MCP 참조 예외, `/auto reverse`와 clean-room handoff.
- Explicit non-goals: `spec.md` `## Outcome Boundary` 목록(OMP·OpenCode writer, 사전 승인, background, Windows 형식, 공유 MCP 목록 등).
- Self-verified: A3 옛 바이너리 실행, yaml.v3 decode, 16개 배치 출하 범위, 인용 줄 번호, 문서 사실(Claude `tools`, Codex timeout 기본값, Gemini `mcp_server_*`), npm pin.
- Reviewer should focus on: Decision State Contract 쓰기 표와 판정, 소유 분류와 round trip, 비대화형 무추가, Preset·Historical 분류, clean-room 경계, Completion Debt only.

## Open Risks
- R1 opt-out과 gate에 막힌 opt-in 프로젝트는 v0.50.113–v0.50.125가 읽지 못한다(REQ-REA-17 줄, CHANGELOG). R2 auto의 login shell PATH와 Codex 프로세스 PATH가 다를 수 있다(nvm·volta). doctor가 버전을 보이고 gate는 지우지 않는다. R3 runtime capture는 사용자 권한으로 대상을 실행한다. 사전 승인 없음, 에이전트가 먼저 묻는다.
- R4 Codex 에이전트는 Autopus가 도구를 제한할 수 없어 clean-room이 instruction과 approval policy에 기댄다. 확인 gate와 executor 입력 제한이 경로를 막는다. R5 npx 첫 실행이 30초를 넘을 수 있다(A2). R6 OMP balanced 프로필이 `claude-opus-5-5`·`gpt-6.1-sol`을 새로 요구한다. catalog에 없으면 기존 blocker가 이름을 낸다.
- R7 명세 문자열은 대상 유래 비신뢰 데이터다(Trust Model, 확인 gate). R8 OpenCode reverse-engineer는 기존 규칙으로 task permission을 받는다. REA 도구가 없어 그 플랫폼에서는 증거를 만들지 못한다. R11 OMP의 MCP 도구 이름은 미확인이라 쓰지 않는다.

## Self-Verify Summary
- Q-CORR-04 | status: PASS | attempt: 1 | files: research.md, spec.md | reason: Reference Discipline이 existing(코드·history·binary·docs)과 [NEW]를 나누고 spec·plan의 새 이름에 [NEW]가 붙는다
- Q-COMP-05 | status: PASS | attempt: 1 | files: research.md, spec.md, plan.md, acceptance.md | reason: INV-REA-01–11이 REQ·task·Must 시나리오에 이어지고 분류·round trip·버전·집합·rewrite는 예상 값 oracle이다
- Q-COMP-06 | status: PASS | attempt: 1 | files: spec.md, research.md | reason: Traceability Matrix가 REQ 20개를 잇고 Reviewer Brief가 범위·non-goal·focus를 좁힌다
- Q-COMP-07 | status: PASS | attempt: 1 | files: research.md, spec.md | reason: CD-1·CD-2와 Evolution Ideas가 나뉘고 후자는 번호가 없다
- Q-COMP-08 | status: PASS | attempt: 1 | files: plan.md | reason: A3 PASS는 실행 증거 파일, A1·A2는 not-run 이유, plan 진술 분류, scope expansion 규칙이 있다
