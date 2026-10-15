# SPEC-REA-001 수락 기준
## Test Scenarios

fixture는 모두 temp project 안에 만든다. 실제 명령(`NewRootCmd`)과 실제 adapter `Generate`·`Update`·clean을 쓴다. Node는 `[NEW] reaNodeProbe` seam으로만 주고, TTY는 `composeTTYProbe`, wizard는 `composeFormsProbe` + huh accessible runner(`composeRunForm`)와 `testing/iotest.OneByteReader` 입력, stdin은 Read를 세는 reader로 준다. exec seam은 부른 명령을 기록한다. 실제 `node`·`npx`·네트워크는 S18·S19에서만 쓴다. "X를 고른다"는 그 라벨의 번호 입력이고 Enter는 기본값이다.

- P3: `git init`한 빈 temp dir에 `--platforms claude-code,codex,antigravity-cli`. P5: P3 + `omp`, `opencode`. OLD: v0.50.125 바이너리로 `auto init --yes --platforms claude-code,codex,antigravity-cli`를 한 project(`features.rea` 없음, rea 항목 없음).
- B: 이 branch에서 결정 없음(action keep)으로 P3를 생성한 결과. Q: P3를 N24로 `auto update --yes --rea`한 결과(세 파일에 managed 항목, key 없음).
- N24 = `v24.11.0`, N20 = `v20.11.0`, N2218 = `v22.18.0`, N0 = 실행 파일 없음(seam 결과). C·G·X 행: S4 표. HOME sandbox: `HOME`, `XDG_CONFIG_HOME`, `CODEX_HOME`를 temp dir로 둔다.

### S1: 결정 key 읽기와 쓰기
Priority: Must
Given OLD autopus.yaml의 사본 8개가 있고 `features` 아래가 (a) rea 없음, (b) `rea: {enabled: true}`, (c) `rea: {enabled: false}`, (d) `rea: {enabled: "true"}`, (e) `rea: {enabled: 1}`, (f) `rea: {enabled: yes}`, (g) `rea: {}`, (h) `rea: {enabled: true, pin: "6.5.0"}`이다
When 각각 `config.Load`하고 (a)·(c)·(g)를 `config.Save`하며 `yaml.Marshal(config.DefaultFullConfig("x"))`를 한다
Then 예상 값은 Enabled가 (a) nil, (b) true, (c) false, (f) true, (g) nil이다. (d)는 ``cannot unmarshal !!str `true` into bool``, (e)는 ``cannot unmarshal !!int `1` into bool``, (h)는 `field pin not found in type config.ReaFeaturesConf`를 담은 오류로 실패한다
And (c)의 저장 결과는 `features` 아래 `rea:`와 그 다음 줄 `enabled: false`를 담고, (a)·(g)의 저장 결과와 DefaultFullConfig marshal 결과에는 `rea:` 줄이 없다

### S2: Decision State 판정
Priority: Must
Given (1) OLD, (2) Q, (3) Q에서 `.mcp.json`의 rea를 C3로 바꾸고 다른 두 파일의 rea를 지운 것, (4) Q + `features.rea.enabled: false`, (5) OLD + `features.rea.enabled: true`, (6) Q에서 `.mcp.json`·`.gemini/settings.json`의 rea를 지우고 platforms를 `claude-code` 하나로 줄인 것(managed는 선택되지 않은 `.codex/config.toml`에만)이 있다
When 각각 `[NEW] resolveReaDecision`을 부르고 `auto doctor --json`의 `data.rea.decision`과 `decision_source`를 읽는다
Then 예상 값은 (1) `undecided`·`none`, (2) `on`·`managed_entry`, (3) `undecided`·`none`, (4) `off`·`features.rea`, (5) `on`·`features.rea`, (6) `undecided`·`none`이다 그리고 (4)를 `auto update --yes`하면 세 파일에 `rea-agents` 문자열이 없어지고 autopus.yaml SHA-256은 실행 전과 같다

### S3: init 결정
Priority: Must
Given N24, HOME sandbox, 빈 temp dir 7개가 있다
When (1) `auto init --yes --platforms claude-code,codex,antigravity-cli`, (2) 같은 platforms의 대화형 wizard에서 REA Confirm을 Enter, (3) wizard에서 REA Confirm을 No, (4) `--yes --no-rea`, (5) `--yes --rea --no-rea`, (6) N20으로 (1), (7) `auto init --yes --platforms opencode,omp`를 실행한다
Then (1)·(2)는 세 파일에 `rea` 항목이 정확히 1개씩이고 각각 MCP Entry Shape Contract의 managed 모양과 같으며, autopus.yaml에 `rea:` 줄이 없고, 출력에 `REA 역공학 MCP를 등록했습니다`가 1번, 옛 바이너리 줄이 0번이다. (2)의 wizard에서 `REA 역공학 MCP` 단계는 1번 열린다
And (3)·(4)는 세 파일에 `rea-agents` 문자열이 없고 autopus.yaml이 `features.rea.enabled: false`를 담으며 출력에 `features.rea는 v0.50.126보다 오래된 auto가 읽지 못합니다`가 정확히 1번 있다
And (5)는 0이 아닌 종료 코드와 `--rea와 --no-rea는 함께 쓸 수 없습니다`로 끝나고 dir에 autopus.yaml과 `.mcp.json`이 없다
And (6)은 rea 항목이 없고 autopus.yaml이 `enabled: true`를 담으며 출력에 `node v20.11.0은 ^22.19.0 || ^24.11.0 || >=26.0.0 범위가 아닙니다`와 옛 바이너리 줄이 1번씩 있다
And (7)은 REA 질문·hint·`features.rea`가 없고 `npx -y rea-agents@6.5.0 setup` 안내 줄이 정확히 1번 있다

### S4: managed와 사용자 소유 판정 (플랫폼별 정확한 모양)
Priority: Must

| ID | 파일 | `rea` 값 | 예상 분류 |
|---|---|---|---|
| C1 | `.mcp.json` | `{"$comment": "REA reverse-engineering MCP (rea-agents@6.5.0)", "command": "/bin/bash", "args": ["-l", "-c", "exec npx -y rea-agents@6.5.0 mcp"]}` | managed |
| C2 | `.mcp.json` | C1의 `$comment`와 args의 `6.5.0`을 `6.4.0`으로 | other_pin |
| C3 | `.mcp.json` | `{"command": "npx", "args": ["-y", "rea-agents@6.5.0", "mcp"]}` | user_owned |
| C4 | `.mcp.json` | C1 + `"env": {"REA_MCP_MAX_RESPONSE_BYTES": "1"}` | user_owned |
| C5 | `.mcp.json` | C1의 `$comment`만 `"mine"` | user_owned |
| C6 | `.mcp.json` | `rea`가 없고 `rea-agents` key에 C1 | absent |
| G1 | `.gemini/settings.json` | `{"command": "/bin/bash", "args": ["-l", "-c", "exec npx -y rea-agents@6.5.0 mcp"]}` | managed |
| G2 | `.gemini/settings.json` | G1 + C1의 `$comment` | user_owned |
| G3 | `.gemini/settings.json` | G1의 셋째 arg가 `exec npx -y rea-agents@latest mcp` | other_pin |
| X1 | `.codex/config.toml` | `[mcp_servers.rea]` 아래 `command = "npx"`, `args = ["-y", "rea-agents@6.5.0", "mcp"]`, `startup_timeout_sec = 30` | managed |
| X2 | `.codex/config.toml` | X1에서 `startup_timeout_sec` 줄을 뺌 | user_owned |
| X3 | `.codex/config.toml` | X1 + `tool_timeout_sec = 120` | user_owned |
| X4 | `.codex/config.toml` | X1의 둘째 줄 앞에 `# mine` 주석 줄 | user_owned |
| X5 | `.codex/config.toml` | X1의 pin이 `6.4.0` | other_pin |
| X6 | `.codex/config.toml` | header가 `[mcp_servers.rea-agents]`이고 본문은 X1 | absent |

Given P3에 위 값을 하나씩 넣은 15개 project와 N24가 있다
When 각 값을 `[NEW] reamcp` 분류 함수로 판정하고, `auto update --yes --rea` 뒤 `auto doctor --json`을 실행한다
Then 판정과 doctor `data.rea.platforms`의 그 플랫폼 값은 표의 예상 분류와 같다
And update 뒤 managed·absent 행의 `rea`는 C1·G1·X1과 같은 값이고, other_pin·user_owned 행의 `rea`는 JSON deep-equal(Codex는 header부터 다음 header 전까지 줄이 같음)로 실행 전과 같으며, C6·X6의 `rea-agents` 항목도 실행 전과 같다

### S5: 제거와 clean의 round trip
Priority: Must
Given B에 사용자 서버 `mine`(`.mcp.json`·`.gemini/settings.json`의 `{"command": "mine-mcp"}`, `.codex/config.toml`의 `[mcp_servers.mine]`과 `command = "mine-mcp"`)을 더하고 `auto update --yes`로 한 번 정규화한 project와 그 세 파일의 SHA-256 H1, 그 사본에서 `.mcp.json`의 rea만 C3로 둔 P3u, N24가 있다
When (1) `auto update --yes --rea`, (2) `auto update --yes --no-rea`, (3) (1) 상태의 사본에서 `auto clean`, (4) P3u에서 (1), (2), `auto clean`을 차례로 실행한다
Then (2) 뒤 세 파일의 SHA-256은 H1과 같고 `.codex/config.toml`에 `mcp_servers.rea` 문자열이 없으며 `.gemini/settings.json`의 `mcpServers`에 `rea` key가 없다. autopus.yaml은 `features.rea.enabled: false`를 담는다
And (3) 뒤 세 파일 어디에도 `rea-agents@6.5.0`이 없고 `mine` 항목은 값이 실행 전과 같다
And (4)의 각 단계 뒤 `.mcp.json`의 `rea` 값은 C3와 deep-equal이고, `.codex/config.toml`·`.gemini/settings.json`의 managed rea는 (1)에서 생겼다가 (2)와 clean 뒤 없어진다

### S6: Node gate
Priority: Must
Given Q와 세 MCP 파일·autopus.yaml의 SHA-256, OLD + `features.rea.enabled: false` 사본이 있다
When `[NEW] reamcp.NodeVersionStatus`에 아래 입력을 주고, Q에서 seam이 `v22.18.0`, 실행 파일 없음, `garbage`를 내게 해 각각 `auto update --yes`를 실행하고, false 사본에서 `auto update --yes`를 실행한다
Then 예상 값: pass = `v22.19.0`, `v22.20.3`, `v24.11.0`, `v24.20.0`, `v26.0.0`, `v27.3.1`, `Welcome\nv24.11.0\n`(마지막 비지 않은 줄). unsupported = `v20.11.0`, `v22.18.0`, `v23.1.0`, `v24.10.0`, `v25.9.0`. unknown = `garbage`, `v26`, `v26.0.0-nightly20260101`, `24.11.0`. absent = 실행 파일 없음
And Q의 세 update 뒤 세 MCP 파일과 autopus.yaml SHA-256은 실행 전과 같고, 출력의 Node 안내는 실행마다 정확히 1줄(`node v22.18.0은 …`, `node를 찾지 못했습니다 …`, `node 버전을 확인하지 못했습니다`)이며 probe 호출은 실행마다 1번이다
And false 사본의 update는 probe를 0번 부른다

### S7: update 질문 상태 기계
Priority: Must
Given OLD 사본들, OLD + `features.rea.enabled: true` 사본 T, OLD + `false` 사본 F, OLD 두 개를 담은 workspace, N24, stdin Read를 세는 reader, TTY seam이 있다
When (1) TTY `auto update`에서 REA 질문에 Enter, 이어 TTY로 한 번 더. (2) TTY에서 `n`, 이어 한 번 더. (3) non-TTY `auto update`. (4) `auto update --yes`. (5) `auto update --auto`. (6) `auto update --plan`. (7) workspace에서 `auto update --workspace --yes`. (8) non-TTY `auto update --rea`, 이어 `auto update --no-rea`. (9) `auto update --rea --no-rea`. (10) workspace에서 `auto update --workspace --yes --rea`. (11) T에서 TTY·non-TTY·`--yes` 각각. (12) F에서 TTY·non-TTY·`--yes` 각각
Then 예상 값은 아래 표와 같다

| 실행 | 질문 수 | stdin Read | 세 MCP 파일의 rea | autopus.yaml `features.rea` | hint 줄 | 옛 바이너리 줄 |
|---|---|---|---|---|---|---|
| (1) 첫째 | 1 | 1 이상 | managed 1개씩 | 없음 | 0 | 0 |
| (2) 첫째 | 1 | 1 이상 | 없음 | `enabled: false` | 0 | 1 |
| (1)·(2) 둘째 | 0 | 0 | 첫째 뒤와 SHA-256 같음 | 첫째 뒤와 같음 | 0 | 0 |
| (3), (4), (5) | 0 | 0 | 없음 | 없음 | 1 | 0 |
| (6) | 0 | 0 | 모든 파일 SHA-256 실행 전과 같음 | 없음 | 1 | 0 |
| (7) | 0 | 0 | repo마다 없음 | 없음 | repo마다 1(모두 2) | 0 |
| (8) `--rea` | 0 | 0 | managed 1개씩 | 없음 | 0 | 0 |
| (8) `--no-rea` | 0 | 0 | 없음 | `enabled: false` | 0 | 1 |
| (9), (10) | 0 | 0 | 모든 파일 SHA-256 실행 전과 같음 | 없음 | 0 | 0 |
| (11) 각각 | 0 | 0 | managed 1개씩 | `enabled: true` 그대로 | 0 | 0 |
| (12) 각각 | 0 | 0 | 없음 | `enabled: false` 그대로 | 0 | 0 |

And (9)는 `--rea와 --no-rea는 함께 쓸 수 없습니다`, (10)은 `--rea·--no-rea는 --workspace와 함께 쓸 수 없습니다`를 담은 오류로 끝난다
And 쓰기가 있는 모든 실행 뒤 `.claude/agents/autopus/reverse-engineer.md`의 tools 줄은 `tools: Read, Grep, Glob, mcp__rea`이고 `.claude/skills/auto-reverse/SKILL.md`는 `확인 / 수정 / 중단`을 담는다

### S8: doctor
Priority: Must
Given P5에서 (a) Q 상태와 N24, (b) (a)를 `auto update --yes --no-rea`한 것, (c) OLD에 omp·opencode를 더한 것, (d) OLD + `features.rea.enabled: true` + `.mcp.json` C3 + `.codex/config.toml` X5 + Antigravity rea 없음과 N2218이 있다
When 각각 `auto doctor --json`과 `auto doctor`를 실행하고 exec seam이 부른 명령을 기록한다
Then (a)의 `data.rea` 예상 JSON은 `{"decision":"on","decision_source":"managed_entry","node":{"status":"pass","version":"v24.11.0"},"platforms":{"antigravity-cli":"managed","claude-code":"managed","codex":"managed"},"guidance":["omp","opencode"]}`이다
And (b)는 `off`·`features.rea`, platforms 세 값 `absent`, guidance `[]`. (c)는 `undecided`·`none`, platforms 세 값 `absent`, guidance `["omp","opencode"]`. (d)는 `on`·`features.rea`, node `{"status":"unsupported","version":"v22.18.0"}`, platforms `{"antigravity-cli":"absent","claude-code":"user_owned","codex":"other_pin"}`이고 `doctor.rea.node`의 severity는 `warning`이다
And 텍스트 출력은 (a) `REA: 켜짐 (managed 항목)`, (b) `REA: 꺼짐 (autopus.yaml)`, (c) `REA: 미결정`, (d) `REA: 켜짐 (autopus.yaml)`을 담고, 기록된 명령은 실행마다 node probe 1개이며 `npx`와 `rea`는 0개다

### S9: 17개 에이전트 인벤토리
Priority: Must
Given N24로 `auto init --yes --platforms claude-code,codex,antigravity-cli,omp,opencode`한 P5가 있다
When 플랫폼마다 생성 에이전트 파일의 이름 집합과 reverse-engineer 파일을 읽는다
Then `.claude/agents/autopus/*.md`, `.codex/agents/*.toml`, `.gemini/agents/autopus/*.md`, `.omp/agents/*.md`, `.opencode/agents/*.md`의 이름 집합(확장자 제외)은 각각 `CanonicalAgentNames()`와 같은 17개이고 `CanonicalAgentNames()[10]`은 `reverse-engineer`다
And Claude 파일은 `tools: Read, Grep, Glob, mcp__rea`, `permissionMode: default`, `model: claude-opus-5-5`, `effort: high` 줄을 담고 tools 줄에 `Write`, `Edit`, `Bash`가 없다
And Antigravity 파일의 tools는 `read_file`, `grep_search`, `glob`, `mcp_rea_*` 순서의 4개, OMP 파일의 tools는 `glob`, `grep`, `read` 3개이고 `mcp_`로 시작하는 값이 없으며, OpenCode 파일의 permission은 `"*": deny`와 allow `glob`, `grep`, `question`, `read`, `skill`, `task`뿐이다. Codex 파일은 `model = "gpt-6.1-sol"`, `model_reasoning_effort = "xhigh"`다
And 다섯 파일의 본문은 `REA MCP 도구를 찾지 못했습니다. auto update --rea로 이 프로젝트에 등록하거나 npx -y rea-agents@6.5.0 setup으로 사용자 범위에 등록하세요.`와 다섯 절 이름 `## Behaviour`, `## Inputs and Outputs`, `## Evidence`, `## Unknowns`, `## Acceptance Items`를 담는다

### S10: 프리셋 표와 분류
Priority: Must
Given `DefaultFullConfig`, `quality.default: ultra`인 사본, HB6 map(v0.50.125 balanced 16개), HU3 map(v0.50.125 ultra 16개), HB6 + `reverse-engineer: sonnet`인 map이 있다
When presets를 읽고 `ClassifyBuiltinQualityPreset`, `ClaudeAgentProfile`, `CodexAgentProfile`, `NativeBalancedAgentCandidate`를 부르고 compose 그룹을 렌더한다
Then 예상 값: balanced·ultra는 17개이고 reverse-engineer는 둘 다 `opus`다. balanced에서 Claude는 (`claude-opus-5-5`, `high`), Codex는 (`gpt-6.1-sol`, `xhigh`), 후보 selector는 anthropic `anthropic/claude-opus-5-5`, openai `openai-codex/gpt-6.1-sol`이다. ultra 사본에서 Claude는 (`opus`, `max`), Codex는 (`gpt-6.1-sol`, `xhigh`)다
And HB6 balanced는 class `historical`, Row `HB6`, RuntimeStandard true, Differs 1이고 그 project의 reverse-engineer도 (`claude-opus-5-5`, `high`)를 받는다. HU3 ultra는 `historical`·`HU3`, 17개 기본값은 `current`, HB6 + `reverse-engineer: sonnet`은 `customized`다
And 그룹 제목은 `탐색·보조 — explorer, annotator, reverse-engineer`, 기본값의 현재 값은 `혼합 (opus 1 · sonnet 2)`이고, 탐색·보조에 `haiku`를 고르면 세 에이전트 모두 `haiku`가 된다
And `balancedRungByAgent`의 key 집합은 `CanonicalAgentNames()`와 같다

### S11: OMP routing 17 + task
Priority: Must
Given P5의 OMP에 `role_model_policy {version: v1, profile: balanced, family: openai}`를 둔 project(SPEC-MODELUX-003 IO 생성 runner)와 그 사본에 `agent_source: native`를 더한 것이 있다
When 각각 `auto update --yes`를 실행하고 같은 run의 plan 미리보기와 `task.agentModelOverrides` 되읽기를 기록한다
Then autopus 모드 override key는 정규 17개와 `task`, 모두 18개이고 `reverse-engineer` 값은 `openai-codex/gpt-6.1-sol:xhigh`이며 나머지 17개 값은 같은 run의 미리보기와 같다
And `config.OMPNativeAgentForRole("reverse-engineer")`는 `task`이고, native 사본의 override key는 native 5개 그대로이며 `.omp/agents`에 생성 파일이 없다
And autopus 모드 OMP manifest의 `.omp/agents/` 항목은 17개이고, OMP 생성 전 project의 migration 미리보기는 `+ .omp/agents/reverse-engineer.md`를 포함한 17줄과 `OMP는 이제 Autopus 에이전트 17개를 씁니다.`로 시작하는 줄을 담는다

### S12: MCP 참조 변환 예외
Priority: Must
Given 본문 네 줄 L1 `Call mcp__rea__list_unknowns, then mcp__rea for the server.`, L2 `Use mcp__context7__query-docs("react", topic="hooks").`, L3 `Fallback mcp__other__y.`, L4 `Mixed mcp__rea__inspect_analysis_view and mcp__other__y.`가 있다
When `ReplacePlatformReferences(body, p)`를 p = codex, gemini, opencode, omp, claude로 부른다
Then 예상 출력은 codex·gemini·opencode·omp에서 L1 그대로, L2 `Use Context7 MCP first; fallback: WebSearch "react hooks docs" (topic: hooks).`, L3 `Fallback Context7 MCP with WebSearch fallback.`, L4 `Mixed mcp__rea__inspect_analysis_view and Context7 MCP with WebSearch fallback.`이고 claude에서는 네 줄 모두 그대로다 그리고 `make generate-templates` 뒤 Codex·Gemini reverse-engineer 산출물은 본문의 `mcp__rea`를 그대로 담고 `Context7 MCP with WebSearch fallback`을 담지 않는다

### S13: router와 reverse detail
Priority: Must
Given `DefaultFullConfig`로 다섯 플랫폼을 `Generate`한 project가 있다
When router와 `auto-reverse` detail을 읽는다
Then Claude root router(`.claude/skills/auto/SKILL.md`)에 `.claude/skills/auto-reverse/SKILL.md`가 정확히 1번 있고 router는 8192 bytes 이하다. Antigravity root router도 자기 경로로 같은 규칙이다. Codex router 설명의 예상 값은 `Autopus 명령 라우터 — setup/status/goal/update/plan/go/fix/review/sync/idea/map/why/verify/secure/test/qa/dev/canary/doctor/reverse 서브커맨드를 해석합니다`다
And 다섯 detail은 각각 `reverse-engineer`, `--auto도 이 확인 gate를 건너뛰지 않습니다`, `확인 / 수정 / 중단`, `.autopus/reverse/<slug>/behaviour-spec.md`, `executor에는 확인된 behaviour-spec.md와 프로젝트 문맥만 넘기고 REA 도구 결과, evidence bundle, 디컴파일 텍스트는 넘기지 않습니다`, `## Acceptance Items`를 담고 Claude detail에는 `run_in_background`가 없다 그리고 frozen 목록에 `reverse`를 더한 router budget test와 Update·Generate parity test가 통과한다

### S14: 사전 승인 없음
Priority: Must
Given S9 project와 `DetectPermissions` 결과가 있다
When `.claude/settings.json`, `.codex/config.toml`, `.gemini/settings.json`, `.omp/`, `.opencode/` 아래 생성 파일을 읽는다
Then `permissions.allow`의 MCP 값은 `mcp__sequential-thinking__sequentialthinking` 하나이고 `mcp__rea`로 시작하는 값이 없다 그리고 어떤 생성 파일에도 `enableAllProjectMcpServers`, `enabledMcpjsonServers`, `enabled_tools`, `default_tools_approval_mode`가 없다

### S15: 꺼져 있으면 bytes 그대로
Priority: Must
Given platforms 3개인 `config.DefaultFullConfig("p")`(결정 없음), 기준 commit 4d73aa91에서 같은 cfg로 만든 세 파일, OLD와 그 사본이 있다
When 이 branch에서 Claude·Codex·Antigravity `Generate`와 `Update`를 돌리고, OLD는 이 branch로, 사본은 기준 commit 바이너리로 `auto update --yes`한다
Then `.mcp.json`, `.codex/config.toml`, `.gemini/settings.json`은 기준 commit 결과와 bytes가 같다 그리고 OLD의 세 파일은 사본의 세 파일과 bytes가 같고 `rea`가 없다

### S16: 사용자 범위 무접촉
Priority: Must
Given HOME sandbox와 그 아래 파일 목록·SHA-256, N24가 있다
When S3 (1), S7 (8), `auto doctor`, `auto clean`을 이 branch로 실행하고 같은 명령을 기준 commit 바이너리로 다른 sandbox에서 실행한다
Then 두 sandbox의 실행 전후 파일 집합 차이는 같고 `.claude.json`, `.codex/config.toml`, `.gemini/settings.json`, `.omp/agent/mcp.json`이 sandbox에 생기지 않는다 그리고 `rea-agents`를 담은 파일은 모두 project root 아래에 있다

### S17: 회귀
Priority: Must
Given 이 branch와 기준 commit 4d73aa91의 test가 있다
When `go test ./...`를 필터 없이 돌리고, `git diff --stat 4d73aa91 -- '*_test.go' '*.golden' 'evals/'`를 T2 분류와 맞추며, `auto eval harness run`, `GOOS=windows GOARCH=amd64 go build ./...`, 신규·변경 `.go` 줄 수 확인, `make generate-templates` 뒤 `git status --porcelain`을 실행한다
Then 모든 test가 통과하고 바뀐 기존 test는 T2 목록과 Existing Test Changes 규칙대로다 그리고 eval harness 실패는 0, Windows build는 성공, 신규·변경 `.go`는 모두 300줄 이하(`update.go` 포함), generate 뒤 반영 안 된 diff는 0이다

### S18: 실제 터미널 smoke
Priority: Must
Given macOS 실제 터미널, 이 branch의 `auto`, Node v24.x, Claude Code·Codex·Antigravity 현재판, `git init`한 빈 temp dir, scratch `CODEX_HOME`, `script -q` 기록이 있다
When `auto init --platforms claude-code,codex,antigravity-cli`에서 REA 단계를 Enter로 넘기고, `claude`를 열어 project MCP 승인 질문에서 rea를 승인한 뒤 메시지 없이 `/mcp`와 `/agents`만 보고, `codex`와 Antigravity CLI의 MCP 목록을 메시지 없이 본다
Then `/mcp`에 rea가 connected이고 도구 목록이 보이며 `/agents`에 reverse-engineer가 있고 그 도구가 REA 도구와 Read·Grep·Glob이다. Codex·Antigravity 목록에 rea가 있다. 하나라도 다르면 그 출력을 적고 CD-1을 남긴다
And transcript를 `$HOME`, scratch 경로, 사용자 이름을 치환해 `evidence/manual-smoke-rea.txt`에 남긴다

### S19: 동의 뒤 live `/auto reverse`
Priority: Must
Given S18 project, 사용자의 모델 호출 동의, 외부 디컴파일러가 필요 없는 작은 로컬 정적 대상(JS 또는 Electron 앱 하나)이 있다
When `/auto reverse <대상> <기능 하나>`를 실행하고 gate에서 확인을 고른다
Then reverse-engineer 결과는 다섯 절, evidence_id 1개 이상, `## Unknowns` 절을 담고 의사코드·assembly 블록이 없다. 새로 생긴 project 파일은 `.autopus/reverse/<slug>/behaviour-spec.md`와 구현 파일뿐이고, executor dispatch 입력(transcript)은 그 경로와 프로젝트 문맥뿐이며 REA 결과 텍스트가 없다. tester 결과는 Acceptance Items마다 판정을 남긴다
And 결과를 `evidence/manual-smoke-rea.txt`에 남긴다. 동의가 없거나 막히면 PASS로 닫지 않고 CD-2를 남긴다

### S20: locale
Priority: Should
Given OLD 사본과 N20이 있다
When `tui.DetectSystemLang` seam을 `en`과 `ko`로 두고 non-TTY `auto update`와 `auto update --yes --rea`를 실행한다
Then en의 예상 출력은 `REA reverse-engineering MCP is undecided; run auto update --rea to enable or auto update --no-rea to decline`, `REA MCP not registered: node v20.11.0 is outside ^22.19.0 || ^24.11.0 || >=26.0.0`, `auto older than v0.50.126 cannot read features.rea`이고 ko는 Form Text의 한국어 문장이다 그리고 두 언어의 autopus.yaml 결과는 bytes가 같다

### S21: 문서
Priority: Should
Given help, README, CHANGELOG가 있다
When `auto init --help`, `auto update --help`, `auto doctor --help`를 실행하고 문서를 읽는다
Then 기본 on, `--rea`·`--no-rea`, Node 범위, `features.rea`의 lazy 쓰기와 옛 바이너리 한계, clean-room 흐름, runtime capture 주의가 있고, 기존 test가 고정한 문장이 그대로이며, `make generate-templates` 뒤 반영 안 된 diff가 없다

## Oracle Acceptance Notes

- Must 시나리오는 concrete expected output을 판정한다: 분류 문자열, `data.rea` 예상 JSON, 출력 줄과 횟수, frontmatter·TOML 예상 값, 파일 SHA-256, 호출 수, 버전 판정. 파일 존재나 종료 코드만으로 닫는 Must는 없다.
- 기대 값은 2026-10-15에 기준 commit 4d73aa91 코드, Claude Code·Codex·Gemini CLI 문서, npm `rea-agents@6.5.0`, auto 0.50.123 실행(`evidence/probe-rea-forward-compat.md`)으로 계산했다. 실행으로 확인한 것은 그 probe와 yaml.v3 decode뿐이다. 코드 사실과 다르면 test가 아니라 이 SPEC을 고친다.
- S11의 다른 17개 override 값과 S10의 history 판정은 같은 run의 미리보기·registry와의 일치로 판정한다. 둘 다 결정적 출력이다. S18·S19는 실제 터미널과 모델 호출이 필요해 기록 전까지 `research.md`의 CD-1·CD-2가 남는다.
