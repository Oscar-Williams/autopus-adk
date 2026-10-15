# SPEC-TEAMAUTO-001 수락 기준

## Test Scenarios

- 디렉터리: `<tmp>/proj`는 명령을 돌리는 project(autopus.yaml), `<tmp>/probe`는 fake EXECPATH script와 counter, `<tmp>/facts`는 facts 파일이다. 파일 무변경 단정은 `<tmp>/proj`만 본다.
- 바이너리: branch `go build -o <tmp>/bin/auto-branch ./cmd/auto`, 기준 `33b5bb55`를 `git worktree add`로 꺼내 같은 방식으로 빌드한 `<tmp>/bin/auto-base`, PATH의 `auto 0.50.125`. stderr·종료 코드 단정은 빌드한 바이너리 기준이다. root는 오류를 `Error: <msg>\n`으로 쓰고 일반 오류의 종료 코드는 1이다(`internal/cli/root.go:203-211`, `:173-179`, evidence R1). in-process `NewRootCmd` test는 같은 `<msg>`를 반환 error로 단정한다.
- CAP-OK: `CLAUDECODE=1`, `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`, `CLAUDE_CODE_ENTRYPOINT=cli`, `CLAUDE_CODE_SESSION_ATTENDED=1`, `CLAUDE_CODE_CHILD_SESSION` 없음, `CLAUDE_CODE_EXECPATH=<tmp>/probe/claude`. 그 파일은 `2.1.288 (Claude Code)`를 출력하고 `<tmp>/probe/probe.count`에 한 줄을 더하는 실행 script다. PATH 맨 앞 `<tmp>/probe/other/claude`는 `9.9.9 (Claude Code)`를 출력하고 `<tmp>/probe/path.touched`를 만든다. CAP-NOENV: CAP-OK에서 `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS`만 뺀 것.
- U3: T1 worker `[pkg/alpha]`, T2 worker `[internal/beta]`, T3 worker `[docs/gamma.md]`, 의존·done 없음. U1: T1 worker `[pkg/alpha]`, T2 worker `[pkg/beta]` depends_on `[T1]`, T3 inline `[README.md]`. UOV: T1 worker `[pkg/alpha]`, T2 worker `[pkg/alpha/b.go]`. UDJ: T1 worker `[pkg/a]`, T2 worker `[pkg/ab]`, T3 worker `[pkg/c]` depends_on `[T1]`, T4 worker `[pkg/d]` done. UGL: T1 worker `[pkg/a/**]`, T2 worker `[pkg/ab/**]`. UFU: T4 worker `[pkg/d]` done, T5 worker `[pkg/e]`, T6 worker `[pkg/f]`, T7 worker `[pkg/g]` depends_on `[T5]`, T8 inline `[docs/h.md]`. UGO: T1 worker `[pkg/alpha/**]`, T2 worker `[pkg/alpha/b.go]`.
- CFG-KO: `platforms: [claude-code]`, `language.ai_responses: ko`, `workflow.auto_team` 없음. CFG-EN·CFG-JA: 언어만 `en`·`ja`. CFG-OFF: CFG-KO + `workflow.auto_team: false`. CFG-LEG: CFG-KO인데 `platforms: [claude]`(옛 이름, `Load`라면 다시 쓰는 입력).
- 따로 적지 않으면 flag는 모두 false, config는 CFG-KO, 명령은 `<tmp>/proj`에서 `auto-branch workflow triage --topology --facts-json <tmp>/facts/f.json --format json`이다.

### S1: 팀을 쓸 수 없는 세션 × 독립 slice 3개
Priority: Must
Given CAP-NOENV와 U3가 있다
When 명령을 실행한다
Then 종료 0이고 decision은 `route: subagent`, `reason: teams_env_missing`, `ready_units: 3`이며 `team_units`와 `tools_fallback_line`이 없다
And `line`은 `Agent Team: 유지 — 세션 환경에 CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1이 없어 subagent로 실행한다`다
And `<tmp>/probe/probe.count`와 `<tmp>/probe/path.touched`가 없다

### S2: 팀 가능 × 준비된 slice 1개
Priority: Must
Given CAP-OK와 U1, U1에서 T1을 done으로 바꾼 U1d, T1·T2를 done으로 바꾼 U1dd가 있다
When 세 facts로 명령을 실행한다
Then 셋 모두 `route: subagent`, `reason: fewer_than_two_ready`이고 `ready_units`는 1, 1, 0이다(U1d의 ready는 T2다)
And U1의 `line`은 `Agent Team: 유지 — 함께 실행할 준비된 slice가 1개라 subagent로 실행한다`, U1dd는 `Agent Team: 유지 — 함께 실행할 준비된 slice가 0개라 subagent로 실행한다`다
And `<tmp>/probe/probe.count`가 없다

### S3: 팀 가능 × 경로를 함께 쓰는 slice 2개
Priority: Must
Given CAP-OK와 UOV, UGO, 그리고 U3에 T4 worker `[internal/beta/x]`를 더한 U3o가 있다
When 세 facts로 명령을 실행한다
Then 셋 모두 `route: subagent`, `reason: ownership_overlap`이다
And UOV·UGO의 `line`은 `Agent Team: 유지 — T1, T2의 owned_paths가 겹쳐 subagent로 실행한다`, U3o는 `Agent Team: 유지 — T2, T4의 owned_paths가 겹쳐 subagent로 실행한다`다
And `<tmp>/probe/probe.count`가 없다

### S4: 팀 가능 × 서로 다른 경로의 slice 2개
Priority: Must
Given CAP-OK와 UDJ(`pkg/a`와 `pkg/ab`는 겹치지 않고 T3은 T1을 기다리며 T4는 done), UGL이 있다
When UDJ로 두 번, UGL로 한 번 명령을 실행한다
Then 세 run 모두 `route: team`, `reason: independent_disjoint_slices`, `ready_units: 2`, `team_units: ["T1", "T2"]`다
And `line`은 `Agent Team: 전환 — 준비된 slice 2개의 owned_paths가 겹치지 않아 native Agent Team으로 실행한다`, `tools_fallback_line`은 `Agent Team: 유지 — SendMessage·task 도구 schema를 불러오지 못해 subagent로 실행한다`다
And UDJ의 `followup_units`는 `["T3"]`(T4는 done이라 없다), UGL은 `[]`다
And `<tmp>/probe/probe.count`는 run마다 한 줄씩 3줄이고 `<tmp>/probe/path.touched`가 없으며 UDJ 두 run의 stdout bytes가 같다

### S4b: 팀 wave 뒤 Route A로 넘길 unit 집합
Priority: Must
Given CAP-OK와 UFU가 있다
When 명령을 실행한다
Then decision은 `route: team`, `ready_units: 2`, `team_units: ["T5", "T6"]`, `followup_units: ["T7", "T8"]`이다
And done인 T4는 `team_units`에도 `followup_units`에도 없다
And UFU에서 T7을 done으로 바꾸면 `followup_units`는 `["T8"]`, T7·T8을 모두 done으로 바꾸면 `[]`다

### S5: 능력 신호 하나씩 바꾸기
Priority: Must
Given U3와, CAP-OK에서 아래 한 가지씩만 바꾼 환경이 있다
When 각 환경에서 명령을 실행한다
Then 결과는 아래 표와 같다
And `CLAUDECODE`와 `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS`를 함께 뺀 환경은 `not_claude_code`다(규칙 5가 6보다 먼저다)

| 바꾼 것 | reason | line 또는 관측 |
|---|---|---|
| `CLAUDE_CODE_ENTRYPOINT=sdk-cli` | `not_interactive` | `Agent Team: 유지 — 대화형 CLI 세션이 아니라(entrypoint=sdk-cli, attended=1) subagent로 실행한다` |
| `CLAUDE_CODE_SESSION_ATTENDED` 없음 | `not_interactive` | `Agent Team: 유지 — 대화형 CLI 세션이 아니라(entrypoint=cli, attended=unset) subagent로 실행한다` |
| `CLAUDE_CODE_ENTRYPOINT=cli; rm -rf /` | `not_interactive` | `Agent Team: 유지 — 대화형 CLI 세션이 아니라(entrypoint=unknown, attended=1) subagent로 실행한다` |
| `CLAUDE_CODE_CHILD_SESSION=1` | `not_top_level` | `Agent Team: 유지 — 하위 세션은 팀 lead가 될 수 없어 subagent로 실행한다` |
| `CLAUDECODE` 없음 | `not_claude_code` | `Agent Team: 유지 — Claude Code 세션이 아니라 subagent로 실행한다` |
| script가 `2.1.245 (Claude Code)` 출력 | `claude_version_old` | `Agent Team: 유지 — Claude Code 2.1.245 버전은 2.1.246보다 낮아 subagent로 실행한다` |
| script가 `2.1.246 (Claude Code)` 출력 | `independent_disjoint_slices` | `route: team`, `ready_units: 3` |
| script가 `2.10.0 (Claude Code)` 출력 | `independent_disjoint_slices` | `route: team` |
| EXECPATH가 `v24.20.0`을 출력하는 `<tmp>/probe/node` | `claude_version_unknown` | `Agent Team: 유지 — 실행 중인 Claude Code 버전을 읽지 못해(execpath=node) subagent로 실행한다`, probe.count 없음(실행 안 함) |
| EXECPATH가 상대 경로 `probe/claude` | `claude_version_unknown` | `…(execpath=claude)…`, probe.count 없음 |
| script가 `probe.count`에 쓴 뒤 30초 sleep | `claude_version_unknown` | `…(execpath=claude)…`, probe.count 1줄, 명령이 20초 안에 끝남 |
| `CLAUDE_CODE_EXECPATH` 없음 | `claude_version_unknown` | `Agent Team: 유지 — 실행 중인 Claude Code 버전을 읽지 못해(execpath=unset) subagent로 실행한다` |

### S6: --solo와 --workflow
Priority: Must
Given CAP-OK와 UDJ, `flags.solo: true`인 facts, `flags.workflow: true`인 facts, `flags.solo`와 `flags.team`이 모두 true인 facts가 있다
When 세 facts로 명령을 실행하고 branch와 `33b5bb55`에서 claude-code `Generate`를 돌린다
Then 앞의 둘은 `route: solo`·`route: workflow`, `reason: explicit_flag`, `ready_units` 없음, `line`은 `Agent Team: 자동 선택 안 함 — --solo 지정됨`·`Agent Team: 자동 선택 안 함 — --workflow 지정됨`이고 `<tmp>/probe/probe.count`가 없다
And 셋째는 종료 1, stdout은 비어 있고 stderr bytes는 `Error: invalid topology facts: topology_conflict\n`이다
And 생성 `.claude/skills/auto-go/SKILL.md`의 `### Step 0` 절(`[TOPOLOGY ERROR] --team, --workflow, and --solo are mutually exclusive; choose only one.` 포함), `#### Route C` 절, `#### Route: Deterministic Workflow` 절은 기준과 bytes가 같다

### S7: 팀을 쓸 수 없을 때 --team은 기존 blocker
Priority: Must
Given CAP-NOENV와 UDJ, `flags.team: true`가 있다
When 명령을 실행하고 S6처럼 두 생성 결과를 비교한다
Then decision은 `route: team`, `reason: explicit_flag`, `line`은 `Agent Team: 자동 선택 안 함 — --team 지정됨`이고 능력을 읽지 않아 `<tmp>/probe/probe.count`가 없다
And 생성 skill의 `#### Route B` 절(다음 `#### Route C` 앞까지)은 기준과 bytes가 같고 ``2. `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` is present in the current session.``과 `If preflight fails, report the blocker and suggest Route A.`를 담는다
And 새 자동 선택 절의 첫 규칙은 `TOPOLOGY_FLAG_COUNT == 0`이다

### S8: 설정 key 값과 옛 바이너리
Priority: Must
Given CAP-OK와 UDJ, 그리고 CFG-KO의 `workflow:` 아래 `auto_team: <v>`만 바꾼 project들이 있다
When 각 project에서 명령, branch `auto-branch update --yes`, v0.50.125 `auto update --yes`를 돌리고 빈 project에서 `CI=true auto-branch init --yes --platforms claude-code`를 돌린다
Then `<v>`가 `false`·`FALSE`면 `route: subagent`, `reason: config_off`, `line` `Agent Team: 유지 — autopus.yaml의 workflow.auto_team이 false라 subagent로 실행한다`, probe.count 없음이고, `true`·`!!bool "true"`·key 없음은 `route: team`이다
And `<v>`가 `"false"`, `null`, `~`, 빈 값, `yes`, `"off"`, `1`이면 종료 1, stdout은 비어 있고 stderr는 `Error: config load failed: `로 시작하며 `workflow.auto_team: must be true or false (line <n>)`(`<n>`은 그 key의 줄)를 담는다
And `auto-branch update --yes` 뒤 `false`·`true` project의 `workflow:` 블록에는 각각 `auto_team: false`·`auto_team: true`가 남고 init 결과에는 `auto_team`이 없다(`grep -c auto_team autopus.yaml`이 0)
And v0.50.125 `auto update --yes`는 `false`·`true` project에서 모두 종료 1과 `field auto_team not found in type config.WorkflowConf`, init 결과에서 종료 0이다

### S9: 비 Claude 플랫폼은 그대로
Priority: Must
Given branch와 `33b5bb55`가 있다
When `git diff 33b5bb55 --name-only`를 보고, codex·antigravity·opencode·omp adapter test를 돌리고, branch에서 빈 project에 그 네 플랫폼 `Generate`를 돌린다
Then `-- templates content`의 변경은 `templates/claude/commands/auto-workflows.md.tmpl`과 `content/skills/agent-teams.md` 둘뿐이다
And `-- pkg/adapter/codex pkg/adapter/antigravity pkg/adapter/opencode pkg/adapter/omp`의 변경은 없고 `go test ./pkg/adapter/codex/... ./pkg/adapter/antigravity/... ./pkg/adapter/opencode/... ./pkg/adapter/omp/...`가 통과한다
And 네 플랫폼 생성 파일에는 `--topology`, `Agent Team Auto-Selection`, `Automatic Entry from Route A`, `auto_team`이 하나도 없다

### S10: ko·en 문구
Priority: Must
Given CAP-OK와 UDJ, U1이 있다
When CFG-EN, CFG-JA, autopus.yaml 없음, CFG-EN + `workflow.auto_team: false`에서 명령을 실행한다
Then CFG-EN·CFG-JA·파일 없음의 UDJ `line`은 모두 `Agent Team: on — 2 ready slices have disjoint owned_paths; running them as a native Agent Team`이고 `tools_fallback_line`은 `Agent Team: off — SendMessage/task tool schemas could not be loaded; running subagents`다
And 파일 없음은 `config_source: default`, CFG-EN의 U1은 `Agent Team: off — 1 ready slice(s) can run together; running subagents`, 끈 CFG-EN은 `Agent Team: off — workflow.auto_team is false in autopus.yaml; running subagents`다

### S11: facts 검증과 기존 triage 회귀
Priority: Must
Given 아래 잘못된 facts, `internal/cli/workflow_triage_test.go`의 기존 facts fixture, 디렉터리 경로, 1 MiB + 1 byte 파일, `{` 한 글자 파일이 있다
When 잘못된 facts로 명령을 실행하고, 나머지 넷으로 `--topology` 없이 `auto-branch`와 `auto-base`를 실행하고, UDJ facts로 `--topology` 없이 `auto-branch`를 실행한다
Then 잘못된 facts는 모두 종료 1, stdout은 비어 있고 stderr bytes는 표와 같다
And 나머지 넷의 stdout·stderr bytes와 종료 코드는 두 바이너리에서 같고, UDJ facts를 `--topology` 없이 주면 종료 1과 stderr `Error: invalid workflow triage facts JSON\n`이다

| facts | stderr |
|---|---|
| `version: 2` | `Error: invalid topology facts: version\n` |
| T1 depends_on `["T9"]`, T1 depends_on `["T1"]` | `Error: invalid topology facts: depends_on\n` |
| unit에 `"id"` key 두 번, 최상위 `"extra": 1` | `Error: invalid topology facts: json\n` |
| worker `owned_paths: []` | `Error: invalid topology facts: owned_paths\n` |
| `["../x"]`, `["/abs"]`, `["pkg/*/x"]`, `["pkg/a/**/**"]`, `["./**"]` | `Error: invalid topology facts: owned_paths\n` |
| `execution: "parallel"` | `Error: invalid topology facts: execution\n` |
| unit 33개 | `Error: invalid topology facts: unit_count\n` |
| id `T 1` | `Error: invalid topology facts: unit_id\n` |

### S12: 출력 형식과 쓰기 없음
Priority: Must
Given CAP-OK, CFG-LEG(`platforms: [claude]`), UDJ, 실행 전 `<tmp>/proj`의 파일 목록과 파일마다 SHA-256이 있다
When `--format json`과 `--format human`으로 한 번씩 실행한다
Then human stdout은 정확히 `Agent Team: 전환 — 준비된 slice 2개의 owned_paths가 겹치지 않아 native Agent Team으로 실행한다\n`이다
And JSON은 `mode: topology`, `unit_source: caller_declared`, `capability_source: observed_process_env`, `config_source: autopus.yaml`, `execution_performed: false`를 담는다
And 실행 뒤 `<tmp>/proj`의 파일 목록과 SHA-256이 실행 전과 같고(autopus.yaml의 `claude`가 다시 쓰이지 않았다) `<tmp>/probe/probe.count`는 2줄, `<tmp>/probe/path.touched`는 없다

### S13: 생성 skill 문구
Priority: Must
Given branch에서 claude-code `Generate`를 돌린 결과와 `33b5bb55` 결과가 있다
When `.claude/skills/auto-go/SKILL.md`와 `.claude/skills/agent-teams/SKILL.md`를 읽는다
Then auto-go의 `#### Route A: Supervisor-Led Pipeline`과 `#### Route B: Native Agent Team` 사이에 `#### Route A Phase 2 Entry: Agent Team Auto-Selection`이 있고, 그 절에 `TOPOLOGY_FLAG_COUNT == 0`, `--continue` 건너뛰기 문장, Skill Surface Contract의 `sh` snippet 다섯 줄 그대로, `TEAM_JSON` 빈 경우 규칙, `select:SendMessage,TaskCreate,TaskUpdate,TaskList,TaskGet`, `tools_fallback_line`이 이 순서로 나온다
And 그 뒤에 shutdown·ack 문장, `team_units`에 없는 Phase 2 unit을 Route A로 구현하는 문장, Gate 2로 가는 문장이 이 순서로 나온다
And 기준과의 diff는 추가 줄뿐이고 모두 그 절 안이며, agent-teams skill에는 `## Automatic Entry from Route A`가 있고 `--topology`는 없다

### S14: 판단 도구 실패 시 생성 snippet
Priority: Must
Given S13 절의 `sh` 블록을 꺼낸 파일, `<tmp>/proj`, UDJ facts를 복사한 `FACTS`와 빈 `TEAM_JSON` 임시 파일이 있다
When `FACTS`·`TEAM_JSON`을 env로 주고 `sh`로 그 블록을 PATH의 `auto`만 바꿔 다섯 번 실행한다: (i) 종료 1 fake, (ii) 종료 0이고 출력 없는 fake, (iii) v0.50.125(로컬. CI에서는 stderr `Error: unknown flag: --topology\n`과 종료 1을 내는 fake), (iv) CAP-OK의 `auto-branch`, (v) `workflow: [`로 깨진 autopus.yaml에서 `auto-branch`
Then (i)(ii)(iii)(v)의 stdout은 정확히 `Agent Team: off — auto workflow triage --topology failed; running subagents\n`, stderr는 비어 있고 `TEAM_JSON`은 0 bytes다
And (iv)의 stdout과 stderr는 비어 있고 `TEAM_JSON`은 `decision.route: team`인 JSON이며, 다섯 run 모두 `FACTS`가 지워졌다

### S15: 실제 터미널 확인
Priority: Must
Given 실제 터미널의 대화형 최상위 `claude`(2.1.246 이상) 세션에 `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`이 있고, scratch project에 SPEC-A(disjoint ready worker T1·T2와 T1에 의존하는 T3)와 SPEC-B(ready worker 1개)가 있다
When (a) SPEC-A, (b) SPEC-B로 `/auto go SPEC-ID --auto`를, (c) `claude --disallowedTools SendMessage` 세션에서 SPEC-A를, (d) (a)를 Gate 2에서 멈춘 뒤 `/auto go SPEC-A --continue`를 돌리고, (e) 같은 세션 Bash와 `claude -p`에서 `CLAUDE_CODE_ENTRYPOINT`·`CLAUDE_CODE_SESSION_ATTENDED`·`CLAUDE_CODE_CHILD_SESSION`과 npm 설치가 있으면 그 세션의 EXECPATH basename을 기록한다
Then (a) transcript의 `Agent Team:` 줄은 `Agent Team: 전환 — 준비된 slice 2개의 owned_paths가 겹치지 않아 native Agent Team으로 실행한다` 하나이고, runtime이 teammate로 보고한 `builder-1`·`builder-2`, shutdown ack, `followup_units`인 T3의 파일 변경(T1·T2는 다시 돌지 않음), Gate 2 시작이 이 순서로 나온다(CD-1)
And (b)는 `Agent Team: 유지 — 함께 실행할 준비된 slice가 1개라 subagent로 실행한다` 한 줄뿐이고 named teammate가 없다(CD-2)
And (c)는 `Agent Team: 유지 — SendMessage·task 도구 schema를 불러오지 못해 subagent로 실행한다` 한 줄뿐이고 named teammate가 없으며, (d)의 `Agent Team:` 줄은 0개다(CD-3)
And (e)의 최상위 세션은 `cli`, `1`, CHILD 없음이고 `claude -p`는 (`cli`, `1`) 쌍이 아니다. 쌍이 같으면 A2는 FAIL이고 규칙 7을 고친 뒤에만 릴리스한다. 기록은 `<home>`·scratch 경로를 바꿔 `evidence/manual-smoke-teamauto.txt`에 있다

### S16: 문서
Priority: Should
Given branch README, CHANGELOG, `auto workflow triage --help`가 있다
When 읽는다
Then README 설정 표에 `workflow.auto_team`(기본: 없음 = on) 행이 있고 CHANGELOG v0.50.126 항목은 기본 on, teammate마다 별도 Claude 세션(토큰 증가), 실험 기능, 끄는 법, 값과 무관한 auto 0.50.113–0.50.125의 거부와 auto 0.50.112 이하가 key를 무시하고 `auto update`에서 지우는 점, `CLAUDE_CODE_EXECPATH`가 Claude Code 바이너리가 아닌 설치의 유지를 적는다
And help에 `--topology`가 있다

## Oracle Acceptance Notes

- S1–S12, S14의 oracle은 route·reason·line 문자열, 수, stderr bytes, 종료 코드, SHA-256, snippet 실행 결과다. S13은 생성 문구의 정확한 문자열과 순서, diff 범위다. S15는 실제 터미널 기록이다.
- 실행 invariant는 실행 oracle에 묶는다. INV-TA-01: S1, S2, S3, S4, S12, S14, S15. INV-TA-02: S2, S3, S4, S4b, S15. INV-TA-03: S4, S5, S15. INV-TA-04: S8. INV-TA-05: S6, S7, S9. INV-TA-06: S10. INV-TA-07: S4b, S15. S13은 문구 보조 확인이다.
