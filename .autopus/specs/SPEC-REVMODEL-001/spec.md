# SPEC-REVMODEL-001: 멀티프로바이더 리뷰 모델 구성 탭

**Status**: draft
**Created**: 2026-10-15
**Domain**: REVMODEL
**Module**: autopus-adk
**Release**: v0.50.126(SPEC-MODELUX-003, SPEC-REA-001, SPEC-TEAMAUTO-001, redline capability와 함께). 작업 branch는 `feat/revmodel-001`이고 기준은 `feat/rea-001`(ee811e2e)이다. SPEC-REA-001이 구성 UI의 에이전트 수를 17개로 바꿨기 때문이다.
**Predecessor**: SPEC-MODELUX-002(폼 흐름, Form Flow Contract, Write Safety Contract), SPEC-MODELUX-003(OMP Autopus Tab의 증명된 model 목록). 그 정의를 그대로 쓰고 다르면 이 SPEC이 우선한다.
**Source PRD**: `prd.md`. 근거, probe, 결정은 `research.md`와 `plan.md` `## Risk-First Integration Probe`에 있다.

## 목적

멀티프로바이더 리뷰(`auto spec review`)와 orchestra 명령이 어느 provider를 쓰고 provider마다 어느 model·thinking으로 돌며 누가 judge인지는 지금 `autopus.yaml`을 손으로 고쳐야만 바뀐다. init wizard에는 Review Gate 예/아니오 단계만 있다(`internal/cli/tui/wizard_steps.go` `buildReviewGateStep`).

2026-10-15 실측에서 OMP catalog는 `google-antigravity/gemini-3.1-pro`를 보였지만 한 번 호출하면 HTTP 404였다. 그래서 그날의 SPEC 리뷰는 gemini 없이 돌았고 receipt의 `degraded_reasons`는 비어 있었다. catalog에 있다는 것만으로는 model이 실제로 답한다고 볼 수 없다.

이 SPEC은 SPEC-MODELUX-002·003의 구성 UI(`auto init`·`auto update`·`auto quality`)에 `멀티프로바이더 리뷰` 탭을 더한다. 이 탭에서 정하는 것은 다섯 가지다: (1) 리뷰에 참여하는 provider(켜기·끄기), (2) OMP backend provider의 model과 thinking(증명된 선택지만), (3) judge(참여 provider만), (4) 적용 전 미리보기, (5) 사용자가 고를 때만 도는 provider당 1회 확인 호출. 쓰는 key는 v0.50.125가 이미 읽는 key뿐이다.

## Outcome Boundary

- Outcome Lock 원문은 `research.md` `## Outcome Lock`이다.
- Mandatory requirements: REQ-RVM-01 ~ REQ-RVM-15, REQ-RVM-19 (Must). REQ-RVM-16 ~ 18은 Should다.
- Explicit non-goals:
  - strategy, timeout, `enabled`, `max_revisions`, `min_providers`, `verdict_threshold` 같은 리뷰 정책 key 편집
  - 참여 여부에서 따라 나오는 것 말고 명령별 provider 목록이나 judge를 따로 편집하는 것
  - CLI backend provider(`backend` 없음)의 model·effort·args·`model_policy` 편집과 CLI→OMP backend 전환(Evolution Idea)
  - `orchestra.providers` 항목 추가·삭제, `backend`·`binary`·`tools`·`subprocess` 변경
  - `--multi`·`--providers` flag 의미 변경, review receipt `degraded_reasons` 규칙 변경(Evolution Idea), `auto doctor --provider-smoke` 변경
  - 확인 호출 자동 실행, 결과 저장(receipt·cache), 새 YAML key
- Completion evidence: Must S1–S12와 S15 PASS, `go test ./internal/cli/... ./pkg/...` 필터 없이 통과, 신규·변경 소스 파일 300줄 이하, `GOOS=windows GOARCH=amd64 go build ./...`, `evidence/manual-smoke-revmodel.txt`(S15), `auto spec validate --strict`. `research.md` `## Completion Debt`의 CD-1이 남아 있으면 완료가 아니다.

## Trust Model

- SPEC-MODELUX-001·002 Trust Model을 그대로 따른다. `autopus.yaml`은 반쯤 신뢰한다. writer는 loader가 환경 변수를 펼친 값을 쓰기 원천으로 쓰지 않고 raw YAML node를 고친다.
- 새 비신뢰 입력은 셋이다. (1) OMP catalog 출력은 기존 `readInstalledOMPModelChoices`의 정규화만 거쳐 쓴다. (2) 확인 호출 응답은 `classifyProviderSmokeResult`가 판정하고 표시 문자열에는 `terminalSafe`를 거친 첫 줄 200 rune까지만 쓴다. (3) provider 이름과 model 문자열은 폼 라벨에 넣기 전에 `terminalSafe`를 거친다.
- 확인 호출은 provider마다 model 호출 1회이고 비용이 든다. 그래서 사용자가 고를 때만 돈다. 새 secret, 승인 단계, 유료 CI는 없다.

## Requirements

### REQ-RVM-01 탭 진입
Priority: Must · EARS: EventDriven

WHEN the operator chooses 구성 편집 in auto init, auto update, or auto quality and the run's config holds at least two `orchestra.providers` entries, THEN THE SYSTEM SHALL list `멀티프로바이더 리뷰` as the last option of the tool tab select and run the Review Tab Contract when it is chosen.

- 진입점은 `chooseTab`(`internal/cli/quality_compose_form_driver.go:155`)과 `edit`(`:133`)의 분기 몇 줄이다. 290줄 파일이라 흐름은 `[NEW] internal/cli/quality_compose_review.go`에 둔다.
- 선택지가 하나뿐이면 tab Select를 건너뛰는 지금 규칙은 그대로다. 이 탭이 더해져 두 개가 되는 fixture의 폼 입력은 Existing Test Changes를 따른다.

### REQ-RVM-02 참여 켜기·끄기
Priority: Must · EARS: EventDriven

WHEN the review tab starts, THEN THE SYSTEM SHALL ask one participation select per `orchestra.providers` entry in `sortedProviderKeys` order, defaulting to 그대로, where a provider is on exactly when `spec.review_gate.providers` names it.

### REQ-RVM-03 provider 목록 쓰기
Priority: Must · EARS: EventDriven

WHEN the operator confirms a review tab edit that turns providers on or off, THEN THE SYSTEM SHALL remove every turned-off provider from `spec.review_gate.providers` and from the `providers` list of every existing `orchestra.commands` entry, append every turned-on provider to the end of each of those lists that lacks it, and leave the membership of every untouched provider unchanged in every list.

- 근거: `auto spec review`는 `--multi`가 없으면 `spec.review_gate.providers`만 쓴다(probe A2, `spec_review_providers.go:39-51`). `--multi`는 계속 `orchestra.commands.review.providers`, `orchestra.providers` key, 기본 provider를 합친다(non-goal).

### REQ-RVM-04 judge
Priority: Must · EARS: EventDriven

WHEN the participation selects are answered, THEN THE SYSTEM SHALL ask the judge among the participating providers only, offer 그대로 only when the current `spec.review_gate.judge` still takes part, otherwise print the judge-off note and default to the first participating provider, and SHALL write the judge by the judge rows of the Config Write Contract.

### REQ-RVM-05 편집 검증
Priority: Must · EARS: Unwanted

IF a review tab edit leaves fewer than two participating providers, names a judge that does not take part, or holds a model or thinking that the attested choices of REQ-RVM-06 do not contain, THEN THE SYSTEM SHALL print the exact error line of the Form Text table, write nothing, and return to the tab's first select.

- 검증은 `[NEW] planReviewComposition`(순수 함수)에 있다. 폼은 이런 값을 내놓지 않지만 같은 함수가 init 메모리 경로와 test 입력도 판정한다. 두 개 이상 규칙은 init이 provider가 둘보다 적으면 Review Gate를 끄는 규칙(`init.go:154-157`)과 같다.

### REQ-RVM-06 증명된 model 선택지
Priority: Must · EARS: EventDriven

WHEN the tab asks the model of a participating provider whose entry has `backend: omp`, THEN THE SYSTEM SHALL offer 그대로 followed by exactly the attested installed models whose `orchestra.ModelFamilyForSelector` family equals the provider's `providerModelFamily` (every attested model when that family is empty), ordered by selector, and SHALL mark 그대로 with `(catalog에 없음)` when the current model is not among them.

- 증명 목록은 SPEC-MODELUX-003의 `installedOMPAttestedModelChoices`(`quality_omp_attested_choices.go:17-37`)를 그대로 쓴다. strict catalog면 보고된 model, metadata 없는 catalog면 built-in 선언에 있는 model만이고 안내 1줄을 쓴다. catalog를 못 읽으면 model Select 없이 안내 1줄을 쓰고 참여·judge 편집은 계속한다.
- family는 런타임이 쓰는 두 함수로 맞춘다: `providerModelFamily`(`orchestra_brainstorm_judge.go:67-79`, claude→anthropic, codex→openai, gemini→google)와 `ModelFamilyForSelector`(`pkg/orchestra/model_family.go:31-43`).

### REQ-RVM-07 thinking과 model 쓰기
Priority: Must · EARS: EventDriven

WHEN the operator picks a model other than 그대로, THEN THE SYSTEM SHALL ask a thinking select over exactly the levels that model reports, defaulting to the current thinking when reported and otherwise to the first reported level, skip the select when the model reports one level or none, and write `orchestra.providers.<p>.model` as `<provider>/<model>:<thinking>`, or `<provider>/<model>` when the model reports no level.

- 근거: OMP backend 런타임은 이 값을 그대로 selector로 쓴다(`omp_review_backend.go:40-44`). Codex 항목이 `backend: omp`이고 `model_policy`가 없으면 `CodexOrchestraProfile`은 적용되지 않는다(probe A1).

### REQ-RVM-08 CLI backend provider
Priority: Must · EARS: StateDriven

WHERE a provider entry has no `backend: omp`, THEN THE SYSTEM SHALL ask no model or thinking select for it, print the CLI-backend note once per provider, and leave its `backend`, `binary`, `args`, `model_policy`, and `model` unchanged, while its participation and judge selects stay available.

- 근거: CLI codex의 model은 `model_policy: quality`일 때 런타임에 `CodexOrchestraProfile`이 args를 덮어 정하고(`orchestra_run_runtime.go:27-35`, `codex_profile.go:93-95`), 기본 gemini(`agy --print`)에는 model flag가 없다(`defaults.go:119`).

### REQ-RVM-09 미리보기
Priority: Must · EARS: EventDriven

WHEN the tab's selects are answered, THEN THE SYSTEM SHALL print the after table and the change lines of the Review Tab Contract in its fixed order before any confirmation, and SHALL print the existing 변경 없음 line and write nothing when the edit equals the start.

### REQ-RVM-10 확인 호출 (opt-in)
Priority: Must · EARS: EventDriven

WHEN the preview of a changed edit has been printed, THEN THE SYSTEM SHALL ask the live check select whose default is 확인 호출 없이 계속, and only when the operator picks 확인 호출 SHALL run the Live Check Contract once per provider of the candidate's review set and print one result line per provider.

### REQ-RVM-11 확인 호출 실패
Priority: Must · EARS: Unwanted

IF any live check result is not 통과, THEN THE SYSTEM SHALL print the live-check failure note, keep the confirmation default on 취소, and write nothing unless the operator explicitly picks 적용.

### REQ-RVM-12 쓰기 안전
Priority: Must · EARS: EventDriven

WHEN the operator applies a review tab edit in auto update or auto quality, THEN THE SYSTEM SHALL write only the keys of the Config Write Contract in one atomic replace of `autopus.yaml` built from the raw document, refuse with the write-refused line and leave the file unchanged when a target node is an alias, a merge key, or holds a `${` placeholder, keep every byte outside the `orchestra` and `spec` sections identical, and then set the run's `cfg.Orchestra` and `cfg.Spec.ReviewGate` from the re-read file.

- writer는 `[NEW] saveReviewComposition`이다. raw document는 `parseAutopusConfigDocument`(`platform_omp_config.go:45`)로 읽고 node만 고친 뒤 `encodeAutopusConfigRewrite`(`:145`)의 parse-back·load 동등성·placeholder 검사를 지난다. 다시 읽기와 cfg 동기화는 OMP 탭의 `ompApply`(`quality_compose_omp.go:263`)와 같다. 그래서 같은 run의 뒤 저장이 이전 구성을 되쓰지 않는다.

### REQ-RVM-13 init 메모리 경로
Priority: Must · EARS: EventDriven

WHEN the operator applies a review tab edit during interactive auto init, THEN THE SYSTEM SHALL apply it to init's in-memory config, which init's save writes, and SHALL apply the wizard's Review Gate answer and the detected provider names before the compose step so that no later init step overwrites `spec.review_gate.providers`.

- 지금 init은 compose 단계 뒤에 `cfg.Spec.ReviewGate.Providers = providerNames`를 한다(`init.go:122-129`). 이 대입을 compose 단계 앞으로 옮긴다. 비대화형 init(`init.go:150-157`)은 그대로다.

### REQ-RVM-14 비대화형 실행
Priority: Must · EARS: Unwanted

IF auto init, auto update, or auto quality runs non-interactively (`--yes`, a non-terminal stdin or stdout, or CI), THEN THE SYSTEM SHALL ask no review tab select, read no OMP catalog for the tab, make no live check call, and write nothing to `autopus.yaml` through this tab.

- 구성 폼은 지금도 `composeFormsTerminal`(`quality_compose_form.go:50-54`)이 참일 때만 열린다. 이 SPEC은 그 gate 밖에 새 입력 경로를 만들지 않는다.

### REQ-RVM-15 옛 바이너리 호환
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL write only keys that v0.50.125 loads (`orchestra.judge`, `orchestra.providers.<p>.model`, `orchestra.commands.<c>.providers`, `orchestra.commands.<c>.judge`, `spec.review_gate.providers`, `spec.review_gate.judge`), add no key outside the v0.50.125 schema, and write only model values that pass `validOMPModelSelector`.

- probe A3: v0.50.125(dd6d6ea2)로 만든 `auto`가 이 모양의 파일을 읽었고(`config get usage_profile` exit 0), 대조군인 모르는 key와 잘못된 selector는 거부했다.

### REQ-RVM-16 메뉴 요약
Priority: Should · EARS: StateDriven

WHERE the review tab is offered, THEN THE SYSTEM SHALL print the review summary line under the composition table each time the menu is drawn.

- ledger의 `brownfield_impact`가 assumed라 Should다. 표 칸은 늘리지 않고 줄 하나만 더한다.

### REQ-RVM-17 locale
Priority: Should · EARS: Ubiquitous

THE SYSTEM SHALL render every new title, option, note, and line of the Form Text table in Korean when tui.DetectSystemLang returns ko and in English otherwise, while written values stay locale-independent.

### REQ-RVM-18 문서
Priority: Should · EARS: Ubiquitous

THE SYSTEM SHALL document the review tab, the keys it writes, the live check and its cost, and the CLI-backend limit in init, update, and quality help, README, and the v0.50.126 CHANGELOG, without removing sentences existing tests check.

### REQ-RVM-19 test seam과 수동 smoke
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL run every new path through the existing seams (composeRunForm, composeFormsProbe, ompCLIFakeRunner, providerSmokeBackendFactory) and SHALL record the real-terminal smoke S15 in `evidence/manual-smoke-revmodel.txt`.

## Contracts

### Review Tab Contract

| 단계 | 규칙 |
|---|---|
| 진입 | tool tab Select의 마지막 선택지 `멀티프로바이더 리뷰`(REQ-RVM-01). 시작 값은 run의 cfg다: update·quality는 파일에서 읽은 값, init은 wizard 답을 넣은 메모리 값 |
| 참여 | provider마다 Select 하나. 켜짐이면 `그대로 — 켜짐`·`끄기`, 꺼짐이면 `그대로 — 꺼짐`·`켜기`. 순서는 `sortedProviderKeys`(`spec_review_helpers.go:49`). `orchestra.providers` 항목이 없는 목록 이름은 보이지 않고 바뀌지 않는다 |
| model | 참여하는 `backend: omp` provider만. `그대로 — <selector>`(지금 값이 증명 목록에 없으면 ` (catalog에 없음)`) 뒤에 같은 family의 증명된 model. catalog 읽기는 탭 진입당 1회이고 모든 provider가 그 결과를 쓴다 |
| thinking | model을 바꿨을 때만, 그 model이 보고한 단계만(REQ-RVM-07) |
| judge | 참여 provider만. 지금 judge가 꺼졌으면 안내 1줄 뒤 `그대로` 없이 첫 참여 provider가 기본이다 |
| 검증 | `planReviewComposition`. 오류면 그 줄을 쓰고 참여 Select로 돌아간다(REQ-RVM-05) |
| 미리보기 | 아래 after 표, 그 뒤 바뀌는 설정 줄. 순서: `spec.review_gate.providers` → `orchestra.commands.<c>.providers`(명령 이름순) → `orchestra.providers.<p>.model`(provider 순) → `spec.review_gate.judge` → `orchestra.judge` → `orchestra.commands.<c>.judge 지움`. 줄 모양은 `<key> <before> → <after>`이고 목록은 `[a, b]`로 쓴다 |
| 확인 호출 | `확인 호출 없이 계속`(기본) · `확인 호출 — 참여 provider <n>개에 model 호출 1회씩`(Live Check Contract) |
| 확인 | 기존 `review`(`quality_compose_form_driver.go:213`)와 같은 `취소`(기본) · `적용` · `다시 편집`. `다시 편집`은 이번 편집 값을 시작 값으로 참여 Select부터 다시 묻는다 |
| 적용 | update·quality는 REQ-RVM-12 writer, init은 REQ-RVM-13 메모리 경로. 성공하면 기존 `적용했습니다` 줄 |

after 표(가로는 `terminalSafe` 뒤 rune 폭으로 맞춘다. 바뀐 칸 끝에 `*`):

```text
provider  리뷰   model                                    judge
claude    켜짐   anthropic/claude-fable-5-1:max           judge
codex     켜짐   openai-codex/gpt-6-luna:high *
gemini    꺼짐 * google-antigravity/gemini-3.1-pro:high
```

CLI backend provider의 model 칸은 `CLI backend`다.

### Config Write Contract

| 선택 | 쓰는 key | 바뀌는 소비처 |
|---|---|---|
| provider 끄기 | `spec.review_gate.providers`와 있는 `orchestra.commands.<c>.providers` 모두에서 그 이름을 지운다 | `auto spec review`(A2), `auto spec review --multi`의 합집합 원천 중 둘, orchestra 명령 brainstorm·plan·review·secure(이 repo 기준, 있는 항목 모두) |
| provider 켜기 | 같은 목록 중 그 이름이 없는 목록 끝에 붙인다 | 같다 |
| 손대지 않은 provider | 어느 목록에서도 바뀌지 않는다(명령별로 다른 지금 구성 유지) | 없음 |
| model (`backend: omp`) | `orchestra.providers.<p>.model` = `<provider>/<model>:<thinking>` | OMP backend로 도는 모든 리뷰·orchestra 호출(A1) |
| judge | `spec.review_gate.judge` = J. `orchestra.judge`는 편집 전 값이 `spec.review_gate.judge`와 같았거나 꺼진 provider면 J로 바꾸고, 그 밖에는 두고, key가 없으면 더하지 않는다 | `auto spec review` judge, debate 명령의 기본 judge |
| 꺼진 provider를 가리키는 `orchestra.commands.<c>.judge` | 그 key를 지운다(`orchestra.judge`로 돌아간다) | 그 명령 |
| 쓰지 않음 | strategy, timeout, `enabled`, `max_revisions`, `min_providers`, `backend`, `binary`, `args`, `model_policy`, `tools`, `subprocess`, provider 항목 추가·삭제, `orchestra.commands` 항목 추가 | — |

- 목록 node나 `spec.review_gate.judge`가 파일에 없으면 v0.50.125 schema에 있는 그 key만 그 mapping에 더한다. 쓰는 목록은 flow sequence(`[a, b]`)다. 이 repo와 `config.Save`가 쓰는 모양이다(`schema_spec.go:14`, `schema_orchestra.go:50`의 `flow` tag).
- 바뀌지 않는 값은 node도 쓰지 않는다. 편집이 시작 값과 같으면 파일을 열어 쓰지도 않는다.

### Live Check Contract

| 항목 | 규칙 |
|---|---|
| 언제 | 미리보기 뒤 사용자가 `확인 호출`을 고를 때만. 비대화형 실행에서는 탭 자체가 열리지 않는다(REQ-RVM-14) |
| 대상 | 후보 구성(메모리, 파일 아님)을 `assembleSpecReviewProviders`(`spec_review_readonly.go:53`)에 넣어 나온 provider 집합. `auto spec review`가 실제로 돌릴 집합이고 read-only 정책과 설치 확인을 거친다. 설치되지 않은 provider는 기존 경고 줄을 쓰고 호출하지 않는다 |
| 호출 | provider마다 1회, 위 순서대로 하나씩. backend는 `providerSmokeBackendFactory`(`doctor_provider_smoke.go:27`, 기본 `selectRoutedBackend`)라 OMP provider는 `ompReviewSession.run`(`omp_review_backend_session.go:31`), CLI provider는 그 binary·args 그대로다. 확인 호출이 더하는 flag는 없다. `OutputFormat`은 `auto doctor --provider-smoke`처럼 `text`다 |
| prompt | `providerSmokePrompt()`(`doctor_provider_smoke.go:123-125`)의 글자 그대로: `Return ONLY this exact token and no other text: AUTOPUS_PROVIDER_SMOKE_OK` |
| timeout | provider마다 60초(`[NEW] composeReviewCheckTimeout`). doctor의 30초 기본값(`doctor.go:106`)은 바꾸지 않는다 |
| 판정 | `classifyProviderSmokeResult`(`doctor_provider_smoke.go:127-144`): `pass`는 통과, `warn`(표시 문자열 없음)과 `fail`은 통과가 아니다 |
| 결과 줄 | `확인 호출 <p> (<model>): 통과` · `확인 호출 <p> (<model>): 응답에 확인 문자열이 없습니다` · `확인 호출 <p> (<model>): 실패 — <detail>`. `<model>`은 OMP면 selector, CLI면 `CLI backend`. `<detail>`은 `terminalSafe`를 거친 첫 줄 200 rune까지 |
| 실패 뒤 | 안내 1줄, 확인 기본값 `취소`(REQ-RVM-11). 결과는 저장하지 않는다 |

### Forward-compat Contract

| 항목 | 규칙 |
|---|---|
| key | REQ-RVM-15의 여섯 경로뿐이다. 모두 v0.50.125 `pkg/config/schema_orchestra.go`·`schema_spec.go`에 같은 yaml tag로 있다(`git show v0.50.125:...`로 확인) |
| 값 | model은 `validOMPModelSelector`(`schema_provider_backend.go:56-75`)를 지나는 값만. 아니면 v0.50.125가 `provider_model_required: <p>`로 파일 전체를 거부한다(A3 대조군) |
| 검증 | S7이 이 branch가 쓴 파일을 v0.50.125 바이너리로 읽는다 |

### Form Text

| 자리 | ko | en |
|---|---|---|
| 탭 | `멀티프로바이더 리뷰` | `Multi-provider review` |
| 참여 | `<p> 리뷰 참여` · `그대로 — 켜짐` · `끄기` · `그대로 — 꺼짐` · `켜기` | `<p> takes part in review` · `keep — on` · `turn off` · `keep — off` · `turn on` |
| model | `<p> model` · `그대로 — <selector>` · ` (catalog에 없음)` | `<p> model` · `keep — <selector>` · ` (not in catalog)` |
| thinking | `<p> thinking (<provider>/<model>)` | 같다 |
| CLI 안내 | `<p>는 CLI backend라 이 탭에서 model을 바꾸지 않습니다` | `<p> runs on a CLI backend; this tab leaves its model unchanged` |
| catalog 안내 | `OMP catalog를 읽지 못해 model은 그대로입니다` | `The OMP catalog is unreadable; models stay unchanged` |
| judge | `judge` · `judge <j>가 꺼져 judge를 다시 고릅니다` | `judge` · `judge <j> was turned off; choose a new judge` |
| 오류 | `리뷰에는 provider가 두 개 이상 필요합니다 (지금 <n>개)` · `judge <j>는 리뷰에 참여하지 않는 provider라 고를 수 없습니다` · `<p> model <selector>는 설치된 OMP catalog가 증명하지 않습니다` · `<p> thinking <t>는 <selector>가 보고하지 않는 단계입니다` | `The review needs at least two providers (now <n>)` · `judge <j> does not take part in the review and cannot be chosen` · `The installed OMP catalog does not attest <p> model <selector>` · `<selector> reports no thinking level <t> for <p>` |
| 확인 호출 | `확인 호출` · `확인 호출 없이 계속` · `확인 호출 — 참여 provider <n>개에 model 호출 1회씩` · 결과 줄(Live Check Contract) · `확인 호출이 실패한 provider가 있습니다. 적용하면 그 provider는 리뷰에서 실패합니다.` | `Live check` · `continue without a live check` · `live check — one model call for each of <n> providers` · `live check <p> (<model>): passed` / `: the reply lacks the check token` / `: failed — <detail>` · `A live check failed; if you apply, that provider keeps failing in reviews.` |
| 쓰기 거부 | `autopus.yaml의 리뷰 설정을 안전하게 고칠 수 없어 아무것도 쓰지 않았습니다: <key>` | `The review settings in autopus.yaml cannot be edited safely; nothing was written: <key>` |
| 메뉴 요약 | `멀티프로바이더 리뷰: claude · codex · gemini(꺼짐) · judge claude` | `Multi-provider review: claude · codex · gemini (off) · judge claude` |

## Existing Test Changes

1. tool tab Select가 선택지 하나라 건너뛰던 폼 test 중 fixture에 `orchestra.providers`가 둘 이상인 것은 이제 그 Select가 열린다. 입력에 그 Select의 기존 기본값을 고르는 Enter(`\n`) 하나를 더한다. 기존 줄 단언은 바꾸지 않는다.
2. init test 중 compose 단계 뒤 `spec.review_gate.providers`를 단언하는 것은 값이 같으므로 바뀌지 않는다(대입 순서만 바뀐다). 바뀌면 T1 목록에 이유를 적는다.
3. 그 밖의 test는 바꾸지 않는다. T1이 파일마다 분류한다.

## 생성 파일 상세

- `internal/cli`: `[NEW] quality_compose_review.go`(+test, 탭 흐름), `[NEW] quality_compose_review_plan.go`(+test, `planReviewComposition`·시작 값·변경 줄·after 표), `[NEW] quality_compose_review_choices.go`(+test, family로 거른 증명 목록), `[NEW] quality_compose_review_write.go`(+test, `saveReviewComposition`·init 메모리 적용), `[NEW] quality_compose_review_check.go`(+test, 확인 호출), `quality_compose_form_driver.go`(분기 몇 줄), `quality_compose_form_text.go`(문구), `init.go`(대입 순서). 메뉴 요약은 `f.draw()` 뒤 한 줄이다.
- 문서: README, CHANGELOG(v0.50.126), help 상수(Should). `content/`·`templates/`를 고치면 `make generate-templates`로 생성 표면을 맞춘다.
- 증거: `[NEW] evidence/manual-smoke-revmodel.txt`. 신규·변경 소스는 300줄 이하다.

## Related SPECs

- SPEC-MODELUX-002(v0.50.125): 폼 흐름, tool tab, 확인 Select, Write Safety Contract를 쓴다. 바뀌는 점은 tool tab 선택지가 하나 늘어난다는 것뿐이다.
- SPEC-MODELUX-003(v0.50.126): `installedOMPAttestedModelChoices`를 그대로 쓴다. OMP 탭(에이전트 route)과 이 탭(리뷰 provider model)은 서로 다른 key를 쓴다.
- SPEC-REVIEWRO-001: OMP review backend와 `auto doctor --provider-smoke` 경로를 확인 호출이 그대로 쓴다.
- SPEC-ADK-REVIEW-INTEGRITY-001: quorum과 `degraded_reasons` 규칙은 바꾸지 않는다(Evolution Idea).
- SPEC-REA-001: 기준 branch. 이 SPEC은 에이전트 표를 바꾸지 않는다.
- Sibling: 없다(`research.md` `## Sibling SPEC Decision`).

## Open Issues

- 없음. CLI backend model 편집 제외는 근거가 있는 범위 결정이고 `research.md` `## Evolution Ideas`와 Reviewer Brief에 적었다.

## Traceability Matrix

| Requirement | Plan Task | Acceptance Scenario | Semantic Invariant |
|-------------|-----------|---------------------|--------------------|
| REQ-RVM-01 | T5 | S1, S11 | INV-RVM-08 |
| REQ-RVM-02 | T2, T5 | S1, S9 | INV-RVM-01 |
| REQ-RVM-03 | T2, T4 | S1, S9 | INV-RVM-01, INV-RVM-08 |
| REQ-RVM-04 | T2, T5 | S3, S9 | INV-RVM-02 |
| REQ-RVM-05 | T2 | S3, S4 | INV-RVM-02, INV-RVM-03 |
| REQ-RVM-06 | T3 | S4 | INV-RVM-03 |
| REQ-RVM-07 | T3, T4 | S2, S5 | INV-RVM-03, INV-RVM-04 |
| REQ-RVM-08 | T3, T5 | S8 | INV-RVM-01 |
| REQ-RVM-09 | T2, T5 | S1, S9, S11 | INV-RVM-08 |
| REQ-RVM-10 | T6 | S5 | INV-RVM-05 |
| REQ-RVM-11 | T6 | S5 | INV-RVM-05 |
| REQ-RVM-12 | T4, T7 | S1, S11, S12 | INV-RVM-07 |
| REQ-RVM-13 | T7 | S10 | INV-RVM-01 |
| REQ-RVM-14 | T8 | S6 | INV-RVM-05 |
| REQ-RVM-15 | T4, T8 | S7 | INV-RVM-06 |
| REQ-RVM-16 | T9 | S13 | - |
| REQ-RVM-17 | T10 | S13 | - |
| REQ-RVM-18 | T10 | S14 | - |
| REQ-RVM-19 | T1, T11 | S15 | INV-RVM-05 |
