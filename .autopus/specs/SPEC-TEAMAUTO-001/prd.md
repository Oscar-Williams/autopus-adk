# PRD: /auto go가 조건이 맞을 때 native Agent Team으로 자동 전환

- **SPEC-ID**: SPEC-TEAMAUTO-001
- **Mode**: Minimal (5 sections)
- **Source**: direct `/auto plan` pre-plan 대화의 ledger(사용자가 제안문을 그대로 승인, 2026-10-15), `33b5bb55` 코드, `evidence/probe-teamauto-capability.md`
- **Status**: Draft
- **Date**: 2026-10-15
- **Target module**: `autopus-adk`
- **Release**: v0.50.126(SPEC-MODELUX-003, SPEC-REA-001, PR #199와 함께), branch `feat/teamauto-001`

## 1. 문제

`/auto go`는 `--team`을 줘야만 native Agent Team(Route B)을 쓴다. flag 없는 기본 Route A는 서로 겹치지 않는 독립 slice가 여럿이어도 subagent로 돌린다. 팀 기능을 켜 둔 사용자도 run마다 `--team`을 기억해야 하고, 반대로 `--team`은 의존 관계가 있는 단계까지 팀 방식으로 바꾼다. 왜 팀을 썼는지, 왜 안 썼는지도 보이지 않는다.

## 2. 목표

- 두 조건이 모두 맞을 때만 Phase 2(Implementation)를 native Agent Team으로 자동 실행한다. (a) 세션이 native team을 실행할 수 있다(Claude Code 2.1.246 이상, `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`, 대화형 최상위 세션). (b) 함께 실행할 수 있고 owned paths가 겹치지 않는 준비된 slice가 2개 이상이다.
- 전환할 때도 안 할 때도 이유 한 줄(ko/en)을 정확히 한 번 보인다.
- `workflow.auto_team: false`로 계속, `--solo`로 한 run만 끈다. `--team`·`--workflow`·`--solo`는 지금 의미 그대로다.
- 판단은 Go 순수 함수라 조합 전체를 `go test`로 판정한다.

## 3. 사용자

- 팀 기능을 켠 Claude Code 사용자(주 사용자): flag 없이 독립 작업을 병렬 팀으로 얻는다.
- 팀 기능을 켜지 않았거나 Claude Code가 아닌 플랫폼(codex, antigravity-cli, opencode, omp)을 쓰는 사용자: 지금과 같다. Claude Code에서는 이유 한 줄만 더 본다.
- 비용을 아끼려는 사용자: teammate마다 별도 Claude 세션이라 토큰이 늘어난다는 공지를 보고 설정으로 끈다.

## 4. 범위 밖

- `/auto plan`, `/auto review`, `auto pipeline run`, Phase 2 밖 phase의 실행 방식. 의존 단계는 subagent 그대로다.
- 비 Claude 플랫폼의 생성 결과, 새 topology flag, topology flag 상호 배타 규칙.
- Route B 계약 밖의 팀 기능(팀 크기 정책, 팀원 model, pane, wave 재판단, 겹치는 slice 재분할), Workflow 기반층과의 결합.

## 5. 성공 기준

- 판단: 팀 불가 × 독립 slice 3개 → subagent + `teams_env_missing` 줄, 팀 가능 × 준비된 slice 1개 → subagent + `fewer_than_two_ready` 줄, 경로를 함께 쓰는 2개 → subagent + `ownership_overlap` 줄, 서로 다른 2개 → team + 전환 줄, 설정 off → `config_off` 줄, 명시 flag → 기존 경로(S1–S8).
- 비 Claude 플랫폼과 명시 flag 절의 생성 bytes가 기준과 같다(S6, S7, S9).
- `workflow.auto_team`은 바이너리가 스스로 쓰지 않아 끄지 않은 프로젝트는 auto 0.50.125에서도 그대로 읽힌다(S8).
- 실제 터미널 Claude Code에서 자동 전환 run 1회와 유지 run 1회를 기록한다(S15, CD-1·CD-2).
