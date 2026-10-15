# SPEC-TEAMAUTO-001 review rounds

## Round 1 (2026-10-15, verdict REVISE, claude + codex)

입력: `review.md`, `review-findings.json`(open 18, deferred 3). 새 근거는 `probe-teamauto-capability.md` Y(yaml.v3 decode)·R(root 오류, loader 쓰기)이다.

| Finding | 판정 | 처리 | 위치 |
|---|---|---|---|
| F-001, F-010 | 수용 | `config.Load`(정규화 때 autopus.yaml을 씀) 대신 `config.LoadPreview` + `os.Lstat`. 파일 없음은 `Lstat`가 정한다. S12에 `platforms: [claude]` fixture | spec Trust Model·Topology Facts Contract, plan T5, acceptance S12 |
| F-002, F-012 | 수용 | 규칙 5: shutdown·ack → `team_units` 밖 Phase 2 unit을 Route A로 → Gate 2. S13이 순서를, S15(a)가 실제 순서를 판정. INV-TA-07 | spec REQ-TA-02·Skill Surface Contract·flowchart, acceptance S13·S15 |
| F-003, F-018 | 수용 | stderr oracle은 빌드한 바이너리 기준 `Error: <msg>\n`, 종료 1(R1). in-process test는 반환 error 문자열 | acceptance 머리말·S6·S8·S11, spec REQ-TA-08 |
| F-004, F-017 | 수용 | `readWorkflowTriageFile`(guard만) 분리를 T5 소유로. 기존 오류 문구 유지, S11이 기준 바이너리와 stdout·stderr·종료 코드 비교 | spec Topology Facts Contract, plan T5, acceptance S11 |
| F-005 | 수용 | `claude_version_unknown` 줄에 `execpath=<basename>`. 문서에 EXECPATH 조건. S15(e)가 npm 설치를 기록 | spec Reason Line·REQ-TA-14, acceptance S5·S15·S16, plan A2 |
| F-006 | 수용 | REQ-TA-11을 "생성 입력·adapter code·golden 무변경 + auto-team 문자열 없음"으로, S9를 같은 강도로 | spec REQ-TA-11, acceptance S9 |
| F-011 | 수용 | owned path 끝 `/**` 하나를 떼고 `cleanPath`. 남은 glob은 `owned_paths` 오류. UGL·UGO fixture | spec Decision·Facts Contract, plan T2, acceptance S3·S4·S11 |
| F-013 | 수용 | 실행 확인(Y): `yes`·`"off"`·`null`이 통과한다. `decodeStrict`에서 raw node `!!bool` 검사, 오류 `workflow.auto_team: must be true or false (line <n>)` | spec REQ-TA-09·Config Contract, plan T6, acceptance S8 |
| F-014 | 수용 | 무변경 단정은 `<tmp>/proj`만, counter는 `<tmp>/probe`. REQ-TA-08은 "agent·teammate 없음, 쓰기 없음, child process는 버전 probe 하나" | spec REQ-TA-08·Trust Model, acceptance S12 |
| F-015 | 수용 | CLI 실패 줄은 생성 `sh` snippet이 출력하고 S14가 그 snippet을 실행한다. schema 실패는 S15(c)(`--disallowedTools SendMessage`), resume은 S15(d). CD-3 | spec REQ-TA-13·Skill Surface Contract, acceptance S14·S15 |
| F-016 | 수용 | 입력 한 값(`TeamInput`, locale 포함), 순수 `NeedsVersionProbe`·`DecideTeam`, 관측은 CLI | spec REQ-TA-04·06·Decision·Capability Contract, plan T1·T4·T5 |
| F-019 | 수용 | 옛 바이너리 거부는 값과 무관(`true` 포함). Migration Notes 추가 | spec Config Contract·Migration Notes, acceptance S8, prd |
| F-020 | 수용 | `auto spec validate .autopus/specs/SPEC-TEAMAUTO-001 --strict` | spec Outcome Boundary, research Outcome Lock |
| F-021 | 수용 | Self-Verify Summary에 체크리스트 24개 전부 | research Self-Verify Summary |
| F-007 (deferred) | 부분 반영 | `flags`·규칙 1·`explicit_flag`가 CLI 단독 호출용임을 계약에 명시. 표면은 유지(flag 차원 test) | spec Decision·Reason Line Contract |
| F-008 (deferred) | 반영 | timeout 행은 probe 1회와 20초 미만으로 | acceptance S5 |
| F-009 (deferred) | 반영 | `[NEW] firstOverlapPair`, 재사용은 판정식만 | spec Decision Contract, plan T1 |

체크리스트 FAIL 항목(Q-CORR-01, Q-CORR-03, Q-COMP-02, Q-COMP-04, Q-COMP-05, Q-FEAS-01, Q-FEAS-03)의 재판정은 `research.md` `## Self-Verify Summary`다.
