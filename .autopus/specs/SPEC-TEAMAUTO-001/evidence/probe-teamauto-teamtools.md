# Probe: 현재 Claude Code의 native team 도구 (2026-10-17)

- 결론: Claude Code 2.1.288·2.1.301의 일반 대화형 세션에는 공유 task-list 도구(`TaskCreate`, `TaskUpdate`, `TaskList`, `TaskGet`)가 없다. 자동 진입의 schema 확인이 이 도구를 요구해, 규칙 8을 지운 뒤에도 `tools_fallback_line`으로 subagent에 남았다. 팀에 실제로 필요한 것은 `Agent`의 `name` 인자와 `SendMessage`(`shutdown_request`·`shutdown_response` 포함)다. `ListAgents`도 있지만 lifecycle에 필요하지 않다.
- 비밀 값은 출력하지 않았다. 이 probe는 대화형 Claude 세션을 새로 띄우지 않았다. R은 lead의 S15 재-smoke 기록이다.

## R. 실제 터미널 (lead, S15 재-smoke, Claude Code 2.1.301, 규칙 8을 지운 바이너리)

| ID | 관측 |
|---|---|
| R1 | Phase 2 진입에서 CLI가 `Agent Team: 전환 — 준비된 slice 2개의 owned_paths가 겹치지 않아 native Agent Team으로 실행한다`를 냈다. 규칙 8 제거가 동작했다 |
| R2 | 세션의 ToolSearch `select:SendMessage,TaskCreate,TaskUpdate,TaskList,TaskGet`이 `SendMessage` 하나만 돌려줬다 |
| R3 | 그래서 세션이 그때의 fallback 줄 `Agent Team: 유지 — SendMessage·task 도구 schema를 불러오지 못해 subagent로 실행한다`를 출력하고 subagent로 갔다 |
| R4 | lead 세션의 deferred-tool 목록에도 `TaskCreate`·`TaskUpdate`·`TaskList`·`TaskGet`이 없다. 이름이 비슷한 `TaskStop`만 있고 이 lifecycle과 상관없다 |

## S. 정적 확인 (bundle 문자열, 2.1.288·2.1.301)

`strings -n 8 <바이너리>`로 꺼냈다. 바이너리 위치는 `probe-teamauto-childsession.md` S와 같다.

- S1: `TaskCreate`는 `shouldDefer:!0`인 deferred tool이고 `isEnabled(){return joe()}`다. 2.1.301 원문:

  ```js
  function eI(){if(a.CLAUDE_CODE_ENABLE_TASKS===!1)return!1;return!0}
  function oV(){if(_c()||UWo())return!0;let e=dFt();if(e===void 0||A9r(e)||x9r(e))return!0;return a.CLAUDE_CODE_ENABLE_TODO_TOOLS===!0}function joe(){return eI()&&oV()}
  ```

  `_c()`는 job 실행 관련 판정, `UWo()`는 launch option `todoToolsOptIn`, `dFt()`는 main loop model, `A9r`는 Bedrock `application-inference-profile` id, `x9r`는 Claude 3.x·4.x model 목록(`claude-opus-4-7`, `claude-haiku-4-5`까지)이다. 그래서 task-list 도구는 `CLAUDE_CODE_ENABLE_TASKS`가 false가 아니면서, job 실행이거나 opt-in이거나 main loop model이 정해지지 않았거나 그 옛 목록·inference profile이거나 `CLAUDE_CODE_ENABLE_TODO_TOOLS=1`일 때만 켜진다. 현재 model의 대화형 세션은 기본으로 꺼져 있다.
- S2: `Agent`의 team 이름 인자 설명은 `Deprecated; ignored. The session has a single implicit team.`이다. `name`은 문자나 숫자로 시작하는 64자 이하 이름이고 `main`, `team-lead`, `user`는 예약어다. `builder-1`·`tester`·`guardian`은 허용된다.
- S3: `ListAgents`는 `list agents you can SendMessage to`라는 tool이다.
- S4: 이 수정 작업을 맡은 teammate 세션(2.1.288)에는 `SendMessage`가 바로 있고 `Agent`, `ToolSearch`, `TaskCreate`는 없다. teammate는 팀을 이끌 수 없고, 새 확인(`SendMessage`와 `Agent`의 `name`)에서 `Agent`가 없어 fallback으로 남는다.

## 결정

- 자동 진입과 `--team`(Route B)의 필수 도구는 `SendMessage`와 `Agent`의 `name` 인자다. task-list 도구는 같은 ToolSearch 호출로 요청하고 돌아오면 쓴다. 없으면 lead가 배정 표를 직접 든다. 배정은 각 팀원 prompt에 담고, 완료는 팀원 결과나 `SendMessage` 보고로 알고, receipt 다섯 필드는 lead가 기록한다.
- 둘 중 하나가 없으면 `tools_fallback_line`을 출력하고 Route A다. 줄 문구를 `Agent Team: 유지 — SendMessage 도구나 Agent의 name 인자를 쓸 수 없어 subagent로 실행한다`(en `Agent Team: off — the SendMessage tool or the Agent name parameter is unavailable; running subagents`)로 바꿨다. task-list 도구가 없다는 것은 fallback 이유가 아니다.
- teardown은 `SendMessage`의 `shutdown_request`와 각 팀원의 `shutdown_response`다. team 이름은 넘기지 않는다.
- 사용자 결정(2026-10-17): v0.50.126에서 이렇게 맞춰 내보낸다.
