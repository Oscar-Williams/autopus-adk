# Probe: 세션 안에서 팀 능력을 결정적으로 읽을 수 있는가, 옛 auto가 새 key를 읽는가 (2026-10-15)

- 실행 위치: 대화형 Claude Code 세션이 띄운 subagent의 Bash tool. 모델 호출 없음, 네트워크 없음.
- 비밀 값은 출력하지 않았다. 환경 변수는 이름만 나열하고 아래 표의 값만 읽었다. 경로는 `<home>`, `<scratch>`로 바꿨다.

## C. 능력 신호 (Bash tool child process)

| ID | 명령 | 관측 |
|---|---|---|
| C1 | `env` 이름 목록(`^(CLAUDE|ANTHROPIC|AUTOPUS)`) | `CLAUDECODE`, `CLAUDE_CODE_CHILD_SESSION`, `CLAUDE_CODE_ENTRYPOINT`, `CLAUDE_CODE_EXECPATH`, `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS`, `CLAUDE_CODE_MESSAGING_SOCKET`, `CLAUDE_CODE_MESSAGING_TOKEN`, `CLAUDE_CODE_SESSION_ATTENDED`, `CLAUDE_CODE_SESSION_ID`, `CLAUDE_DISABLE_PATH_WARNING`, `CLAUDE_EFFORT`, `CLAUDE_PID`. `CLAUDE_AGENT_SDK_VERSION`은 없다 |
| C2 | 값 읽기 | `CLAUDECODE=1`, `CLAUDE_CODE_ENTRYPOINT=cli`, `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`, `CLAUDE_CODE_SESSION_ATTENDED=1`, `CLAUDE_CODE_CHILD_SESSION=1`(이 probe는 subagent 안에서 돌았다) |
| C3 | `"$CLAUDE_CODE_EXECPATH" --version`, `file -b`, `claude --version` | EXECPATH basename `claude.exe`, `Mach-O 64-bit executable arm64`, 출력 `2.1.288 (Claude Code)`. PATH의 `claude`(`<home>/.nvm/versions/node/v24.20.0/bin/claude`)는 `2.1.299 (Claude Code)`다. PATH 버전은 실행 중인 세션 버전이 아니다 |
| C4 | `[ -t 0 ]`, `[ -t 1 ]` | 둘 다 TTY가 아니다. Bash tool 안에서 TTY로는 대화형 여부를 알 수 없다 |
| C5 | EXECPATH 바이너리 문자열 | `CLAUDE_CODE_ENTRYPOINT`가 여러 곳에서 읽히고 `sdk-cli` 문자열이 11번 나온다. print mode 값의 정적 단서일 뿐 실행 확인은 아니다(A2) |

## F. 옛 auto와 `workflow.auto_team` (`auto 0.50.125`, PATH)

격리: `<scratch>/home`을 HOME으로, `<scratch>/proj`에서 `git init` 뒤 `CI=true auto init --yes --platforms claude-code`. 생성된 `autopus.yaml`의 `workflow:` 블록은 4칸 들여쓰기 `coverage_threshold: 85` 한 줄이다.

| ID | 입력 | 명령 | 관측 |
|---|---|---|---|
| F1 | `workflow:` 아래 `auto_team: false` | `auto doctor` | ``[ERROR] autopus.yaml 로드 실패: parse config <scratch>/proj/autopus.yaml: yaml: unmarshal errors:`` |
| F2 | 같음 | `auto update --yes` | 종료 1, ``line 163: field auto_team not found in type config.WorkflowConf (unknown keys are rejected: fix the typo or delete the key)`` |
| F3 | key 없는 init 결과 | `auto update --yes`, `grep -c auto_team` | 종료 0, `0` |
| F4 | `{}` facts | `auto workflow triage --topology --facts-json <scratch>/f.json --format json` | 종료 1, `Error: unknown flag: --topology` |

결론: 신호 넷(CLAUDECODE, AGENT_TEAMS, ENTRYPOINT·ATTENDED, EXECPATH 버전)은 세션 안 child process에서 결정적으로 읽힌다. 버전은 PATH가 아니라 EXECPATH에서 읽어야 한다. typed key는 있으면 옛 auto를 모두 막으므로 기본값은 key 없음(=on)이고 바이너리는 key를 스스로 쓰지 않는다.

## Y. yaml.v3 v3.0.1 bool decode (review round 1, 2026-10-15)

worktree module(`go.mod`: `gopkg.in/yaml.v3 v3.0.1`) 안의 임시 `main`으로 `workflow:\n  auto_team: <v>`를 `KnownFields(true)` decoder로 `AutoTeam *bool`에 읽었다. 임시 파일은 지웠다.

| `<v>` | node ShortTag | 결과 |
|---|---|---|
| `false`, `False` / `true`, `TRUE` | `!!bool` | false / true |
| `!!bool "false"` | `!!bool` | false |
| `"false"`, `"true"` | `!!str` | 오류 ``line 2: cannot unmarshal !!str `false` into bool`` |
| `off`, `"off"`, `no` | `!!str` | 오류 없이 false |
| `yes`, `"yes"`, `y` | `!!str` | 오류 없이 true |
| `null`, `~`, 빈 값 | `!!null` | 오류 없이 nil(=기본 on) |
| `1`, `0` | `!!int` | 오류 ``cannot unmarshal !!int `1` into bool`` |

marshal: `false` → `workflow:\n    auto_team: false\n`, nil → `workflow: {}\n`(key 없음). 결론: `KnownFields`와 `*bool`만으로는 "bool이 아닌 값 거부"가 성립하지 않는다(YAML 1.1 호환 문자열과 null이 통과). raw node의 ShortTag가 `!!bool`인지 따로 확인해야 한다.

## R. root 오류 출력과 쓰기 없는 load (review round 1)

| ID | 관측 |
|---|---|
| R1 | v0.50.125 `auto workflow triage --facts-json /nonexistent-teamauto` → 종료 1, stderr bytes `Error: facts file must be a readable regular file at most 1 MiB\n`. `internal/cli/root.go:203-211` `runRootCommand`가 `Error: %s\n`을 쓰고 `exitCodeForError`(`:173-179`)의 일반 값은 1이다 |
| R2 | `pkg/config/loader.go:16-19` `Load`는 `loadConfig(dir, true)`라 platform 이름을 정규화하면 `os.WriteFile`로 autopus.yaml을 다시 쓴다(`:90-95`). `:22-25` `LoadPreview`는 `loadConfig(dir, false)`로 쓰지 않는다. 파일이 없으면 둘 다 `DefaultFullConfig`를 오류 없이 돌려준다(`:63`) |
