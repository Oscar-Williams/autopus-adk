# Probe A3: 옛 auto가 이 SPEC이 쓰는 파일을 읽는가 (2026-10-15)

- 바이너리: PATH의 `auto 0.50.123 (commit: c447badc, built: 2026-10-04T09:37:09Z)`. strict loader(`reservedConfigKeys`, `KnownFields(true)`)는 `git show v0.50.113:pkg/config/loader_strict.go`부터 있고 `v0.50.112`에는 없다. v0.50.125도 같은 구조다.
- 격리: `<scratch>/home`을 HOME으로 두고 `<scratch>/proj`에서 `git init` 뒤 `CI=true auto init --yes --platforms claude-code`(P3는 `claude-code,codex,antigravity-cli`). init 뒤 HOME에 파일이 생기지 않았다. 네트워크 호출 없음.

| ID | 입력 | 명령 | 관측 |
|---|---|---|---|
| V1 | autopus.yaml `features:` 아래 `rea:` / `enabled: true` | `auto doctor` | 종료 0, ``[ERROR] autopus.yaml 로드 실패: parse config <scratch>/v1/autopus.yaml: yaml: unmarshal errors: line 160: field rea not found in type config.FeaturesConf (unknown keys are rejected: fix the typo or delete the key)`` |
| V1 | 같음 | `auto update --yes` | 종료 1, ``Error: 설정 로드 실패: … line 160: field rea not found in type config.FeaturesConf …`` |
| V2 | 파일 끝 `future_extension:` / `rea:` / `enabled: true` | `auto doctor` | 종료 0, `[OK] autopus.yaml (mode: full)` |
| V2 | 같음 | `auto update --yes` | 종료 0. 뒤 autopus.yaml에 `future_extension`이 없다(typed Save가 지웠다) |
| P3 | 세 파일에 SPEC의 managed 모양 rea 항목(Claude `$comment` 포함, Antigravity, Codex 세 줄) | `auto update --yes` | 종료 0. `.mcp.json`·`.gemini/settings.json`의 `rea` 값이 넣은 값과 같고 `.codex/config.toml`의 `[mcp_servers.rea]`와 세 줄이 그대로다 |

## yaml.v3 decode (scratch module, gopkg.in/yaml.v3 v3.0.1, `KnownFields(true)`, `Enabled *bool`)

| 입력 | 결과 |
|---|---|
| `rea: {enabled: "true"}` | ``line 2: cannot unmarshal !!str `true` into bool`` |
| `rea: {enabled: 1}` | ``line 2: cannot unmarshal !!int `1` into bool`` |
| `rea: {enabled: yes}` | true |
| `rea: {enabled: false}` | false, 다시 marshal하면 `enabled: false` |
| `rea: {}` | nil |

## 문서와 registry (2026-10-15)

- Claude Code subagents(WebFetch `code.claude.com/docs/en/sub-agents`): `tools`는 `mcp__<server>`와 `mcp__<server>__*`를 받는다. 해석되는 항목이 하나도 없으면 실행되지 않는다. background subagent는 MCP 도구를 모두 받고 built-in 대부분을 잃는다. permissionMode 값은 default, acceptEdits, auto, dontAsk, bypassPermissions, plan.
- Codex MCP(WebFetch `learn.chatgpt.com/docs/extend/mcp?surface=cli`): STDIO `startup_timeout_sec` 기본 10초, `tool_timeout_sec` 기본 60초, project `.codex/config.toml`은 trusted project만.
- Gemini CLI subagents(WebFetch `github.com/google-gemini/gemini-cli` docs): `tools` wildcard `*`, `mcp_*`, `mcp_server_*`. 생략하면 부모 세션 도구를 모두 상속한다.
- npm(`npm view rea-agents@6.5.0`): MIT, `engines.node` `^22.19.0 || ^24.11.0 || >=26.0.0`, bin `rea`·`rea-agents` → `scripts/rea.mjs`, `dist-tags.latest` 6.5.0. REA README: static JS·.NET 분석은 대상을 실행하지 않고 runtime capture는 사용자 권한으로 대상을 실행한다. README에는 `rea setup`의 Codex timeout 값이 없다.

결론: `features.rea`를 늘 쓰면 새 프로젝트마다 옛 auto의 모든 명령이 막힌다. reserved key는 옛 update가 지운다. managed 항목은 옛 update를 지나도 남는다. 그래서 결정은 항목이 담고 key는 항목이 결정을 담지 못할 때만 쓴다.

## npm project config (보안 감사 L4, 2026-10-27)

- npm: 11.20.0(`/bin/bash -l`이 찾는 Homebrew node v26.12.1), 11.19.0(nvm node v24.20.0), 10.9.8(nvm node v22.23.0). 소스 확인은 `@npmcli/config` `loadLocalPrefix`·`loadProjectConfig`, `libnpmexec` `run-script.js`("we always run in cwd, not --prefix")다.
- 방법: 행마다 새 scratch HOME(빈 cache, user config는 stub registry를 가리키는 `registry=` 한 줄)과 그 안의 project에서 `npx …`를 한 번 돌렸다. stub registry(127.0.0.1의 로컬 HTTP 서버)는 가짜 `rea-agents@6.5.0`과 그 의존성 가짜 `@modelcontextprotocol/server@2.3.1`을 낸다. project 쪽 `.npmrc`는 닫힌 port `127.0.0.1:47109`를 가리킨다. 모든 행에 같은 env `npm_config_fetch_retries=0`·`npm_config_audit=false`·`npm_config_update_notifier=false`·`npm_config_fund=false`를 줬다(registry 해석에 들지 않는 key). stub이 요청을 받고 가짜 bin이 돌면 그 열의 설정이 이겼고, 닫힌 port에 `ECONNREFUSED`면 project 쪽이 이겼다.

| project 쪽 파일 | 지금 모양 `npx -y rea-agents@6.5.0 mcp` | `--registry=<stub>` | env `npm_config_registry=<stub>` | `npx -y --prefix=/ rea-agents@6.5.0 mcp` |
|---|---|---|---|---|
| `.npmrc` `registry=` | 닫힌 port(`/rea-agents`) | stub | stub | stub |
| `.npmrc` `@modelcontextprotocol:registry=` | 닫힌 port(`/@modelcontextprotocol%2fserver`) | 닫힌 port | 닫힌 port | stub |
| 상위 디렉터리 `.npmrc`(상위에 `package.json`, project에는 없음) | 닫힌 port | 측정 안 함 | 측정 안 함 | stub |
| `.npmrc` `userconfig=./u.npmrc`, u에 `registry=` | 닫힌 port | stub | 측정 안 함 | 측정 안 함(아래 행) |
| 같음, u에 `registry=`와 scope 줄 | 측정 안 함 | 닫힌 port(scope 요청) | 측정 안 함 | stub |
| `.npmrc` `globalconfig=./g.npmrc`, g에 `registry=`와 scope 줄 | 닫힌 port(user config 없음) | 측정 안 함 | 측정 안 함 | stub |

- `--prefix=/` 행: 가짜 bin의 `process.cwd()`가 project 디렉터리였고, project 디렉터리에 파일이 생기지 않았으며, npx cache 디렉터리 이름(`_npx/ad9a3fdbcad992c6`)이 지금 모양과 같다. npm 11.19.0은 `--prefix=/`의 registry·scope·상위 디렉터리 세 행이, 10.9.8은 지금 모양의 registry·scope 행, `--registry=`의 scope 행, `--prefix=/`의 registry·scope·상위 디렉터리·userconfig(scope 줄 포함) 행이 11.20.0과 같다.
- 부작용: `npm config get globalconfig`가 `--prefix=/`에서 `/etc/npmrc`다(기본은 Homebrew `/opt/homebrew/etc/npmrc`, nvm `<nvm>/versions/node/v24.20.0/etc/npmrc`). 이 기계에는 세 파일이 모두 없다.
- 실제 package: scratch HOME의 npm cache에 기존 `rea-agents@6.5.0` npx 설치를 복사해 tarball을 받지 않고, project `.npmrc`가 `registry=`와 `@modelcontextprotocol:registry=`로 닫힌 port를 가리키는 디렉터리에서 실제 항목 명령을 돌렸다. `/bin/bash -l -c "exec npx -y --prefix=/ rea-agents@6.5.0 mcp"`는 MCP `initialize`에 1.8초 만에 `serverInfo {"name": "rea", "version": "6.5.0"}`로 답했고, `npm exec`와 REA node 프로세스의 cwd가 project 디렉터리였다. cache에 남은 요청은 `https://registry.npmjs.org/rea-agents` 하나다. 같은 디렉터리에서 지금 모양 `exec npx -y rea-agents@6.5.0 mcp`는 warm npx cache인데도 `GET http://127.0.0.1:47109/rea-agents … ECONNREFUSED`로 끝났다(npx는 실행마다 manifest를 registry에서 다시 받는다).

결론: registry만 고정하는 `--registry=`와 `npm_config_registry`는 project의 scope 줄을 막지 못한다. command line `--prefix=/`는 project·상위 디렉터리 `.npmrc`를 통째로 읽지 않아 scope 줄과 `userconfig=`·`globalconfig=` 우회까지 닫고, 사용자 `~/.npmrc`와 서버의 cwd는 그대로 둔다.
