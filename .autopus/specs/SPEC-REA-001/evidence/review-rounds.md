# SPEC-REA-001 review 반영표

## Round 1 (2026-10-15, verdict REVISE: claude REVISE, codex REJECT)

| Finding | Severity | 반영 | 위치 |
|---|---|---|---|
| J-01 OMP balanced가 Opus 5.5·Sol을 새로 요구, fixture 규칙 부재 | major | 기존 fail closed 정책을 명시(REQ-REA-11), 규칙 7로 catalog fixture 세 개에 두 model 추가, S11 runner catalog 지정과 model 없는 사례, R6·CHANGELOG·Gate 1 Open Issue | spec Preset Contract·Existing Test Changes·Open Issues, acceptance S11, plan T16 |
| J-02 `auto clean`은 공개 명령이 아님 | major | 정리 경로를 `auto platform remove <platform>`(adapter Clean)로 바꾸고 새 명령은 non-goal | REQ-REA-04, Decision State Contract, S5, S16 |
| J-03 JSON merge가 다시 marshal해 bytes 보존이 거짓 | major | 정직하게 좁힘: JSON은 값 보존 + 기준 commit 정규화(BASE 대조 bytes), Codex는 bytes. 서식 보존은 non-goal, Gate 1 Open Issue | REQ-REA-03·04, MCP Entry Shape Contract, S4, S5, Oracle Notes |
| J-04 wizard seam은 결과만 바꿈 | major | tui 패키지에서 huh accessible + 주입 reader로 REA form을 직접 돌리고 CLI는 `runInitWizard` 결과 stub | REQ-REA-20, S3, T10, T11 |
| J-05 subagent에 AskUserQuestion 없음, permissionMode 덮임 | major | runtime capture 동의를 메인 세션으로(REQ-REA-21, Runtime Capture Request, 동의 줄), 도구 권한 질문은 동의가 아님 | Trust Model, Agent·Route Contract, S9, S13, S19 |
| J-06 fork mode에서 foreground 보장 불가, `/agents` UI 변경 | major | 실행 모드에 기대지 않음(non-goal), S13의 `run_in_background` 검사 삭제, S18은 `/mcp`만, 에이전트 로드는 S19 transcript로 | Agent Contract, S13, S18, S19, A1 |
| J-07 executor·tester 입력 계약 | major | Reverse Handoff Contract(`SPEC: N/A`, `Task: reverse/<slug>`, `Behaviour Spec:`, B-/A- ID, gate에서 확인한 경로, Phase), tester Scaffold 먼저, executor.md·tester.md 한 문장 | spec, T20, S13, S19, Open Issues |
| J-08 clone이 `RuntimeConf`를 잃음 | major | `cloneHarnessConfig`가 `Runtime.ReaMCP`를 옮기고 update는 preview 분기 전에 action을 정함, 감지 플랫폼 사례 | Decision State Contract, T12, S7 (14)(15) |
| J-09 Antigravity command template | minor | `[NEW] templates/gemini/commands/auto/reverse.toml.tmpl`, detail 경로 `.gemini/skills/autopus/auto-reverse/SKILL.md` | Route Contract, T18, S13 |
| J-10 `tui.DetectSystemLang`은 seam이 아님 | minor | `composeLanguage` seam으로 | REQ-REA-18·20, S20 |
| J-11 CI truthy | minor | 질문 판정에 `composeGate.mode()` 사용, CI 열과 S7 (16) | REQ-REA-06, Update Prompt Contract |
| J-12 OMP·OpenCode 안내가 안 되는 경로를 약속 | minor | 안내·미등록 문장이 "분석은 세 도구에서"라고 말함 | REQ-REA-09, Form Text |
| J-13 Claude 두 pin 자리 불일치 | minor | 불일치는 user_owned, S4 C7 | MCP Entry Shape Contract, S4 |
| J-14 Trust Model과 artifact 모순 | minor | behaviour-spec의 소유·commit 후보·보존 규칙 명시 | Trust Model, Route Contract |
| J-15 'write nothing' 문구 | minor(deferred) | 고침: rea 항목·`features.rea`·REA stdin으로 좁힘 | REQ-REA-06 |
| J-16 Codex 하위 table | minor | `[mcp_servers.rea.*]`가 있으면 user_owned, lexer 규칙, S4 X7 | MCP Entry Shape Contract, T4, S4 |
| J-17 M과 `ActionSkip` | minor | M은 쓰기 뒤 다시 읽음, 지운 파일은 다시 만들지 않고 `enabled: true` | Decision State Contract, S7 (13) |
| J-18 off 범위 | minor | off는 선택되지 않은 대상 파일도 `pkg/reamcp` 변환 + SafeWorkspacePath로 sweep | REQ-REA-04, T9, S2 (6) |
| J-19 exit 127 | minor | exec 경계 판정표(127 + 빈 stdout = absent), seam `reaNodeExec` | Node Gate Contract, T8, S6 |
| J-20 TTY stdin Read=0 모순 | minor | REA 질문 수만 세고 stdin Read는 non-TTY에서만 단언, compose 입력 Enter | S7, 규칙 8 |
| J-21 S4 전후 분류 모순 | minor | "update 전"과 "`--rea` 뒤" 열로 나눔 | S4 |
| J-22 CD 종료 oracle 부족 | minor | CD-1·CD-2가 A1·A2 PASS와 다섯 route 경로·Antigravity 호출·Codex 시간에 묶임 | Completion Debt, S18, S19, probe 표 |
| J-23 S16 BASE 비교 | minor | BASE는 지원 명령만, 새 flag는 branch 전후 직접 비교 | S16 |
| J-24 slug | minor | 규칙·fallback `feature`·예시 | Route Contract, S13 |
| J-25 재-init 우선순위 | minor | flag > 기존 결정 > wizard/on, 기존 key 이전, S3 (8)(9) | REQ-REA-02, T11 |
| J-26 `mcp__reader__x` | suggestion(deferred) | 고침: 첫 segment 정확 비교와 S12 L5 | REQ-REA-12, S12, T17 |
| J-27 Windows gate | suggestion(deferred) | 문서로 알림(POSIX 실행 형식 한계), gate 동작은 그대로 | REQ-REA-19, MCP Entry Shape Contract |
| J-28 v0.50.125의 17개 preset 분류 | suggestion(deferred) | 문서로 알림 | REQ-REA-19, R1 |
| J-29 REQ-REA-15 문구 | suggestion(deferred) | 고침: `permissions.allow`와 자동 승인 key로 한정 | REQ-REA-15 |
| F-REA-002 Antigravity symlink 경계 | major(codex) | REA 쓰기는 `SafeWorkspacePath`, Antigravity Generate·Clean에 확인 추가, S16 symlink 네 사례 | Trust Model, MCP Entry Shape Contract, T5–T7, S16 |

| Checklist FAIL | 반영 |
|---|---|
| Q-FEAS-02 | J-09 template, J-01 fixture 규칙 |
| Q-FEAS-03 | J-04 wizard 분리, J-01 runner catalog, J-10 seam, J-02 명령, J-06 smoke UI |
| Q-CORR-01, 02, 04 | J-02 `auto platform remove`, J-04 seam 범위, Reference Discipline 갱신 |
| Q-COMP-02 | J-18 sweep, J-08 action 전달, J-07 handoff에 owner task |
| Q-COMP-04 | 정리 경로·handoff·CD 종료 조건 |
| Q-COMP-05 | J-03 보존 oracle, J-20, J-21 |
| Q-FEAS-01 | J-03, J-06, J-07 |
| Q-SEC-01 | J-05 동의 상태 전이(REQ-REA-21) |
| Q-SEC-02 | F-REA-002 symlink 경계 |
| Q-COH-02 | J-22 CD 종료 조건 |

## Round 2 (2026-10-15, 27 resolved, 3 open)

| Finding | Severity | 반영 | 위치 |
|---|---|---|---|
| J-22 S18 cold cache를 보장하지 못함 | major | Codex를 다른 도구보다 먼저, 도구마다 새 빈 `npm_config_cache`와 Codex용 새 scratch HOME으로 재고 실행 전후 cache 상태로 cold를 증명한다. 증명하지 못하면 PASS가 아니다 | S18, plan A2, research CD-1 |
| J-22 동의 허용 경로에 Must 실행 시나리오 없음 | major | `/auto reverse … --runtime`이 첫 결과를 Runtime Capture Request로 만들고, S19 (c) 거절·(e) 허용을 로컬 fixture `fixture/hello.js`로 결정적으로 실행한다. 허용 전 호출 0, 재-dispatch prompt의 `Runtime capture consent:` 줄, 허용 뒤 호출 1번 이상 | REQ-REA-21 주석, Route·Agent Contract, S13, S19, research CD-2 |
| F-001 S19(d) dispatch 0이 route와 모순 | minor | dispatch 정확히 1번(reverse-engineer)이 미등록 문장으로 끝나고 executor·tester dispatch와 새 파일 0 | S19 (d) |
| J-14 behaviour-spec commit 범위 | minor | 로컬 전용으로 결정: 생성 `.gitignore`(`internal/cli/init_helpers.go:68`)에 `.autopus/reverse/`, `rootTrackedPrefixes`는 그대로, S3·S19가 확인, 규칙 9 | Trust Model, Route Contract 4, T18, S3, S19, Open Issues |

| Checklist FAIL | 반영 |
|---|---|
| Q-COMP-04 | S19 (e) 허용 경로 |
| Q-COMP-05 | S18 cold 증명, S19 동의 호출 수, S19(d) dispatch 1회 |
| Q-FEAS-02 | `.autopus/reverse/` 로컬 전용과 T18·규칙 9 |
| Q-FEAS-03 | S18 절차와 S19(d) oracle |

## Gate 1 (2026-10-15, 사용자 결정)

| 항목 | 결정 | 반영 |
|---|---|---|
| Q1 `features.rea` lazy 쓰기 | 유지 | 변경 없음, Open Issues resolved |
| J-03 JSON 값 보존·Codex bytes | 유지 | 변경 없음 |
| J-14 `.autopus/reverse/` 로컬 전용, `--runtime`, tester Scaffold 먼저 | 유지 | 변경 없음 |
| J-01 OMP balanced analysis rung | 변경: fail closed 대신 "세션 모델로 대체" | 새 analysis rung의 model이 catalog에 없으면 reverse-engineer만 override key 없음(세션 model 상속, `.omp/agents/reverse-engineer.md`는 `model:` 없음), 대체 줄 1개, 실행 성공. 기존 rung과 operator pin은 `omp_profile_candidate_unavailable`·pin 증명 오류 그대로. REQ-REA-11·19, Preset Contract, Form Text, 규칙 7, T23, S11(K1–K4), S21, INV-REA-13, Open Issues |

## 보안 감사 M1·M2 (2026-10-16, lead 결정 M2)

| 항목 | 결정 | 반영 |
|---|---|---|
| M2 subagent의 runtime capture 도구 | subagent는 갖지 않는다 | Reverse Engineer Agent Contract: Claude `mcp__rea__<정적 도구>`, Antigravity `mcp_rea_<정적 도구>` 135개(분류 `evidence/rea-tool-catalog-6.5.0.md`, 코드 `pkg/reamcp/tools.go`). Reverse Route Contract 2: 허용 뒤 메인 세션이 runtime 도구를 직접 부르고 evidence_id만 `Runtime capture evidence:` 줄로 넘긴다. REQ-REA-10·21, S9·S13·S19 |
| M1 Codex·OpenCode 도구 제한 | Codex 역할 파일이 `sandbox_mode`를 받으므로 reverse-engineer에만 `read-only`, OpenCode는 `task` 제외 | Agent Contract Codex·OpenCode 행, research R4(MCP 호출은 sandbox 밖이라 권고 수준 잔여 위험), CHANGELOG |
