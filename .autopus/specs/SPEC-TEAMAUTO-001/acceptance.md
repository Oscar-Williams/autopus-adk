# SPEC-TEAMAUTO-001 수락 기준

## Test Scenarios

판단·CLI test는 실제 명령(`NewRootCmd`)을 temp dir에서 돌리고 env는 `t.Setenv`로 준다. 순수 함수 test는 env 조회와 버전 probe를 주입한다. 기준 바이너리는 PATH의 `auto 0.50.125`, 기준 생성은 origin/main `33b5bb55`의 adapter `Generate`다. 실제 터미널은 S15에서만 쓴다.

- CAP-OK: `CLAUDECODE=1`, `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`, `CLAUDE_CODE_ENTRYPOINT=cli`, `CLAUDE_CODE_SESSION_ATTENDED=1`, `CLAUDE_CODE_CHILD_SESSION` 없음, `CLAUDE_CODE_EXECPATH=<tmp>/bin/claude`. 그 파일은 `2.1.288 (Claude Code)`를 출력하고 `<tmp>/probe.count`에 한 줄을 더하는 실행 script다. PATH 맨 앞 `<tmp>/other/claude`는 `9.9.9 (Claude Code)`를 출력하고 `<tmp>/path.touched`를 만든다.
- CAP-NOENV: CAP-OK에서 `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS`만 뺀 것.
- U3: T1 worker `[pkg/alpha]`, T2 worker `[internal/beta]`, T3 worker `[docs/gamma.md]`. 의존 없음, done 없음.
- U1: T1 worker `[pkg/alpha]`, T2 worker `[pkg/beta]` depends_on `[T1]`, T3 inline `[README.md]`.
- UOV: T1 worker `[pkg/alpha]`, T2 worker `[pkg/alpha/b.go]`.
- UDJ: T1 worker `[pkg/a]`, T2 worker `[pkg/ab]`, T3 worker `[pkg/c]` depends_on `[T1]`, T4 worker `[pkg/d]` done.
- CFG-KO: `language.ai_responses: ko`, `workflow.auto_team` 없음. CFG-EN·CFG-JA: 같은데 `en`·`ja`. CFG-OFF: CFG-KO + `workflow.auto_team: false`.
- 따로 적지 않으면 flag는 모두 false, config는 CFG-KO, 명령은 `auto workflow triage --topology --facts-json <file> --format json`이다.

### S1: 팀을 쓸 수 없는 세션 × 독립 slice 3개
Priority: Must
Given CAP-NOENV와 U3가 있다
When 명령을 실행한다
Then 종료 0이고 decision은 `route: subagent`, `reason: teams_env_missing`, `ready_units: 3`이며 `team_units`와 `tools_fallback_line`이 없다
And `line`은 `Agent Team: 유지 — 세션 환경에 CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1이 없어 subagent로 실행한다`다
And `<tmp>/probe.count`와 `<tmp>/path.touched`가 없다

### S2: 팀 가능 × 준비된 slice 1개
Priority: Must
Given CAP-OK와 U1, U1에서 T1을 done으로 바꾼 U1d, U1에서 T1·T2를 done으로 바꾼 U1dd가 있다
When 세 facts로 명령을 실행한다
Then 셋 모두 `route: subagent`, `reason: fewer_than_two_ready`이고 `ready_units`는 1, 1, 0이다(U1d의 ready는 T2다)
And U1의 `line`은 `Agent Team: 유지 — 함께 실행할 준비된 slice가 1개라 subagent로 실행한다`, U1dd는 `… slice가 0개라 …`다
And `<tmp>/probe.count`가 없다

### S3: 팀 가능 × 경로를 함께 쓰는 slice 2개
Priority: Must
Given CAP-OK와 UOV, 그리고 U3에 T4 worker `[internal/beta/x]`를 더한 U3o가 있다
When 두 facts로 명령을 실행한다
Then 둘 다 `route: subagent`, `reason: ownership_overlap`이다
And UOV의 `line`은 `Agent Team: 유지 — T1, T2의 owned_paths가 겹쳐 subagent로 실행한다`, U3o는 `Agent Team: 유지 — T2, T4의 owned_paths가 겹쳐 subagent로 실행한다`다
And `<tmp>/probe.count`가 없다

### S4: 팀 가능 × 서로 다른 경로의 slice 2개
Priority: Must
Given CAP-OK와 UDJ가 있다(`pkg/a`와 `pkg/ab`는 겹치지 않고 T3은 T1을 기다리며 T4는 done이다)
When 명령을 두 번 실행한다
Then decision은 `route: team`, `reason: independent_disjoint_slices`, `ready_units: 2`, `team_units: ["T1", "T2"]`다
And `line`은 `Agent Team: 전환 — 준비된 slice 2개의 owned_paths가 겹치지 않아 native Agent Team으로 실행한다`, `tools_fallback_line`은 `Agent Team: 유지 — SendMessage·task 도구 schema를 불러오지 못해 subagent로 실행한다`다
And `<tmp>/probe.count`는 run마다 한 줄씩 2줄이고 `<tmp>/path.touched`가 없으며 두 run의 stdout bytes가 같다

### S5: 능력 신호 하나씩 바꾸기
Priority: Must
Given U3와, CAP-OK에서 아래 한 가지씩만 바꾼 환경이 있다
When 각 환경에서 명령을 실행한다
Then 결과는 아래 표와 같다
And `CLAUDECODE`와 `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS`를 함께 뺀 환경은 `not_claude_code`다(규칙 5가 6보다 먼저다)

| 바꾼 것 | reason | line 또는 관측 |
|---|---|---|
| `CLAUDE_CODE_ENTRYPOINT=sdk-cli` | `not_interactive` | `Agent Team: 유지 — 대화형 CLI 세션이 아니라(entrypoint=sdk-cli, attended=1) subagent로 실행한다` |
| `CLAUDE_CODE_SESSION_ATTENDED` 없음 | `not_interactive` | `…(entrypoint=cli, attended=unset)…` |
| `CLAUDE_CODE_ENTRYPOINT=cli; rm -rf /` | `not_interactive` | `…(entrypoint=unknown, attended=1)…` |
| `CLAUDE_CODE_CHILD_SESSION=1` | `not_top_level` | `Agent Team: 유지 — 하위 세션은 팀 lead가 될 수 없어 subagent로 실행한다` |
| `CLAUDECODE` 없음 | `not_claude_code` | `Agent Team: 유지 — Claude Code 세션이 아니라 subagent로 실행한다` |
| script가 `2.1.245 (Claude Code)` 출력 | `claude_version_old` | `Agent Team: 유지 — Claude Code 2.1.245 버전은 2.1.246보다 낮아 subagent로 실행한다` |
| script가 `2.1.246 (Claude Code)` 출력 | `independent_disjoint_slices` | `route: team`, `ready_units: 3` |
| script가 `2.10.0 (Claude Code)` 출력 | `independent_disjoint_slices` | `route: team` |
| EXECPATH가 `v24.20.0`을 출력하는 `<tmp>/bin/node` | `claude_version_unknown` | `<tmp>/probe.count` 없음(실행 안 함) |
| EXECPATH가 상대 경로 `bin/claude` | `claude_version_unknown` | 실행 안 함 |
| script가 10초 sleep | `claude_version_unknown` | 명령이 7초 안에 끝남 |
| `CLAUDE_CODE_EXECPATH` 없음 | `claude_version_unknown` | `Agent Team: 유지 — 실행 중인 Claude Code 버전을 읽지 못해 subagent로 실행한다` |

### S6: --solo와 --workflow
Priority: Must
Given CAP-OK와 UDJ, `flags.solo: true`인 facts, `flags.workflow: true`인 facts, `flags.solo`와 `flags.team`이 모두 true인 facts가 있다
When 세 facts로 명령을 실행하고 branch와 `33b5bb55`에서 claude-code `Generate`를 돌린다
Then 앞의 둘은 `route: solo`·`route: workflow`, `reason: explicit_flag`, `ready_units` 없음, `line`은 `Agent Team: 자동 선택 안 함 — --solo 지정됨`·`Agent Team: 자동 선택 안 함 — --workflow 지정됨`이고 `<tmp>/probe.count`가 없다
And 셋째는 종료 1, stderr `invalid topology facts: topology_conflict`, stdout은 비어 있다
And 생성 `.claude/skills/auto-go/SKILL.md`의 `### Step 0` 절(`[TOPOLOGY ERROR] --team, --workflow, and --solo are mutually exclusive; choose only one.` 포함), `#### Route C` 절, `#### Route: Deterministic Workflow` 절은 기준과 bytes가 같다

### S7: 팀을 쓸 수 없을 때 --team은 기존 blocker
Priority: Must
Given CAP-NOENV와 UDJ, `flags.team: true`가 있다
When 명령을 실행하고 S6처럼 두 생성 결과를 비교한다
Then decision은 `route: team`, `reason: explicit_flag`, `line`은 `Agent Team: 자동 선택 안 함 — --team 지정됨`이고 능력을 읽지 않아 `<tmp>/probe.count`가 없다
And 생성 skill의 `#### Route B` 절(다음 `#### Route C` 앞까지)은 기준과 bytes가 같고 ``2. `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` is present in the current session.``과 `If preflight fails, report the blocker and suggest Route A.`를 담는다
And 새 자동 선택 절의 첫 조건은 `TOPOLOGY_FLAG_COUNT == 0`이다

### S8: 설정 key와 옛 바이너리
Priority: Must
Given CAP-OK와 UDJ, 그리고 CFG-OFF, CFG-KO + `workflow.auto_team: true`, CFG-KO + `workflow.auto_team: "false"`인 세 project가 있다
When 각 project에서 명령, branch `auto update --yes`, v0.50.125 `auto update --yes`를 돌리고 빈 project에서 branch `CI=true auto init --yes --platforms claude-code`를 돌린다
Then CFG-OFF는 `route: subagent`, `reason: config_off`, `line` `Agent Team: 유지 — autopus.yaml의 workflow.auto_team이 false라 subagent로 실행한다`이고 `<tmp>/probe.count`가 없다. `true`는 `route: team`이다
And `"false"`는 종료 1, stderr `config load failed`이고 `config.Load` 오류는 그 줄 번호와 `cannot unmarshal !!str`을 담는다
And branch `auto update --yes` 뒤 CFG-OFF의 `workflow:` 블록에는 `auto_team: false`가 남고 `AutoTeamEnabled()`는 false다
And v0.50.125 `auto update --yes`는 CFG-OFF에서 종료 1과 `field auto_team not found in type config.WorkflowConf`, branch init 결과(`grep -c auto_team autopus.yaml`이 0)에서 종료 0이다

### S9: 비 Claude 플랫폼은 그대로
Priority: Must
Given 빈 project와 같은 config가 있고 build version·시각은 두 쪽에 같은 test seam 값으로 준다(그런 seam이 없는 필드는 T9가 목록으로 남기고 그 필드만 빼고 비교한다)
When `33b5bb55`와 branch에서 codex, antigravity-cli, opencode, omp `Generate`를 돌리고 branch에서 다섯 플랫폼을 함께 생성한다
Then 네 플랫폼의 생성 파일 집합이 같고 파일마다 SHA-256이 같다
And `git diff 33b5bb55 --name-only -- templates content`는 `templates/claude/commands/auto-workflows.md.tmpl`과 `content/skills/agent-teams.md`만 보인다
And 다섯 플랫폼 생성에서 `auto workflow triage --topology`를 담은 파일은 모두 `.claude/` 아래이고 `.claude/skills/auto-go/SKILL.md`를 포함한다

### S10: ko·en 문구
Priority: Must
Given CAP-OK와 UDJ, U1이 있다
When CFG-EN, CFG-JA, autopus.yaml 없음, CFG-EN + `workflow.auto_team: false`에서 명령을 실행한다
Then CFG-EN·CFG-JA·파일 없음의 UDJ `line`은 모두 `Agent Team: on — 2 ready slices have disjoint owned_paths; running them as a native Agent Team`이고 `tools_fallback_line`은 `Agent Team: off — SendMessage/task tool schemas could not be loaded; running subagents`다
And 파일 없음은 `config_source: default`, CFG-EN의 U1은 `Agent Team: off — 1 ready slice(s) can run together; running subagents`, 끈 CFG-EN은 `Agent Team: off — workflow.auto_team is false in autopus.yaml; running subagents`다

### S11: facts 검증과 기존 triage 회귀
Priority: Must
Given 아래 잘못된 facts와 `internal/cli/workflow_triage_test.go`의 기존 facts fixture가 있다
When 잘못된 facts로 명령을 실행하고, 기존 fixture로 `--topology` 없이 branch와 `33b5bb55` 빌드를 실행하고, UDJ facts로 `--topology` 없이 실행한다
Then 잘못된 facts는 모두 종료 1이고 stderr가 표의 문구와 같다
And 기존 fixture의 stdout은 두 빌드에서 bytes가 같고, UDJ facts를 `--topology` 없이 주면 종료 1과 `invalid workflow triage facts JSON`이다

| facts | stderr |
|---|---|
| `version: 2` | `invalid topology facts: version` |
| T1 depends_on `["T9"]` | `invalid topology facts: depends_on` |
| T1 depends_on `["T1"]` | `invalid topology facts: depends_on` |
| unit에 `"id"` key 두 번 | `invalid topology facts: json` |
| 최상위 `"extra": 1` | `invalid topology facts: json` |
| worker `owned_paths: []` | `invalid topology facts: owned_paths` |
| `owned_paths: ["../x"]`, `["/abs"]`, `["pkg/*"]` | `invalid topology facts: owned_paths` |
| `execution: "parallel"` | `invalid topology facts: execution` |
| unit 33개 | `invalid topology facts: unit_count` |
| id `T 1` | `invalid topology facts: unit_id` |

### S12: 출력 형식과 쓰기 없음
Priority: Must
Given CAP-OK, CFG-KO, UDJ, 실행 전 temp dir의 파일 목록과 SHA-256이 있다
When `--format json`과 `--format human`으로 실행한다
Then human stdout은 정확히 `Agent Team: 전환 — 준비된 slice 2개의 owned_paths가 겹치지 않아 native Agent Team으로 실행한다\n`이다
And JSON은 `mode: topology`, `unit_source: caller_declared`, `capability_source: observed_process_env`, `config_source: autopus.yaml`, `execution_performed: false`를 담는다
And 실행 뒤 temp dir의 파일 목록과 SHA-256이 실행 전과 같다

### S13: 생성 skill 문구
Priority: Must
Given branch에서 claude-code `Generate`를 돌린 결과와 `33b5bb55` 결과가 있다
When `.claude/skills/auto-go/SKILL.md`와 `.claude/skills/agent-teams/SKILL.md`를 읽는다
Then auto-go의 `#### Route A: Supervisor-Led Pipeline`과 `#### Route B: Native Agent Team` 사이에 `#### Route A Phase 2 Entry: Agent Team Auto-Selection`이 있고, 그 절에 `TOPOLOGY_FLAG_COUNT == 0`, `auto workflow triage --topology --facts-json "$FACTS" --format json`, `select:SendMessage,TaskCreate,TaskUpdate,TaskList,TaskGet`, `tools_fallback_line`, `Agent Team: off — auto workflow triage --topology failed; running subagents`가 이 순서로 나오고 "exactly one" 줄 규칙과 Gate 2 전 shutdown·ack 문장이 있다
And 기준과의 diff는 추가 줄뿐이고 모두 그 절 안이다
And agent-teams skill에는 `## Automatic Entry from Route A`가 있고 `--topology`는 없다

### S14: 판단 도구가 없거나 실패할 때
Priority: Must
Given PATH의 v0.50.125와 branch 바이너리, `workflow: [`로 깨진 autopus.yaml project가 있다
When v0.50.125로 명령을 실행하고 branch 바이너리로 깨진 project에서 명령을 실행한다
Then v0.50.125는 종료 1과 `Error: unknown flag: --topology`다(F4)
And branch는 종료 1과 `config load failed`이고, 두 경우 모두 생성 skill의 규칙(S13)대로 고정 실패 줄과 Route A로 이어진다

### S15: 실제 터미널 확인
Priority: Must
Given 실제 터미널의 대화형 최상위 `claude`(2.1.246 이상) 세션에 `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`이 있고, scratch project에 Phase 1 표가 (a) 준비된 disjoint worker unit 2개, (b) 준비된 worker unit 1개인 SPEC 둘이 있다
When 두 SPEC으로 `/auto go SPEC-ID --auto`를 돌리고, 같은 세션의 Bash와 `claude -p`에서 `CLAUDE_CODE_ENTRYPOINT`·`CLAUDE_CODE_SESSION_ATTENDED`·`CLAUDE_CODE_CHILD_SESSION`을 기록한다
Then (a)는 `Agent Team: 전환 — 준비된 slice 2개의 owned_paths가 겹치지 않아 native Agent Team으로 실행한다` 한 줄, runtime이 teammate로 보고한 `builder-1`·`builder-2`, shutdown ack, 이어진 Gate 2를 보인다(CD-1)
And (b)는 `Agent Team: 유지 — 함께 실행할 준비된 slice가 1개라 subagent로 실행한다` 한 줄만 보이고 named teammate가 없다(CD-2)
And 최상위 세션은 `cli`, `1`, CHILD 없음이고 `claude -p`는 `cli`·`1` 쌍이 아니다. 쌍이 같으면 A2는 FAIL이고 규칙 7을 고친 뒤에만 릴리스한다
And 기록은 `<home>`·scratch 경로를 바꿔 `evidence/manual-smoke-teamauto.txt`에 있다

### S16: 문서
Priority: Should
Given branch README, CHANGELOG, `auto workflow triage --help`가 있다
When 읽는다
Then README 설정 표에 `workflow.auto_team`(기본: 없음 = on) 행이 있고 CHANGELOG v0.50.126 항목은 기본 on, teammate마다 별도 Claude 세션(토큰 증가), 실험 기능, 끄는 법, auto 0.50.125 이하의 거부를 적는다
And help에 `--topology`가 있다

## Oracle Acceptance Notes

- S1–S12, S14의 oracle은 route·reason·line 문자열, 수, stderr 문구, bytes·SHA-256 비교다. S13은 생성 문구의 정확한 문자열과 순서, diff 범위다. S15는 실제 터미널 기록이다.
- INV-TA-01: S1–S4, S12–S14. INV-TA-02: S2–S4, S15. INV-TA-03: S4, S5, S15. INV-TA-04: S8. INV-TA-05: S6, S7, S9. INV-TA-06: S10.
