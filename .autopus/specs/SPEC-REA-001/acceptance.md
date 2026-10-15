# SPEC-REA-001 수락 기준
## Test Scenarios

fixture는 모두 temp project 안에 만든다. 실제 명령(`NewRootCmd`)과 실제 adapter `Generate`·`Update`·`Clean`을 쓴다. Node는 exec 경계 seam `[NEW] reaNodeExec`, TTY는 `composeTTYProbe`, CI는 `composeCIEnv`, locale은 `composeLanguage`로 준다. CLI의 init wizard는 `runInitWizard` 결과 stub으로, REA 단계 자체는 `internal/cli/tui` test에서 huh accessible mode와 주입한 reader로 돈다. REA 질문 수는 출력의 질문 문장 수로 세고 stdin Read 수는 non-TTY 실행에서만 단언한다. 실제 `node`·`npx`·모델 호출은 S18·S19에서만 쓴다.

- P3: `git init`한 빈 temp dir에 `--platforms claude-code,codex,antigravity-cli`. P5: P3 + `omp`, `opencode`. OLD: v0.50.125 바이너리로 `auto init --yes --platforms claude-code,codex,antigravity-cli`를 한 project. BASE: 기준 commit 4d73aa91로 만든 바이너리.
- B: 이 branch에서 결정 없음(action keep)으로 P3를 생성한 결과. Q: P3를 N24로 `auto update --yes --rea`한 결과(세 파일에 managed 항목, key 없음). N24 = `v24.11.0` 종료 0, N20 = `v20.11.0` 종료 0, N2218 = `v22.18.0` 종료 0(seam 결과).

### S1: 결정 key 읽기와 쓰기
Priority: Must
Given OLD autopus.yaml의 사본 8개가 있고 `features` 아래가 (a) rea 없음, (b) `rea: {enabled: true}`, (c) `rea: {enabled: false}`, (d) `rea: {enabled: "true"}`, (e) `rea: {enabled: 1}`, (f) `rea: {enabled: yes}`, (g) `rea: {}`, (h) `rea: {enabled: true, pin: "6.5.0"}`이다
When 각각 `config.Load`하고 (a)·(c)·(g)를 `config.Save`하며 `yaml.Marshal(config.DefaultFullConfig("x"))`를 한다
Then 예상 값은 Enabled가 (a) nil, (b) true, (c) false, (f) true, (g) nil이다. (d)는 ``cannot unmarshal !!str `true` into bool``, (e)는 ``cannot unmarshal !!int `1` into bool``, (h)는 `field pin not found in type config.ReaFeaturesConf`를 담은 오류로 실패한다 그리고 (c)의 저장 결과는 `features` 아래 `rea:`와 그 다음 줄 `enabled: false`를 담고, (a)·(g)의 저장 결과와 DefaultFullConfig marshal 결과에는 `rea:` 줄이 없다

### S2: Decision State 판정과 off sweep
Priority: Must
Given (1) OLD, (2) Q, (3) Q에서 `.mcp.json`의 rea를 S4의 C3로 바꾸고 다른 두 파일의 rea를 지운 것, (4) Q + `features.rea.enabled: false`, (5) OLD + `features.rea.enabled: true`, (6) Q에서 `.mcp.json`·`.gemini/settings.json`의 rea를 지우고 platforms를 `claude-code` 하나로 줄인 것(managed는 선택되지 않은 `.codex/config.toml`에만)이 있다
When 각각 `[NEW] resolveReaDecision`을 부르고 `auto doctor --json`의 `data.rea.decision`과 `decision_source`를 읽은 뒤 (4)에서 `auto update --yes`, (6)에서 `auto update --yes --no-rea`를 실행한다
Then 예상 값은 (1) `undecided`·`none`, (2) `on`·`managed_entry`, (3) `undecided`·`none`, (4) `off`·`features.rea`, (5) `on`·`features.rea`, (6) `undecided`·`none`이다
And (4)의 update 뒤 세 파일에 `rea-agents` 문자열이 없고 autopus.yaml SHA-256은 실행 전과 같다. (6)의 update 뒤 선택되지 않은 `.codex/config.toml`에 `mcp_servers.rea`가 없고 나머지 줄은 bytes가 같으며 autopus.yaml은 `enabled: false`를 담는다

### S3: init 결정
Priority: Must
Given N24, HOME sandbox, 빈 temp dir 7개, `features.rea.enabled: false`이고 rea 항목이 없는 기존 project E와 그 사본 E2가 있다
When (1) `auto init --yes --platforms claude-code,codex,antigravity-cli`, (2) 대화형 init에서 `runInitWizard` stub이 `Rea: true`, (3) stub이 `Rea: false`, (4) `--yes --no-rea`, (5) `--yes --rea --no-rea`, (6) N20으로 (1), (7) `auto init --yes --platforms opencode,omp`, (8) E에서 `auto init --yes`, (9) E2에서 `auto init --yes --rea`를 실행하고, tui test에서 REA 단계 form을 accessible 입력 `\n`과 `n\n`으로 돌린다
Then (1)·(2)는 세 파일에 `rea` 항목이 정확히 1개씩이고 각각 MCP Entry Shape Contract의 managed 모양과 같으며, `.gitignore`가 `.autopus/reverse/`를 정확히 1번 담고, autopus.yaml에 `rea:` 줄이 없고, 출력에 `REA 역공학 MCP를 등록했습니다`가 1번, 옛 바이너리 줄이 0번이다
And (3)·(4)는 세 파일에 `rea-agents`가 없고 autopus.yaml이 `features.rea.enabled: false`를 담으며 `features.rea는 v0.50.126보다 오래된 auto가 읽지 못합니다`가 정확히 1번 있다. (5)는 0이 아닌 종료 코드와 `--rea와 --no-rea는 함께 쓸 수 없습니다`로 끝나고 dir에 autopus.yaml과 `.mcp.json`이 없다
And (6)은 rea 항목이 없고 autopus.yaml이 `enabled: true`를 담으며 `node v20.11.0은 ^22.19.0 || ^24.11.0 || >=26.0.0 범위가 아닙니다`와 옛 바이너리 줄이 1번씩 있다. (7)은 REA 질문·hint·`features.rea`가 없고 OMP·OpenCode 안내 줄이 정확히 1번 있다
And (8)은 `enabled: false`를 그대로 두고 rea 항목이 없으며 옛 바이너리 줄이 0번이다. (9)는 `enabled: true`와 managed 항목 1개씩을 남기고 옛 바이너리 줄이 0번이다
And tui test에서 `REA 역공학 MCP` 제목은 1번 나오고 결과는 `\n` → true, `n\n` → false다. `--rea`나 `--no-rea`가 있거나 대상 플랫폼이 없으면 `buildStepList`가 REA 단계를 넣지 않는다

### S4: managed와 사용자 소유 판정 (플랫폼별 정확한 모양)
Priority: Must

| ID | 파일 | `rea` 값 | update 전 | `--rea` 뒤 |
|---|---|---|---|---|
| C1 | `.mcp.json` | `{"$comment": "REA reverse-engineering MCP (rea-agents@6.5.0)", "command": "/bin/bash", "args": ["-l", "-c", "exec npx -y rea-agents@6.5.0 mcp"]}` | managed | managed |
| C2 | `.mcp.json` | C1의 두 pin 자리를 모두 `6.4.0`으로 | other_pin | other_pin |
| C3 | `.mcp.json` | `{"command": "npx", "args": ["-y", "rea-agents@6.5.0", "mcp"]}` | user_owned | user_owned |
| C4 | `.mcp.json` | C1 + `"env": {"REA_MCP_MAX_RESPONSE_BYTES": "1"}` | user_owned | user_owned |
| C5 | `.mcp.json` | C1의 `$comment`만 `"mine"` | user_owned | user_owned |
| C6 | `.mcp.json` | `rea`가 없고 `rea-agents` key에 C1 | absent | managed |
| C7 | `.mcp.json` | C1에서 args의 pin만 `6.4.0`(`$comment`는 6.5.0) | user_owned | user_owned |
| G1 | `.gemini/settings.json` | `{"command": "/bin/bash", "args": ["-l", "-c", "exec npx -y rea-agents@6.5.0 mcp"]}` | managed | managed |
| G2 | `.gemini/settings.json` | G1 + C1의 `$comment` | user_owned | user_owned |
| G3 | `.gemini/settings.json` | G1의 셋째 arg가 `exec npx -y rea-agents@latest mcp` | other_pin | other_pin |
| X1 | `.codex/config.toml` | `[mcp_servers.rea]` 아래 `command = "npx"`, `args = ["-y", "rea-agents@6.5.0", "mcp"]`, `startup_timeout_sec = 30` | managed | managed |
| X2 | `.codex/config.toml` | X1에서 `startup_timeout_sec` 줄을 뺌 | user_owned | user_owned |
| X3 | `.codex/config.toml` | X1 + `tool_timeout_sec = 120` | user_owned | user_owned |
| X4 | `.codex/config.toml` | X1의 둘째 줄 앞에 `# mine` 주석 줄 | user_owned | user_owned |
| X5 | `.codex/config.toml` | X1의 pin이 `6.4.0` | other_pin | other_pin |
| X6 | `.codex/config.toml` | header가 `[mcp_servers.rea-agents]`이고 본문은 X1 | absent | managed |
| X7 | `.codex/config.toml` | X1 뒤에 `[mcp_servers.rea.env]`와 `A = "1"` | user_owned | user_owned |

Given P3에 위 값을 하나씩 넣은 17개 project와 N24가 있다
When 각 값을 `[NEW] reamcp`로 판정하고, `auto update --yes --rea` 뒤와 이어서 `auto update --yes --no-rea` 뒤에 `auto doctor --json`을 실행한다
Then 판정은 "update 전" 열, `--rea` 뒤 doctor `data.rea.platforms`의 그 플랫폼 값은 "`--rea` 뒤" 열과 같고, managed가 된 행의 `rea`는 C1·G1·X1과 같은 값이다
And other_pin·user_owned 행의 `rea`는 `--rea`와 `--no-rea` 뒤 모두 입력과 JSON deep-equal이고 X2–X5·X7은 rea 서버 table 줄(X7은 `[mcp_servers.rea.env]` 포함)의 bytes가 입력과 같으며, C6·X6의 `rea-agents` 항목도 입력과 같다

### S5: 제거의 round trip과 `auto platform remove`
Priority: Must
Given B에 사용자 서버 `mine`을 비정규 서식(4칸 들여쓰기, key 순서 `{"note": "café", "command": "mine-mcp"}`)으로 `.mcp.json`·`.gemini/settings.json`에 손으로 넣고 `.codex/config.toml`에 `[mcp_servers.mine]`과 `command = "mine-mcp"  # keep`을 더한 P3m, BASE로 P3m 사본을 `auto update --yes`한 세 파일 bytes R, `.mcp.json`의 rea만 C3인 P3u, N24가 있다
When P3m에서 (1) `auto update --yes --rea`, (2) `auto update --yes --no-rea`를 실행하고, (1) 상태의 사본에서 (3) `auto platform remove codex`와 `auto platform remove antigravity-cli`, P3u에서 (1)과 (2)를 실행한다
Then (2) 뒤 세 파일의 bytes는 R과 같다(기준 commit의 정규화만 남는다). `mine` 값은 입력과 deep-equal이고 `[mcp_servers.mine]` 두 줄은 입력과 bytes가 같으며 `.codex/config.toml`에 `mcp_servers.rea`가, `.gemini/settings.json`에 `rea` key가 없다. autopus.yaml은 `enabled: false`를 담는다
And (3) 뒤 `.codex/config.toml`과 `.gemini/settings.json`에 `rea-agents@6.5.0`이 없고 `.mcp.json`의 managed rea는 남는다. P3u의 각 단계 뒤 `.mcp.json`의 `rea`는 C3와 deep-equal이다

### S6: Node gate (exec 경계)
Priority: Must
Given Q와 세 MCP 파일·autopus.yaml의 SHA-256, OLD + `features.rea.enabled: false` 사본이 있다
When `[NEW] reamcp`의 판정에 아래 exec 결과(stdout, 종료 상태, 시작 오류)를 주고, Q에서 seam이 `v22.18.0` 종료 0, 빈 stdout 종료 127, `garbage` 종료 0을 내게 해 각각 `auto update --yes`를 실행하고, false 사본에서 `auto update --yes`를 실행한다
Then 예상 값: pass = 종료 0의 `v22.19.0`, `v22.20.3`, `v24.11.0`, `v24.20.0`, `v26.0.0`, `v27.3.1`, `Welcome\nv24.11.0\n`. unsupported = 종료 0의 `v20.11.0`, `v22.18.0`, `v23.1.0`, `v24.10.0`, `v25.9.0`. absent = `/bin/bash` 시작 실패(ENOENT), 빈 stdout 종료 127. unknown = 종료 0의 `garbage`, `v26`, `v26.0.0-nightly20260101`, `24.11.0`, 종료 1의 `v24.11.0`, `bash: rc` 출력 뒤 종료 127, 5초 timeout
And Q의 세 update 뒤 세 MCP 파일과 autopus.yaml SHA-256은 실행 전과 같고, Node 안내는 실행마다 정확히 1줄(`node v22.18.0은 …`, `node를 찾지 못했습니다 …`, `node 버전을 확인하지 못했습니다`)이며 probe 호출은 실행마다 1번이다. false 사본의 update는 probe를 0번 부른다

### S7: update 질문 상태 기계
Priority: Must
Given OLD 사본들, OLD + `enabled: true` 사본 T, OLD + `enabled: false` 사본 F, OLD 두 개를 담은 workspace, platforms가 `claude-code` 하나이고 manifest에 있던 `.mcp.json`을 지운 project M1, platforms가 `claude-code` 하나이고 codex가 감지되는 project M2, N24, TTY seam이 있다
When (1) TTY에서 REA 질문에 Enter, 이어 TTY로 한 번 더. (2) TTY에서 `n`, 이어 한 번 더. (3) non-TTY. (4) `--yes`. (5) `--auto`. (6) `--plan`. (7) `--workspace --yes`. (8) non-TTY `--rea`, 이어 `--no-rea`. (9) `--rea --no-rea`. (10) `--workspace --yes --rea`. (11) T에서 TTY·non-TTY·`--yes`. (12) F에서 같은 셋. (13) M1에서 `--yes --rea`. (14) `--plan --rea`. (15) M2에서 `--yes --rea`. (16) TTY + `CI=true`로 `auto update`를 실행한다. TTY 실행은 compose 단계 입력도 Enter로 준다
Then 예상 값은 아래 표와 같다

| 실행 | REA 질문 | 세 MCP 파일의 rea | `features.rea` | hint | 옛 바이너리 줄 |
|---|---|---|---|---|---|
| (1) 첫째 | 1 | managed 1개씩 | 없음 | 0 | 0 |
| (1)·(2) 둘째 | 0 | 첫째 뒤와 SHA-256 같음 | 첫째 뒤와 같음 | 0 | 0 |
| (2) 첫째 | 1 | 없음 | `enabled: false` | 0 | 1 |
| (3), (4), (5), (16) | 0 | 없음 | 없음 | 1 | 0 |
| (6) | 0 | 모든 파일 SHA-256 실행 전과 같음 | 없음 | 1 | 0 |
| (7) | 0 | repo마다 없음 | 없음 | repo마다 1(모두 2) | 0 |
| (8) `--rea` / `--no-rea` | 0 / 0 | managed 1개씩 / 없음 | 없음 / `enabled: false` | 0 / 0 | 0 / 1 |
| (9), (10) | 0 | 모든 파일 SHA-256 실행 전과 같음 | 없음 | 0 | 0 |
| (11) 각각 | 0 | managed 1개씩 | `enabled: true` 그대로 | 0 | 0 |
| (12) 각각 | 0 | 없음 | `enabled: false` 그대로 | 0 | 0 |
| (13) | 0 | `.mcp.json`이 다시 생기지 않음 | `enabled: true` | 0 | 1 |
| (14) | 0 | 모든 파일 SHA-256 같음, 미리보기에 세 파일 변경 | 없음 | 0 | 0 |
| (15) | 0 | `.mcp.json`과 새 `.codex/config.toml`에 managed 1개씩 | 없음 | 0 | 0 |

And TTY가 아닌 실행의 stdin Read는 0이다. (9)는 `--rea와 --no-rea는 함께 쓸 수 없습니다`, (10)은 `--rea·--no-rea는 --workspace와 함께 쓸 수 없습니다`를 담은 오류로 끝난다
And 쓰기가 있는 모든 실행 뒤 `.claude/agents/autopus/reverse-engineer.md`의 tools 줄은 `tools: Read, Grep, Glob, mcp__rea`이고 `.claude/skills/auto-reverse/SKILL.md`는 `확인 / 수정 / 중단`을 담는다

### S8: doctor
Priority: Must
Given P5에서 (a) Q 상태와 N24, (b) (a)를 `auto update --yes --no-rea`한 것, (c) OLD에 omp·opencode를 더한 것, (d) OLD + `features.rea.enabled: true` + `.mcp.json` C3 + `.codex/config.toml` X5 + Antigravity rea 없음과 N2218이 있다
When 각각 `auto doctor --json`과 `auto doctor`를 실행하고 exec seam이 부른 명령을 기록한다
Then (a)의 `data.rea` 예상 JSON은 `{"decision":"on","decision_source":"managed_entry","node":{"status":"pass","version":"v24.11.0"},"platforms":{"antigravity-cli":"managed","claude-code":"managed","codex":"managed"},"guidance":["omp","opencode"]}`이다
And (b)는 `off`·`features.rea`, platforms 세 값 `absent`, guidance `[]`. (c)는 `undecided`·`none`, platforms 세 값 `absent`, guidance `["omp","opencode"]`. (d)는 `on`·`features.rea`, node `{"status":"unsupported","version":"v22.18.0"}`, platforms `{"antigravity-cli":"absent","claude-code":"user_owned","codex":"other_pin"}`이고 `doctor.rea.node`의 severity는 `warning`이다
And 텍스트 출력은 (a) `REA: 켜짐 (managed 항목)`, (b) `REA: 꺼짐 (autopus.yaml)`, (c) `REA: 미결정`, (d) `REA: 켜짐 (autopus.yaml)`을 담고, 기록된 명령은 실행마다 node probe 1개이며 `npx`와 `rea`는 0개다

### S9: 17개 에이전트 인벤토리와 본문
Priority: Must
Given N24로 `auto init --yes --platforms claude-code,codex,antigravity-cli,omp,opencode`한 P5가 있다
When 플랫폼마다 생성 에이전트 파일의 이름 집합과 reverse-engineer 파일을 읽는다
Then `.claude/agents/autopus/*.md`, `.codex/agents/*.toml`, `.gemini/agents/autopus/*.md`, `.omp/agents/*.md`, `.opencode/agents/*.md`의 이름 집합(확장자 제외)은 각각 `CanonicalAgentNames()`와 같은 17개이고 `CanonicalAgentNames()[10]`은 `reverse-engineer`다
And Claude 파일은 `tools: Read, Grep, Glob, mcp__rea`, `permissionMode: default`, `model: claude-opus-5-5`, `effort: high` 줄을 담고 tools 줄에 `Write`, `Edit`, `Bash`가 없다. Antigravity tools는 `read_file`, `grep_search`, `glob`, `mcp_rea_*` 순서의 4개, OMP tools는 `glob`, `grep`, `read`이고 `mcp_`로 시작하는 값이 없으며, OpenCode permission은 `"*": deny`와 allow `glob`, `grep`, `question`, `read`, `skill`, `task`뿐이다. Codex 파일은 `model = "gpt-6.1-sol"`, `model_reasoning_effort = "xhigh"`다
And 다섯 파일의 본문은 다섯 절 이름(`## Behaviour`, `## Inputs and Outputs`, `## Evidence`, `## Unknowns`, `## Acceptance Items`), ID 형식 `B-1`·`A-1`, `## Runtime Capture Request`(target, tool, will run, why), Form Text의 두 미등록 문장을 담고 `AskUserQuestion`을 담지 않는다

### S10: 프리셋 표와 분류
Priority: Must
Given `DefaultFullConfig`, `quality.default: ultra`인 사본, HB6 map(v0.50.125 balanced 16개), HU3 map(v0.50.125 ultra 16개), HB6 + `reverse-engineer: sonnet`인 map이 있다
When presets를 읽고 `ClassifyBuiltinQualityPreset`, `ClaudeAgentProfile`, `CodexAgentProfile`, `NativeBalancedAgentCandidate`를 부르고 compose 그룹을 렌더한다
Then 예상 값: balanced·ultra는 17개이고 reverse-engineer는 둘 다 `opus`다. balanced에서 Claude는 (`claude-opus-5-5`, `high`), Codex는 (`gpt-6.1-sol`, `xhigh`), 후보 selector는 `anthropic/claude-opus-5-5`와 `openai-codex/gpt-6.1-sol`이다. ultra에서 Claude는 (`opus`, `max`), Codex는 (`gpt-6.1-sol`, `xhigh`)다
And HB6 balanced는 class `historical`, Row `HB6`, RuntimeStandard true, Differs 1이고 그 project의 reverse-engineer도 (`claude-opus-5-5`, `high`)를 받는다. HU3 ultra는 `historical`·`HU3`, 17개 기본값은 `current`, HB6 + `reverse-engineer: sonnet`은 `customized`다
And 그룹 제목은 `탐색·보조 — explorer, annotator, reverse-engineer`, 기본값의 현재 값은 `혼합 (opus 1 · sonnet 2)`이고 탐색·보조에 `haiku`를 고르면 세 에이전트 모두 `haiku`가 된다. `balancedRungByAgent`의 key 집합은 `CanonicalAgentNames()`와 같다

### S11: OMP routing 17 + task와 fail closed
Priority: Must
Given P5의 OMP에 `role_model_policy {version: v1, profile: balanced, family: openai}`를 둔 project, SPEC-MODELUX-003 IO 생성 runner(catalog = Existing Test Changes 규칙 7로 두 model을 더한 `ompCLIProfileNativeCatalogJSON()`), 그 사본에 `agent_source: native`를 더한 것, catalog에서 `openai-codex/gpt-6.1-sol`을 뺀 사본이 있다
When 각각 `auto update --yes`를 실행하고 같은 run의 plan 미리보기와 `task.agentModelOverrides` 되읽기를 기록한다
Then autopus 모드 override key는 정규 17개와 `task`, 모두 18개이고 `reverse-engineer` 값은 `openai-codex/gpt-6.1-sol:xhigh`이며 나머지 17개 값은 같은 run의 미리보기와 같다
And `config.OMPNativeAgentForRole("reverse-engineer")`는 `task`이고 native 사본의 override key는 native 5개 그대로이며 `.omp/agents`에 생성 파일이 없다. model을 뺀 사본은 `omp_profile_candidate_unavailable`과 `openai-codex/gpt-6.1-sol`을 담은 오류로 끝나고 OMP 파일·overlay bytes가 실행 전과 같다
And autopus 모드 OMP manifest의 `.omp/agents/` 항목은 17개이고, OMP 생성 전 project의 migration 미리보기는 `+ .omp/agents/reverse-engineer.md`를 포함한 17줄과 `OMP는 이제 Autopus 에이전트 17개를 씁니다.`로 시작하는 줄을 담는다

### S12: MCP 참조 변환 예외
Priority: Must
Given 본문 다섯 줄 L1 `Call mcp__rea__list_unknowns, then mcp__rea for the server.`, L2 `Use mcp__context7__query-docs("react", topic="hooks").`, L3 `Fallback mcp__other__y.`, L4 `Mixed mcp__rea__inspect_analysis_view and mcp__other__y.`, L5 `Prefix mcp__reader__x.`가 있다
When `ReplacePlatformReferences(body, p)`를 p = codex, gemini, opencode, omp, claude로 부른다
Then 예상 출력은 codex·gemini·opencode·omp에서 L1 그대로, L2 `Use Context7 MCP first; fallback: WebSearch "react hooks docs" (topic: hooks).`, L3 `Fallback Context7 MCP with WebSearch fallback.`, L4 `Mixed mcp__rea__inspect_analysis_view and Context7 MCP with WebSearch fallback.`, L5 `Prefix Context7 MCP with WebSearch fallback.`이고 claude에서는 다섯 줄 모두 그대로다. `make generate-templates` 뒤 Codex·Gemini reverse-engineer 산출물은 본문의 `mcp__rea`를 그대로 담는다

### S13: router, command, reverse detail, handoff 문장
Priority: Must
Given `DefaultFullConfig`로 다섯 플랫폼을 `Generate`한 project가 있다
When router, `.gemini/commands/auto/reverse.toml`, 다섯 `auto-reverse` detail, `.claude/agents/autopus/executor.md`, `tester.md`를 읽는다
Then Claude root router(`.claude/skills/auto/SKILL.md`)는 `.claude/skills/auto-reverse/SKILL.md`를 정확히 1번 담고 8192 bytes 이하다. Antigravity router와 `.gemini/commands/auto/reverse.toml`은 `.gemini/skills/autopus/auto-reverse/SKILL.md`를 담는다. Codex router 설명의 예상 값은 `Autopus 명령 라우터 — setup/status/goal/update/plan/go/fix/review/sync/idea/map/why/verify/secure/test/qa/dev/canary/doctor/reverse 서브커맨드를 해석합니다`다
And 다섯 detail은 각각 `--runtime`, `Runtime capture wanted: yes`, `## Runtime Capture Request`, `Runtime capture consent:`, `허용 / 거절`, `확인 / 수정 / 중단`, `--auto도 이 확인 gate를 건너뛰지 않습니다`, slug 예 `로그인 화면` → `feature`, `.autopus/reverse/<slug>/behaviour-spec.md`, `SPEC: N/A`, `Task: reverse/<slug>`, `Behaviour Spec:`, `Phase: Test Scaffold`, `Phase: Testing`, `dispatch prompt에 REA 결과, evidence bundle, 디컴파일 텍스트를 넣지 않습니다`를 담는다
And executor.md·tester.md는 `SPEC: N/A`이고 `Behaviour Spec:` 경로가 있으면 그 파일이 요구사항 원천이라는 문장을 담고, frozen 목록에 `reverse`를 더한 Claude·Antigravity router·command test와 Update·Generate parity test가 통과한다

### S14: 사전 승인 없음
Priority: Must
Given S9 project와 `DetectPermissions` 결과가 있다
When `.claude/settings.json`, `.codex/config.toml`, `.gemini/settings.json`, `.omp/`, `.opencode/` 아래 생성 파일을 읽는다
Then `permissions.allow`의 MCP 값은 `mcp__sequential-thinking__sequentialthinking` 하나이고 `mcp__rea`로 시작하는 값이 없으며, 어떤 생성 파일에도 `enableAllProjectMcpServers`, `enabledMcpjsonServers`, `enabled_tools`, `default_tools_approval_mode`가 없다

### S15: 꺼져 있으면 bytes 그대로
Priority: Must
Given platforms 3개인 `config.DefaultFullConfig("p")`(결정 없음), BASE가 같은 cfg로 만든 세 파일, OLD와 그 사본이 있다
When 이 branch에서 Claude·Codex·Antigravity `Generate`와 `Update`를 돌리고, OLD는 이 branch로, 사본은 BASE로 `auto update --yes`한다
Then `.mcp.json`, `.codex/config.toml`, `.gemini/settings.json`은 BASE 결과와 bytes가 같고, OLD의 세 파일은 사본의 세 파일과 bytes가 같으며 `rea`가 없다

### S16: 사용자 범위 무접촉과 symlink 경계
Priority: Must
Given HOME sandbox 두 개와 그 파일 목록·SHA-256, N24, P3 사본 넷: (a) `.gemini/settings.json`이 sandbox 밖 파일 O1을 가리키는 symlink, (b) `.gemini`가 밖 디렉터리 O2를 가리키는 symlink, (c) `.mcp.json`이 O3 symlink, (d) `.codex`가 O4 symlink가 있다
When 평범한 P3에서 이 branch로 `auto init --yes`, `auto update --yes`, `auto doctor`, `auto platform remove codex`를, 같은 명령을 BASE로 다른 sandbox에서 실행하고, 이 branch로만 `auto update --yes --rea`와 `--no-rea`를 실행한다. (a)–(d)에서 `auto update --yes --rea`, `--no-rea`, (a)·(b)에서 `auto platform remove antigravity-cli`를 실행한다
Then BASE가 지원하는 네 명령의 sandbox 파일 집합 차이는 두 sandbox에서 같고, `--rea`·`--no-rea` 실행 전후 branch sandbox의 파일 집합과 SHA-256은 같으며 `.claude.json`, `.codex/config.toml`, `.gemini/settings.json`, `.omp/agent/mcp.json`이 sandbox에 생기지 않는다 그리고 (a)–(d)에서 REA 내용을 바꾸는 실행은 그 플랫폼이 symlink 경로를 담은 오류로 끝나고 O1–O4의 bytes는 실행 전과 같다

### S17: 회귀
Priority: Must
Given 이 branch와 기준 commit 4d73aa91의 test가 있다
When `go test ./...`를 필터 없이 돌리고, `git diff --stat 4d73aa91 -- '*_test.go' '*.golden' 'evals/'`를 T2 분류와 맞추며, `auto eval harness run`, `GOOS=windows GOARCH=amd64 go build ./...`, 신규·변경 `.go` 줄 수 확인, `make generate-templates` 뒤 `git status --porcelain`을 실행한다
Then 모든 test가 통과하고 바뀐 기존 test는 T2 목록과 Existing Test Changes 규칙 1–8대로이며, eval harness 실패 0, Windows build 성공, 신규·변경 `.go`는 모두 300줄 이하(`update.go` 포함), generate 뒤 반영 안 된 diff 0이다

### S18: 실제 터미널 smoke (모델 호출 없음, Codex 먼저)
Priority: Must
Given macOS 실제 터미널, 이 branch의 `auto`, Node v24.x, 버전을 기록한 Claude Code·Codex·Antigravity CLI, `git init`한 빈 temp dir, 도구마다 새로 만든 빈 cache 디렉터리 `<scratch>/npm-codex`·`npm-claude`·`npm-agy`, Codex용 새 scratch HOME과 CODEX_HOME, `script -q` 기록이 있다
When `auto init --platforms claude-code,codex,antigravity-cli`에서 REA 단계를 Enter로 넘기고, 다른 도구를 띄우기 전에 `npm_config_cache=<scratch>/npm-codex`와 scratch HOME으로 trusted `codex`를 메시지 없이 열어 rea 도구가 보일 때까지의 시간을 잰 뒤, `npm_config_cache=<scratch>/npm-claude`로 `claude`를 열어 project MCP 승인 질문에서 rea를 승인하고 `/mcp`만 보고, `npm_config_cache=<scratch>/npm-agy`로 Antigravity CLI의 MCP 목록을 본다
Then Codex 실행 전 `<scratch>/npm-codex`와 scratch HOME의 `.npm`에 rea-agents가 없고 실행 뒤 그중 하나에 `rea-agents@6.5.0`이 생기며(cold 증명), 30초 안에 rea 도구가 보인다(A2 시작 절반 PASS). cold를 증명하지 못하거나 30초를 넘으면 시간과 cache 상태를 적고 FAIL이다. Claude `/mcp`에 rea가 connected이고 도구 목록이 보이며 Antigravity 목록에 rea가 connected다 그리고 transcript를 `$HOME`, scratch 경로, 사용자 이름을 치환해 `evidence/manual-smoke-rea.txt`에 남긴다. 하나라도 다르면 CD-1이 남는다

### S19: 동의 뒤 live `/auto reverse`
Priority: Must
Given S18 project, 사용자의 모델 호출 동의, 외부 디컴파일러가 필요 없는 작은 로컬 정적 대상(JS 또는 Electron 앱 하나), 실행하면 계산한 한 줄을 출력하는 로컬 fixture `fixture/hello.js`, `--no-rea`로 REA를 끈 사본이 있다
When Claude Code에서 (a) `/auto reverse <대상> <기능>` → 확인, (b) 수정 한 번 뒤 중단, (c) `/auto reverse fixture/hello.js "실행 출력" --runtime` → 동의 질문에서 거절, (e) 같은 명령 → 허용 → 확인, (d) 끈 사본에서 `/auto reverse`를 실행하고, Antigravity와 Codex에서 각각 `/auto reverse`를 gate까지 실행한 뒤 중단한다
Then (a)에서 reverse-engineer subagent가 `mcp__rea__` 도구를 1번 이상 부르고(A1 Claude PASS), 결과가 다섯 절·`B-`·`A-` ID·evidence_id 1개 이상을 담고 의사코드·assembly가 없다. 새 project 파일은 `.autopus/reverse/<slug>/behaviour-spec.md`와 gate에서 확인한 경로 아래 파일뿐이고 `git status --porcelain`은 `.autopus/reverse/`를 보이지 않으며, 세 dispatch prompt는 `SPEC: N/A`, `Task: reverse/<slug>`, `Behaviour Spec:`, ID, 경로, Phase를 담고 REA 결과 텍스트가 없으며, tester Testing은 `A-`마다 판정을 남긴다
And (b)의 수정은 분석을 한 번 다시 보내고 중단 뒤에는 파일이 생기지 않고 dispatch가 더 없다. (c)와 (e)의 첫 결과는 `## Runtime Capture Request`이고 그때까지 runtime capture 도구 호출은 0번이다. (c)는 거절 뒤에도 호출이 0번이고 `## Unknowns`에 빈 곳이 적힌다. (e)는 허용 뒤 다시 보낸 dispatch prompt가 `Runtime capture consent: <tool> on fixture/hello.js`를 담고 그 뒤 runtime capture 도구 호출이 1번 이상이다(REQ-REA-21 허용 경로). (d)는 dispatch가 정확히 1번(reverse-engineer)이고 그 결과가 미등록 문장으로 끝나며 executor·tester dispatch와 새 파일이 0이다
And Antigravity의 reverse-engineer는 `mcp_rea_*`로 해석된 REA 도구를 1번 이상 부르고(A1 Antigravity PASS), Codex의 reverse-engineer는 REA 도구 호출 1번 이상을 60초 안에 끝낸다(A2 도구 절반 PASS). 결과를 `evidence/manual-smoke-rea.txt`에 남기고, 동의가 없거나 하나라도 FAIL이면 PASS로 닫지 않고 CD-2를 남긴다

### S20: locale
Priority: Should
Given OLD 사본과 N20이 있다
When `composeLanguage` seam을 `en`과 `ko`로 두고 non-TTY `auto update`와 `auto update --yes --rea`를 실행한다
Then en의 예상 출력은 `REA reverse-engineering MCP is undecided; run auto update --rea to enable or auto update --no-rea to decline`, `REA MCP not registered: node v20.11.0 is outside ^22.19.0 || ^24.11.0 || >=26.0.0`, `auto older than v0.50.126 cannot read features.rea`이고 ko는 Form Text의 한국어 문장이며, 두 언어의 autopus.yaml 결과는 bytes가 같다

### S21: 문서
Priority: Should
Given help, README, CHANGELOG가 있다
When `auto init --help`, `auto update --help`, `auto doctor --help`를 실행하고 문서를 읽는다
Then REQ-REA-19의 항목(기본 on, flag, Node 범위, lazy key와 옛 바이너리, v0.50.125가 17개 preset을 customized로 보는 점, OMP balanced fail closed와 pins-only 탈출, Windows의 POSIX 실행 형식 한계, clean-room 흐름, runtime capture 동의)이 있고, 기존 test가 고정한 문장이 그대로이며, `make generate-templates` 뒤 반영 안 된 diff가 없다

## Oracle Acceptance Notes

- Must 시나리오는 concrete expected output을 판정한다: 분류 문자열, `data.rea` 예상 JSON, 출력 줄과 횟수, frontmatter·TOML 예상 값, 파일 bytes와 SHA-256, 호출 수, exec 결과 판정. 파일 존재나 종료 코드만으로 닫는 Must는 없다.
- JSON 사용자 항목의 oracle은 값(deep-equal)과 "BASE와 같은 정규화"(S5의 R)이고, Codex 사용자 table과 off 뒤 파일은 bytes다. 이는 MCP Entry Shape Contract의 보존 규칙과 같다.
- 기대 값은 2026-10-15에 기준 commit 4d73aa91 코드, Claude Code·Codex·Gemini CLI 문서, npm `rea-agents@6.5.0`, auto 0.50.123 실행(`evidence/probe-rea-forward-compat.md`)으로 계산했다. 코드 사실과 다르면 test가 아니라 이 SPEC을 고친다. S18·S19는 실제 터미널과 모델 호출이 필요하고, 기록 전이나 probe 행이 FAIL이면 `research.md`의 CD-1·CD-2가 남는다.
