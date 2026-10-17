# Probe: CLAUDE_CODE_CHILD_SESSION은 최상위 세션과 subagent를 가르는가 (2026-10-17)

- 결론: 가르지 못한다. Claude Code는 세션이 띄우는 모든 process에 `CLAUDE_CODE_CHILD_SESSION=1`을 넣는다. 대화형 최상위 세션의 Bash tool도 이 값을 보므로 규칙 8(`not_top_level`)이 자동 선택을 늘 막았다. `plan.md` A2의 "최상위 대화형 세션에는 `CLAUDE_CODE_CHILD_SESSION`이 없다"는 FAIL이다.
- 비밀 값은 출력하지 않았다. `CLAUDE_CODE_MESSAGING_SOCKET`, `CLAUDE_CODE_MESSAGING_TOKEN`, `CLAUDE_CODE_SESSION_ID` 값은 `<redacted>`로, 경로는 `<home>`·`<scratch>`로 바꿨다. 이 probe는 대화형 Claude 세션을 새로 띄우지 않았다. H는 lead가 띄운 S15 smoke 기록이다.

## H. 실제 터미널 (S15 smoke, Claude Code 2.1.301)

lead가 tmux PTY에서 `<scratch>/smoke-kit/team-proj`(RC 바이너리로 `auto init` 한 project, SPEC-A는 disjoint ready worker T1·T2와 T1에 의존하는 T3)를 열고 `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 claude`를 실행했다. 기록은 script 출력 `<scratch>/smoke-kit/31b-team.txt`, `31c-team.txt`다. 배너는 `Claude Code v2.1.301`이다.

| ID | 위치 | 관측 |
|---|---|---|
| H1 | 최상위 대화형 세션에서 사용자가 `! env \| grep -E 'CLAUDE_CODE_(ENTRYPOINT\|SESSION_ATTENDED\|CHILD_SESSION)'` 실행(31c) | `CLAUDE_CODE_ENTRYPOINT=cli`, `CLAUDE_CODE_SESSION_ATTENDED=1`, `CLAUDE_CODE_CHILD_SESSION=1` |
| H2 | 같은 세션의 `/auto go SPEC-A --auto`, Phase 2 진입(31b, 31c) | 두 run 모두 `Agent Team: 유지 — 하위 세션은 팀 lead가 될 수 없어 subagent로 실행한다`. 팀이 켜질 plan인데 subagent로 남았다 |
| H3 | 같은 세션 Bash tool과 그 세션이 띄운 subagent의 Bash tool 환경 비교(lead 보고) | 두 목록이 같다. 차이는 subagent에 `CLAUDE_EFFORT`가 없는 것 하나이고, 이 값은 계약이 아니다(S2) |

## T. agent team teammate의 Bash tool (Claude Code 2.1.288)

이 수정 작업을 맡은 teammate(lead가 띄운 named teammate, 최상위 세션이 아님)의 Bash tool에서 `env | grep -E '^(CLAUDE|ANTHROPIC|AUTOPUS|AI_AGENT)'`를 실행하고 비밀 값을 바꿨다.

| 이름 | 값 |
|---|---|
| `AI_AGENT` | `claude-code_2-1-288_agent` |
| `CLAUDECODE` | `1` |
| `CLAUDE_CODE_CHILD_SESSION` | `1` |
| `CLAUDE_CODE_ENTRYPOINT` | `cli` |
| `CLAUDE_CODE_EXECPATH` | `/opt/homebrew/lib/node_modules/@anthropic-ai/claude-code/bin/claude.exe`, `--version` 첫 줄 `2.1.288 (Claude Code)` |
| `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` | `1` |
| `CLAUDE_CODE_MESSAGING_SOCKET` | `<redacted>` |
| `CLAUDE_CODE_MESSAGING_TOKEN` | `<redacted>` |
| `CLAUDE_CODE_SESSION_ATTENDED` | `1` |
| `CLAUDE_CODE_SESSION_ID` | `<redacted>` |
| `CLAUDE_DISABLE_PATH_WARNING` | `1` |
| `CLAUDE_EFFORT` | `max` |
| `CLAUDE_PID` | `<pid>` |

- stdin·stdout은 TTY가 아니다.
- 판단이 읽던 env 다섯(`CLAUDECODE`, `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS`, `CLAUDE_CODE_ENTRYPOINT`, `CLAUDE_CODE_SESSION_ATTENDED`, `CLAUDE_CODE_CHILD_SESSION`)의 값이 최상위 세션(H1)과 모두 같다.
- 2026-10-15 probe C1(subagent 안, `probe-teamauto-capability.md`)의 이름 목록은 위 표에서 `AI_AGENT`를 뺀 것과 같다. `AI_AGENT`는 그 probe의 grep 범위 밖이었다. C1에는 `CLAUDE_EFFORT`가 있었으므로 H3의 차이도 subagent의 성질이 아니다.

## S. 정적 확인: child process env builder

`strings -n 8 <바이너리>`로 꺼낸 bundle 문자열이다. 2.1.288(위 EXECPATH)과 2.1.301(PATH의 npm 설치, `<home>/.nvm/versions/node/v24.20.0/lib/node_modules/@anthropic-ai/claude-code/bin/claude.exe`, Mach-O arm64)이 같은 함수를 담는다. minify 이름만 다르다(`N1e`, `U5e`). 2.1.288 원문:

```js
let n={CLAUDECODE:"1",CLAUDE_CODE_SESSION_ID:e.sessionId,CLAUDE_CODE_CHILD_SESSION:"1",CLAUDE_CODE_SESSION_ATTENDED:YY()?"1":"0",CLAUDE_PID:String(process.pid)};
if(e.source==="agent")n.AI_AGENT=cto("agent");
if(e.effortLevel!==void 0)n.CLAUDE_EFFORT=e.effortLevel;
```

- S1: `CLAUDE_CODE_CHILD_SESSION`은 조건 없이 `"1"`이다. 뜻은 "Claude Code 세션이 띄운 process"이고 subagent 여부와 상관없다. 같은 bundle의 문구도 그렇게 쓴다: `this looks like a command run by a Claude Code session, but that could not be confirmed (its session marker may be inherited)`, `Transcript saving is off — inherited CLAUDE_CODE_CHILD_SESSION marker`.
- S2: `CLAUDE_EFFORT`는 effort 수준이 정해졌을 때만 붙는다. 최상위 여부와 상관없다.
- S3: Claude Code는 시작할 때 자기 환경의 값을 `setChildSession(Boolean(a.CLAUDE_CODE_CHILD_SESSION))`로 읽는다. 이 변수가 알려 주는 것은 "이 claude process를 다른 Claude Code 세션이 띄웠다"뿐이다. 세션의 Bash tool이 띄운 `auto`에게는 늘 `1`이다.

## 결정과 남은 확인

- 규칙 8과 `not_top_level` 줄을 지웠다. CLI는 `CLAUDE_CODE_CHILD_SESSION`을 읽지 않는다. 환경 변수 중 최상위를 가르는 계약된 값은 없다(`CLAUDE_EFFORT`, `AI_AGENT`는 effort 설정과 호출 출처에 따른다).
- subagent는 세션 쪽 schema 확인으로 subagent에 남는다. route가 `team`이어도 `SendMessage`를 불러오지 못하거나 `Agent`에 `name` 인자가 없으면 `tools_fallback_line`을 출력하고 Route A로 간다(REQ-TA-07). 처음에는 공유 task-list 도구까지 요구했으나 같은 날 그 도구가 기본 세션에 없음을 확인해 선택으로 바꿨다(`probe-teamauto-teamtools.md`). 생성 auto-go·agent-teams skill이 이를 적는다.
- CAP-OK fixture는 실제 최상위 세션처럼 `CLAUDE_CODE_CHILD_SESSION=1`을 담는다.
- A2의 나머지(print mode가 (`cli`, `1`) 쌍이 아닌지, npm 설치 세션의 EXECPATH 버전)는 아직 S15(e) 확인 전이다. 2.1.301 npm 설치의 실제 파일 이름이 `claude.exe`라는 것은 파일로만 확인했고, 그 세션의 `CLAUDE_CODE_EXECPATH` 값은 기록되지 않았다.
