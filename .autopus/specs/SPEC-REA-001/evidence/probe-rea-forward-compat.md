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
