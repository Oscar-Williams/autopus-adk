# SPEC-REA-001 리서치

branch `feat/rea-001`, 기준 origin/main `4d73aa91`(PR #197 merge 뒤)에서 2026-10-15에 읽었다. 근거는 worktree Read·`rg`, `git show <tag>`, auto 0.50.123 실행과 문서 확인(`evidence/probe-rea-forward-compat.md`), npm registry다. revision 2·3은 review round 1(J-01–J-29)과 round 2(J-22, J-14, F-001)를, revision 4는 Gate 1 사용자 결정을 반영했다(`evidence/review-rounds.md`). 확인하지 못한 것은 implementation_assumption이나 not-run probe로 적었다. 인용한 요청·ledger 문구는 비신뢰 입력 증거로만 요약했다.

## 기존 코드 분석
- MCP: Claude 템플릿 `templates/claude/mcp.json.tmpl:1-15`, merge `pkg/adapter/claude/claude_files.go:25-87`(decode 뒤 `json.MarshalIndent`, managed 서버를 지우고 다시 넣음), 판정 `claude_clean.go:209-229`, clean `:176-207`. Codex `templates/codex/config.toml.tmpl:30-32`, merge `codex_config_structural.go:48-127`(key 단위 `:10-27`, header는 남김), clean `codex_clean_merge.go:180-213`, 호출 `codex_settings.go:24-52`. Antigravity `templates/gemini/settings/settings.json.tmpl:1-12`, merge `antigravity_settings.go:18-60,187-198`(덮기만 함), Generate 쓰기 `antigravity.go:130-137`(`os.WriteFile`), clean `antigravity_lifecycle_clean.go:69-110`(DeepEqual 서버만, `os.WriteFile`). 경로 확인 helper `adapter.SafeWorkspacePath`(`pkg/adapter/transaction_helpers.go:278`, symlink를 건너는 경로 거부). 사전 승인 MCP는 `mcp__sequential-thinking__sequentialthinking` 하나(`pkg/content/hooks.go:196`)이고 `enableAllProjectMcpServers`는 없다.
- 설정·CLI: `FeaturesConf{CC21}`(`pkg/config/schema.go:97-100`), `RuntimeConf`(`runtime.go:13-15`, yaml `-`), strict decode `loader_strict.go:98-125`, typed `Save`(`loader.go:130-153`). `cloneHarnessConfig`(`internal/cli/update_preview.go:20-35`)는 YAML round trip이고 preview(`update_config_preview.go:24`)·감지 플랫폼(`platform_update.go:64`)·platform add/remove(`platform.go:106,184`)가 쓴다. 정리 명령은 `auto platform remove`(`platform.go:142-195`, `cleanPlatformThenSave` → adapter Clean)뿐이고 `clean` 명령은 없다. init은 `DefaultFullConfig`를 새로 만든다(`init.go:68-69`). wizard seam `runInitWizard = tui.RunInitWizard`는 결과만 바꾼다(`quality_compose_driver_gate.go:24-26`). update 순서: 감지(L167) → 언어 질문(L185-188) → compose(L206) → 기존 플랫폼(L226) → 감지 플랫폼(L248), preview는 L112-164에서 끝난다. compose gate `composeGate.mode()`·`composeCITruthy`(`:31,66`), locale seam `composeLanguage`(`quality_compose_text.go:8`), `ActionSkip`(`pkg/adapter/manifest.go:137`).
- 에이전트·프리셋: 정규 16개 `quality_tier.go:17-21`, balanced rung `role_model_policy_balanced.go:20-68`(rung마다 후보 하나), tier `native_balanced.go:33-38`, ultra `defaults.go:136-147`, registry `quality_preset_history.go:52-104`, capability `role_model_policy_matrix.go:38-55`, opus 사다리 `quality_effective.go:53-70`·`codex_profile.go:118-135,182-191`, native 접힘 `omp_native_agents.go:11-25`, 그룹 `quality_compose_groups.go:24-37`. OMP 후보 부재 blocker `omp_profile_candidate_unavailable`(`platform_omp_profile_preview.go:75`), catalog fixture `platform_omp_profile_flags_test.go:74,87`, `quality_compose_omp_autopus_test.go:29`(Opus 5.5·Sol 없음).
- 도구 변환·라우트: Claude는 model·effort만 고친다(`claude_agent_content.go:67-104`). Gemini 허용 목록 `agent_transformer.go:159-178`, OMP `agent_transformer_omp_render.go:135-155`, OpenCode `agent_transformer_opencode.go:60-80`, Codex TOML 도구 칸 없음(`agent_transformer.go:113`). MCP 참조 변환 `skill_transformer_replace_mcp.go:22,50`. 라우트 Claude `claude_workflow_skills.go:22-42`, router 표 `auto-router.md.tmpl:55-75`, Antigravity command `templates/gemini/commands/auto/*.toml.tmpl`(19개, `router_budget_test.go:42-66`), Codex `codex_workflow_specs.go:14-17`, OMP `omp_specs.go:25`, OpenCode `opencode_specs.go:28`. executor 입력 `content/agents/executor.md:31-45,104-106`, tester 입력 `tester.md:75-77`.

## Outcome Lock
- User-visible outcome: Node 조건을 충족한 새 `auto init` 프로젝트에서 Claude Code·Codex·Antigravity가 프로젝트 범위 REA MCP(`npx -y rea-agents@6.5.0 mcp`)를 띄운다. `/auto reverse`로 reverse-engineer가 evidence_id가 붙은 행동 명세와 unknowns를 내고, 메인 세션에서 사용자가 동의·확인한 명세만으로 executor가 구현하고 tester가 검증한다. 기존 프로젝트는 대화형 update에서 한 번만 묻고 비대화형 실행은 REA를 더하지 않는다. 사용자 범위 파일은 건드리지 않는다.
- Mandatory requirements: REQ-REA-01–17, 20, 21 (Must). REQ-REA-18, 19는 Should.
- Explicit non-goals: `spec.md` `## Outcome Boundary`.
- Completion evidence: Must S1–S19 PASS, A1·A2 PASS, `auto eval harness run`, Windows cross-build, 300줄, `evidence/manual-smoke-rea.txt`, `auto spec validate --strict`.

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
- unresolved_fields: done_evidence(assumed → Completion Debt CD-1·CD-2). Gate 1 확인 대상은 `spec.md` Open Issues 4건이다.

## Open Question Decisions
| Q | 결정 | 근거 |
|---|---|---|
| Q1 | `features.rea.enabled`(*bool)는 managed 항목이 결정을 담지 못할 때만 쓴다. M은 쓰기 뒤 다시 읽는다 | A3: key는 옛 auto의 모든 명령을 막고 managed 항목은 옛 update 뒤에도 남는다. `ActionSkip`(J-17) |
| Q2 | balanced에 opus rung `balancedRungAnalysis`. OMP catalog에 그 model이 없으면 reverse-engineer만 세션 model, 기존 rung은 blocker | matrix 불변식, 도구 간 같은 model. J-01은 Gate 1에서 사용자가 "세션 모델로 대체"를 골랐다 |
| Q3 | Codex는 `startup_timeout_sec = 30`만, `tool_timeout_sec`는 기본 60초 | Codex 문서 기본 10초·60초, A2 not-run |
| Q4 | `tools: Read, Grep, Glob, mcp__rea`(보안 감사 M2로 `mcp__rea__<정적 도구>` 135개로 좁힘, `evidence/rea-tool-catalog-6.5.0.md`), 실행 모드에 기대지 않음, 동의·확인은 메인 세션 | Claude 문서의 server-level 이름, subagent에 AskUserQuestion 없음과 fork mode(J-05, J-06) |
| Q5 | CLI가 실행당 1회 exec probe(POSIX login shell, Windows PATH, 5초), 127은 absent | 서버 실행 형식과 같은 shell(J-19) |
| Q6 | `--auto`도 동의 질문과 확인 gate에서 멈춘다 | PRD Anti-goal "자동 승인으로 끝까지", 사용자 결정 2 |
| Q7 | 정확한 모양 + pin 목록, Claude 두 pin 자리 일치, Codex 하위 table은 user_owned | R6, J-13, J-16 |
| Q8 | 공유 MCP 목록은 만들지 않는다. REA 모양·변환만 `pkg/reamcp` 한 곳 | 템플릿 불변으로 REQ-REA-16, 소비처 5곳(세 adapter, CLI sweep, doctor) |
| Q9 | `.autopus/reverse/<slug>/behaviour-spec.md`, slug 규칙과 `-2` 보존, 생성 `.gitignore`로 로컬 전용 | 에이전트에 쓰기 도구가 없다. 제3자 대상 유래이고 `rootTrackedPrefixes`(`sync_verify_policy.go:26-29`)에 없다(J-14, J-24) |
| Q10 | HB6(balanced 16, v0.50.116–v0.50.125, runtimeStandard), HU3(ultra 16, v0.50.115–v0.50.125) | `git show v0.50.115/116/125` 비교 |
| Q11 | OMP·OpenCode 에이전트에 MCP 이름 없음, 안내는 세 도구에서 분석하라고 말한다. Antigravity는 tools에 MCP 이름을 쓰지 않고 `inheritMcp: true`(2026-10-17, `## agy 1.4 Smoke`. 이전 결정 `mcp_rea_<정적 도구>`는 agy가 모르는 도구로 거부했다) | OMP 이름 미확인(R11), J-12, agy 1.4.0 probe |
| 정리 | 정리 경로는 `auto platform remove`. off는 선택되지 않은 파일도 sweep | `clean` 명령 없음(J-02), J-18 |
| 보존 | JSON은 값 보존 + 기존 정규화, Codex는 bytes | 기존 merge가 다시 marshal한다(J-03) |
| handoff | tester(Test Scaffold) → executor → tester(Testing), `SPEC: N/A`·`Task: reverse/<slug>` | executor·tester 입력 계약(J-07) |
| action | `cloneHarnessConfig`가 `Runtime.ReaMCP`를 옮기고 preview 전에 해석 | clone이 `yaml:"-"`를 잃는다(J-08) |
| 동의 입구 | `/auto reverse … --runtime`이 첫 결과를 Runtime Capture Request로 만든다 | 허용 경로를 S19에서 결정적으로 관측(J-22) |

### Q1 forward-compat 비교 (A3 실행 근거)
| 안 | v0.50.113–v0.50.125 | undecided 구분 | 대가 |
|---|---|---|---|
| A 늘 `features.rea` | 새 init 프로젝트 전부에서 모든 명령 실패(update 종료 1) | key 하나, 단순 | MODELUX-003보다 넓다(그쪽은 opt-in일 때만) |
| B `future_extension.rea` | 읽힌다 | 옛 update가 block을 지워 결정이 사라진다 | reserved 이름 남용, loader 변경 |
| D 별도 상태 파일 | 영향 없음 | 된다 | ledger의 features flag와 다르고 새 소유 파일이 생긴다 |
| F(선택) 항목이 담지 못할 때만 key | 기본 경로에서 단절 없음. opt-out과 gate에 막힌 opt-in만 단절 | key > managed 항목 > undecided | 판정이 대상 파일을 읽는다. managed 항목을 손으로 모두 지우면 한 번 더 묻는다 |

### Q2 balanced opus 비교
| 안 | matrix 불변식 | 도구 간 model | 대가 |
|---|---|---|---|
| 새 opus rung(선택) | 지킨다 | Claude Opus 5.5·Codex Sol·OMP 같은 후보 | OMP catalog에 model이 없으면 reverse-engineer만 세션 model로 내려간다(사용자 결정, 대체 줄 1개) |
| top rung | 지킨다 | Fable·Astra | 사용자 결정 `opus`와 다르다 |
| matrix 밖 | 깬다 | OMP만 deep_reasoning route(Fable) | 정규 집합 일치 test에 예외 |

## Pinned Version
| 항목 | 값 | source | checked_at |
|---|---|---|---|
| REA | `rea-agents@6.5.0`(MIT, `dist-tags.latest` 6.5.0, bin `rea`·`rea-agents`) | `npm view rea-agents@6.5.0`, REA README | 2026-10-15 |
| Node 범위 | `^22.19.0`, `^24.11.0`, `>=26.0.0` 중 하나 | 같은 결과의 `engines.node` | 2026-10-15 |
- greenfield가 아니라 Technology Stack Decision은 두지 않는다. rea-agents는 Go 의존성이 아니라 사용자의 AI 도구가 서버를 띄울 때 npx가 받는 runtime 의존성이고 auto 자신은 네트워크를 쓰지 않는다.

## Visual Planning Brief
흐름도(init·update·doctor command-flow, 동의 loop와 clean-room 경계가 있는 `/auto reverse` sequence)와 wireframe은 `plan.md`이고 `wireframe intent: assumed`다.

## Minimality Decision Matrix
| Ladder step | Evidence | Decision | Receipt item |
|-------------|----------|----------|--------------|
| actual need | Outcome Lock, ledger goal·constraints, PRD P0 17개, review round 1 | proceed | 등록·결정·에이전트·라우트·동의·clean-room |
| existing code/helper/pattern | 세 adapter merge·Clean, `auto platform remove`, `SafeWorkspacePath`, `WriteFileIfChanged`, `composeGate.mode()`, `composeLanguage`, wizard Confirm, Historical registry, route customBody, executor·tester 입력 계약 | reuse | 판단 밖 경로는 기존 함수 |
| stdlib/native | `strconv`·`strings`(버전), `encoding/json`, `os/exec`(probe), `context`(5초) | use | semver 라이브러리 없음 |
| existing dependency | yaml.v3(`*bool`), huh(Confirm, accessible test) | reuse | 새 Go 의존성 0 |
| new dependency or new abstraction | 새 Go 의존성 없음. rea-agents는 runtime npx 의존성. 새 abstraction은 `pkg/reamcp`(모양·분류·변환·exec 판정), `ReaMCPAction`, `balancedRungAnalysis`, Runtime Capture Request, Reverse Handoff 값 | accepted | 소비처 5곳이 같은 모양을 써야 해서 한 곳 |
| minimum sufficient verification | 실제 adapter·명령 test, seam, BASE 대조 bytes, symlink 사례, 수동 smoke 2건과 A1·A2 | required checks | security·data-loss·deterministic-oracle gate 유지 |

## Semantic Invariant Inventory
| ID | source clause | invariant type | affected outputs | acceptance IDs |
|----|---------------|----------------|------------------|----------------|
| INV-REA-01 | "persists the answer so it is never asked again" | state transition: key > managed 항목 > undecided, 질문은 undecided 대화형에서만 | 질문 수, autopus.yaml, MCP 파일 | S2, S3, S7 |
| INV-REA-02 | "non-TTY never adds rea and prints a one-line hint" | 비대화형 undecided: rea 항목·key 쓰기와 REA stdin 0, hint 1 | MCP 파일, autopus.yaml, stdout | S7 |
| INV-REA-03 | "exactly one Autopus-managed rea entry, user entries untouched" | classification: 정확한 모양 + pin 목록, 사용자 항목 값 보존 | 분류 값, entry 값 | S4, S8 |
| INV-REA-04 | "remove only the managed entry, no residue" | round trip: off 뒤 파일 = BASE 정규화 결과 | 세 파일 bytes | S2, S5 |
| INV-REA-05 | "Node ^22.19.0 or ^24.11.0 or >=26.0.0" | ordering: exec 결과와 version triple 비교 | pass·unsupported·absent·unknown | S6 |
| INV-REA-06 | "agent inventory 17 across all generated surfaces" | set equality with CanonicalAgentNames, OMP 17 + task | 에이전트 파일 이름, override key | S9, S11 |
| INV-REA-07 | "existing 16-agent presets must not be misread as customisations" | classification: 정확한 map → Historical row | class, Row, RuntimeStandard | S10 |
| INV-REA-08 | "mcp__rea__x preserved, mcp__other__y still rewritten" | parser rewrite per reference | 렌더된 줄 | S12 |
| INV-REA-09 | "executor implements from the confirmed spec only" | data-flow boundary | detail 문장, dispatch 입력 | S13, S19 |
| INV-REA-10 | "MCP files unchanged while REA is not enabled" | bytes equality | 세 파일 bytes | S15 |
| INV-REA-11 | "never touch user-level files, no pre-approved permissions" | absence over write set, symlink 거부 | HOME sandbox, 밖 파일 bytes, permissions.allow | S14, S16 |
| INV-REA-12 | "runtime capture runs targets with user permissions" | state transition: 요청 → 메인 세션 허용 → 호출, 거절이면 호출 0 | 동의 질문, 도구 호출 수 | S13, S19 |
| INV-REA-13 | Gate 1 "세션 모델로 대체" | partition: 새 analysis rung 후보 부재 → reverse-engineer key 없음 + 대체 줄 1, 기존 rung·operator pin 부재 → blocker | override key 집합, 출력 줄, 종료 상태 | S11 |

## Feature Coverage Map
| Outcome slice | Covered by | Status |
|---------------|------------|--------|
| Happy path: 새 init → 세 파일에 rea, 17 에이전트, `/auto reverse` | REQ-REA-02, 03, 10, 11, 13, 14 / S3, S9, S13, S18, S19 | covered |
| Error/recovery: Node 없음·구버전, flag 충돌, off·platform remove, 사용자 항목·삭제 파일, 미등록, 동의 거절, gate 수정·중단 | REQ-REA-04–07, 10, 13, 21 / S2–S7, S9, S13, S19 | covered |
| Integration boundary: 세 writer·Clean·sweep, symlink, OMP·OpenCode 안내, 5개 route·command, OMP 17 + task와 analysis rung 대체, preset·history, MCP 참조 변환, clone, 옛 바이너리 | REQ-REA-03, 09, 11–13, 16, 17 / S4, S5, S7, S10–S12, S15, S16, A3 | covered |
| CLI surface: wizard, flag, update 1회 질문(CI 포함), hint, doctor, 옛 바이너리 줄, ko/en | REQ-REA-02, 06–08, 17, 18 / S3, S7, S8, S20 | covered |
| Verification: 결정적 test, 실제 터미널, 동의한 live, Windows build, eval | REQ-REA-20 / S17–S19, A1, A2 | covered(S18·S19는 CD-1·CD-2) |
| Docs/ops: README, CHANGELOG, help, pin, 사전 승인 없음, 한계 | REQ-REA-15, 19 / S14, S21 | covered |

## Completion Debt
| Item | Blocks | Required resolution |
|------|--------|---------------------|
| CD-1 실제 터미널(모델 호출 없음): Codex를 먼저, 도구마다 빈 `npm_config_cache`로 열어 cold를 cache 상태로 증명한 30초 시작, Claude `/mcp`의 rea connected와 도구 목록, agy `/mcp`의 `autopus_rea` connected와 도구 수 | REQ-REA-03의 런타임 증거, A2 시작 절반 | S18 PASS 기록과 A2 시작 절반 PASS. FAIL이면 고친 뒤 다시 |
| CD-2 동의 뒤 live: Claude 경로(확인, 수정·중단, `--runtime` 거절, `--runtime` 허용 뒤 재-dispatch와 capture 호출, 미등록 1회 dispatch), agy 에이전트 시작(`--agent` 3개 `OK`)과 reverse-engineer의 `call_mcp_tool` 정적 도구 호출, Codex 도구 호출 60초 | REQ-REA-10, 14, 21의 런타임 증거, A1, A2 도구 절반 | S19 PASS 기록과 A1·A2 PASS. 기록만으로는 닫히지 않는다 |

## agy 1.4 Smoke
2026-10-17 실제 터미널 smoke와 probe(Antigravity CLI 1.4.0, `strings`로 본 바이너리 struct tag·내장 changelog·도구 설명, scratch 프로젝트의 local echo MCP 서버).
- MCP 위치: agy는 사용자 `~/.gemini/config/mcp_config.json`과 `<plugin>/mcp_config.json`에서만 MCP 서버를 읽는다. 프로젝트 `.gemini/settings.json`의 `mcpServers`로는 `/mcp`에 서버가 없었고, `.agents/plugins/autopus/mcp_config.json`의 `{"mcpServers": {...}}`는 `autopus_<server>`로 connected였다(1.2.2: 플러그인 서버는 `<plugin>_<server>`). `agy mcp list`는 사용자 범위만 보인다.
- 에이전트 위치: agy는 `.agents/plugins/<plugin>/agents/`와 `.agents/agents/`를 읽고 `.gemini/agents/`는 읽지 않는다. Gemini CLI 0.52는 `.gemini/agents/*.md`만(하위 디렉터리 없이) 읽어, `.gemini/agents/autopus/`는 어느 쪽도 읽지 않는다. 같은 이름의 워크스페이스 플러그인은 사용자 범위의 오래된 `autopus` 플러그인을 가린다.
- 도구 이름: `tools:`에 registry 밖 이름이 하나라도 있으면 `failed to construct executor: ... unknown component: tool "<name>" not found in registry`로 모델 호출 전에 끝난다(종료 3). 해석됨: `read_file`, `view_file`, `write_file`, `write_to_file`, `edit_file`, `replace_file_content`, `multi_replace_file_content`, `grep_search`, `find_by_name`, `list_dir`, `run_command`, `search_web`, `read_url_content`. 거부: `glob`, `replace`, `run_shell_command`, `codebase_search`, `command_status`. 고른 대응: Edit → `edit_file`(core 편집 도구 `cortex/core/tools/edit_file.go`, Target·Replacement가 한 곳에 정확히 맞아야 하고 여러 곳은 AllowMultiple. 도입 버전은 내장 changelog에 없고 1.4.0에서 확인했다), Glob → `find_by_name`(fd, glob 패턴), Bash → `run_command`.
- MCP와 subagent: Markdown 에이전트 frontmatter의 MCP field는 `inheritMcp`(바이너리 tag `yaml:"inheritMcp"`, 1.1.6 changelog)이고 `enable_mcp_tools`는 `define_subagent` 도구의 JSON 인자다. probe에서 `tools: [read_file]`인 플러그인 subagent 다섯 변형(key 없음, `inheritMcp` true·false, `enable_mcp_tools` true·false)이 모두 `call_mcp_tool`로 echo 서버를 불렀다(서버 기록의 `parent_conversation_id`, `--dangerously-skip-permissions`). print mode에서는 같은 호출을 운영자 없이 `soft-denying`했다. 그래서 1.4.0의 tools 목록은 어떤 subagent의 MCP도 막지 못하고 호출별 권한만 남는다. 이는 reverse-engineer만의 일이 아니다: REA가 등록된 project에서는 executor·tester를 포함한 모든 Autopus 에이전트가 `autopus_rea`의 runtime 도구까지 부를 수 있고, 막는 것은 에이전트 본문과 agy의 호출별 MCP 권한이다.
- rule 적용(probe): 플러그인 rule 중 `trigger: always_on`을 단 것만 메인 세션과 subagent 모두에 들어갔고, `trigger`가 없는 rule(지금 Autopus 플러그인 rule의 형식)은 어느 쪽에도 없었다. 프로젝트 `GEMINI.md`는 둘 다 읽었다. 그래서 REA 가드는 `trigger: always_on` 플러그인 rule `rea-mcp-guard.md`로 둔다. 다른 Autopus 플러그인 rule에 `trigger: always_on`이 없어 agy에서는 `GEMINI.md`가 `@`로 부르는 다섯 개(lore-commit, file-size-limit, subagent-delegation, language-policy, techstack-freshness)만 늘 들어가고 나머지는 model 판단으로만 읽힌다. 이 SPEC 밖의 기존 상태이고 후속(v0.50.127)으로 남긴다. 이번에는 코드를 바꾸지 않는다.
- read-only 역할: agy Markdown 에이전트 구조체에는 read-only·plan field가 없다. `commandExecutionPolicy`는 명령 승인 정책이고, probe에서 `request-review`·`strict`·임의 값을 준 에이전트는 agy가 읽지 않았다. 그래서 Write·Edit가 없는 역할도 source의 Bash대로 `run_command`를 받는다.
- live 확인(새 init project): `/mcp`에 `autopus_rea`가 도구 207개로 connected였고, 메인 세션이 보낸 reverse-engineer subagent가 `autopus_rea/list_unknowns` 승인 요청을 띄운 뒤 허용하자 결과를 돌려주었다. `agy --agent auto-agent-<explorer·executor·reverse-engineer> -p 'Reply with exactly: OK'`는 셋 다 `OK`(종료 0)였고 RC가 만든 explorer는 `glob`·`run_shell_command`로 종료 3이었다.

## Evolution Ideas
These are optional improvements and do not block sync completion.

| Idea | Why not required now | Promotion trigger |
|------|----------------------|-------------------|
| OMP·OpenCode 프로젝트 MCP writer와 그 에이전트의 REA 도구 이름 | 사용자 범위 경계, OMP 이름 미확인 | OMP·OpenCode 사용자 요청 |
| 모든 MCP 서버의 Windows 실행 형식 | 기존 서버도 POSIX 형식이다 | Windows 사용자 보고 |
| JSON 사용자 서식을 지키는 원문 편집기 | 기존 merge가 모든 update에서 이미 정규화한다 | 서식 보존 요청 |
| `auto platform add`가 REA 결정을 풀어 바로 등록 | 다음 update가 더한다 | 사용자 요청 |
| doctor의 REA 새 버전 안내, 읽기 전용 REA 도구 사전 승인, Codex 에이전트별 sandbox | 수동 pin, 도구 감사 전, 역할 설정 미확인 | 사용자 요청 |
| Antigravity 플러그인 rule 전부에 `trigger: always_on`(후속 v0.50.127) | 지금은 `GEMINI.md`가 `@`로 부르는 다섯 개만 agy에서 늘 들어간다(`## agy 1.4 Smoke`). REA 가드는 이미 always_on이고 나머지 rule의 적용 범위를 바꾸는 일은 이 SPEC 밖이다 | v0.50.127 계획 |

## Sibling SPEC Decision
| Decision | Reason | Sibling SPEC IDs |
|----------|--------|------------------|
| none | Outcome Lock은 등록 → 분석 → 동의·확인 → 구현의 한 흐름이다. "MCP 등록만"은 에이전트 안내가 `--rea`에 기대고, "16→17 파급"은 배포·migration 순서 경계가 없다. 같은 module·같은 릴리스라 나누면 순서 의존만 생긴다. task 22개, 소스 파일 test 포함 약 55개로 기준(25개 초과 task이면서 40개 초과 파일)을 동시에 넘지 않는다 | None |

## Reference Discipline
| Reference | Type | Verification |
|-----------|------|--------------|
| 위 `pkg/config`, `pkg/content`, `pkg/adapter/{claude,codex,antigravity,omp,opencode}`, `pkg/adapter/{transaction_helpers,writeutil,manifest}.go`, `internal/cli`(`update_preview.go`, `platform.go`, `platform_update.go`, `quality_compose_driver_gate.go`, `quality_compose_text.go`), `templates/{claude,codex,gemini,shared}`, `content/agents/{executor,tester}.md`, `pkg/spec` 경로와 줄 | existing | Read·`rg`(2026-10-15, 4d73aa91) |
| `git show v0.50.112·113·115·116·125`, auto 0.50.123, npm registry, Claude Code·Codex·Gemini CLI 문서, review round 1이 인용한 Claude 문서(fork mode, `/agents`) | existing (history, binary, registry, docs) | `git show`, 실행, `npm view`, WebFetch, review 인용 |
| `ReaFeaturesConf`, `ReaMCPAction`, `RuntimeConf.ReaMCP`, `pkg/reamcp`(`shape.go`, `node.go`, `ManagedPins`), `balancedRungAnalysis`, HB6, HU3, `reaNodeExec`, `resolveReaDecision`, `rea_decision.go`, `rea_node_probe.go`, `init_rea.go`, `update_rea.go`, `doctor_rea.go`, `wizard_rea_step.go`, `claude_rea_mcp.go`, `codex_rea_mcp.go`, `antigravity_rea_mcp.go`, `claude_reverse_route.go`, `hooks_rea_test.go`, `content/agents/reverse-engineer.md`, `templates/shared/auto-reverse-contract.md.tmpl`, `templates/gemini/skills/auto-reverse/SKILL.md.tmpl`, `templates/gemini/commands/auto/reverse.toml.tmpl`, `evidence/test-classification.md`, `evidence/manual-smoke-rea.txt` | [NEW] planned addition | 부재 확인(`rg`, `ls`) |

## Reviewer Brief
- Intended scope: REA 결정(lazy key)·등록·제거(off sweep, `auto platform remove`), Node gate, init·update·doctor, reverse-engineer 에이전트와 17개 파급, MCP 참조 예외, `/auto reverse`의 메인 세션 동의·확인과 handoff.
- Explicit non-goals: `spec.md` `## Outcome Boundary` 목록(새 정리 명령, JSON 서식 보존, foreground 보장, OMP·OpenCode writer, 사전 승인, Windows 형식 등).
- Self-verified: A3 옛 바이너리 실행, yaml.v3 decode, 출하 범위, round 1이 지적한 코드 사실(clone, wizard seam, platform remove, merge 재marshal, executor·tester 입력, Antigravity command, OMP fixture), 문서 사실, npm pin.
- Reviewer should focus on: round 1·2와 Gate 1 반영의 정확성, Decision State 쓰기와 M, 보존 범위와 BASE 대조, 동의 상태 전이, handoff 입력, OMP analysis rung 대체가 기존 rung을 바꾸지 않는지, Completion Debt only.

## Open Risks
- R1 opt-out과 gate에 막힌 opt-in 프로젝트는 v0.50.113–v0.50.125가 읽지 못한다(REQ-REA-17 줄). v0.50.125는 17개 preset을 customized로 본다(J-28, REQ-REA-19 문서). R2 login shell PATH와 Codex 프로세스 PATH가 다를 수 있다. doctor가 버전을 보이고 gate는 지우지 않는다. R3 runtime capture는 사용자 권한으로 대상을 실행한다. 메인 세션 동의 뒤에만 부른다(REQ-REA-21).
- R4 Codex 에이전트는 도구를 제한할 수 없어 clean-room이 instruction, 메인 세션 gate, handoff 입력 제한에 기댄다. 보안 감사 M1(2026-10-16)로 reverse-engineer 역할 파일에 `sandbox_mode = "read-only"`(Codex 0.162.0 역할 파일 field 표에서 확인)를 더해 shell 쓰기는 막지만, MCP 호출은 sandbox 밖이라 runtime REA 도구 금지는 여전히 instruction과 메인 세션 동의 질문의 권고 수준이다. R5 npx 첫 실행이 30초를 넘을 수 있다(A2). R6 catalog에 Opus 5.5·Sol이 없는 OMP balanced 프로젝트에서 reverse-engineer는 세션 model로 돈다(사용자 결정). 품질이 세션 model에 달리지만 실행은 막히지 않고 대체 줄이 알린다.
- R7 명세 문자열은 대상 유래 비신뢰 데이터다(Trust Model, 확인 gate). R8 OpenCode reverse-engineer는 `task` permission이 없어 다른 에이전트를 보내지 못하고(보안 M1) REA 도구가 없어 증거를 만들지 못한다. R9 runtime capture 도구는 rea-agents@6.5.0 `toolEffects.js`의 효과 감사로 이름이 정해졌다(정적 135, runtime 68, 호출자 경로 쓰기 4, `evidence/rea-tool-catalog-6.5.0.md`). Claude Code subagent는 정적 도구만 이름으로 받는다(보안 M2). R10 Antigravity에서는 agy가 Markdown 에이전트에 MCP 도구를 하나씩 주지 못하고 `tools:`가 MCP를 막지 못해, REA가 등록된 project의 모든 Autopus 에이전트(reverse-engineer, executor, tester 등)에 REA runtime 도구까지 보인다(`## agy 1.4 Smoke`). always-on 플러그인 rule `rea-mcp-guard.md`가 `autopus_rea` 호출을 reverse-engineer와 `/auto reverse` 흐름으로 제한하지만, runtime 도구 금지는 R4의 Codex처럼 rule·본문과 메인 세션 동의, agy가 호출마다 묻는 MCP 권한의 권고 수준이다. 그 권한 질문의 허용은 runtime capture 동의가 아니고, 운영자가 "항상 허용"을 고르면 모든 subagent에 적용되므로 호출마다 답한다.

## Self-Verify Summary
- Q-CORR-01 | status: PASS | attempt: 2 | files: spec.md, research.md, acceptance.md | reason: `auto clean`을 `auto platform remove`로 바꾸고 clone·wizard seam·merge 재marshal·executor 입력을 코드로 다시 확인했다
- Q-CORR-02 | status: PASS | attempt: 2 | files: spec.md, plan.md, research.md | reason: 새 명령은 없고 새 파일·함수·template(`reverse.toml.tmpl` 포함)은 모두 [NEW]다
- Q-CORR-04 | status: PASS | attempt: 2 | files: research.md | reason: Reference Discipline이 existing(코드·history·binary·docs·review 인용)과 [NEW]를 나누고 wizard seam 범위를 바로 적었다
- Q-COMP-02 | status: PASS | attempt: 2 | files: plan.md | reason: off sweep(T9), action 전달(T12 clone), handoff 입력(T20), Antigravity command(T18), fixture(T16)에 owner task가 있다
- Q-COMP-04 | status: PASS | attempt: 3 | files: spec.md, acceptance.md, research.md | reason: 동의 허용 경로가 `--runtime` fixture로 S19 (e)의 Must 실행 시나리오가 되었고 CD-2가 그 관측에 묶인다
- Q-COMP-05 | status: PASS | attempt: 3 | files: research.md, spec.md, plan.md, acceptance.md | reason: INV-REA-01–12가 Must 시나리오에 이어지고, S18 cold는 cache 상태로, S19 동의는 호출 수로, S19(d)는 dispatch 1회로 판정해 route와 맞는다
- Q-COMP-06 | status: PASS | attempt: 1 | files: spec.md, research.md | reason: Traceability Matrix가 REQ 21개를 잇고 Reviewer Brief가 범위·focus를 좁힌다
- Q-COMP-07 | status: PASS | attempt: 1 | files: research.md, spec.md | reason: CD-1·CD-2와 Evolution Ideas가 나뉘고 후자는 번호가 없다
- Q-COMP-08 | status: PASS | attempt: 2 | files: plan.md | reason: A3 PASS는 실행 증거, A1·A2는 not-run 이유와 닫는 시나리오, 진술 분류, scope expansion 규칙이 있다
- Q-FEAS-01 | status: PASS | attempt: 2 | files: spec.md, plan.md | reason: 보존 범위를 기존 merge에 맞췄고 동의·확인을 메인 세션으로, handoff를 executor·tester 입력에 맞췄다
- Q-FEAS-02 | status: PASS | attempt: 3 | files: plan.md, spec.md | reason: `reverse.toml.tmpl`, fixture 규칙 7·8과 함께 behaviour-spec을 `.gitignore`로 로컬 전용으로 정하고 T18·규칙 9에 소유를 두었다
- Q-FEAS-03 | status: PASS | attempt: 3 | files: acceptance.md, plan.md | reason: S18은 Codex를 먼저 도구별 빈 cache로 재고 cache 상태로 cold를 증명하며, S19(d)는 route 1단계와 같은 dispatch 1회 oracle이다
- Q-SEC-01 | status: PASS | attempt: 2 | files: spec.md, acceptance.md | reason: Runtime Capture Request → 메인 세션 허용 → 재분석, 거절이면 호출 0(REQ-REA-21, S19)
- Q-SEC-02 | status: PASS | attempt: 2 | files: spec.md, plan.md, acceptance.md | reason: REA 쓰기는 `SafeWorkspacePath`를 거치고 S16이 파일·부모 symlink 네 사례로 밖 bytes 불변을 판정한다
- Q-COH-02 | status: PASS | attempt: 2 | files: research.md, plan.md | reason: CD-1·CD-2는 S18·S19 기록과 A1·A2 PASS가 모두 있어야 닫힌다
