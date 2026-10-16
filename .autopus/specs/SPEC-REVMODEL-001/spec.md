# SPEC-REVMODEL-001: 멀티프로바이더 리뷰 모델 구성 탭

**Status**: approved
**Created**: 2026-10-15
**Revised**: 2026-10-16 (review round 1 J-001–J-021: 참여는 `spec.review_gate.providers`만, 줄 splice writer, `runProviderTransportSmoke` 재사용, baseline·working 분리. round 2 J-010·F-001: S4(b) 음성 사례, baseline 1개 경로와 `나가기`. Gate 1 G1-1·G1-2 사용자 확정. 2026-10-16 사용자 결정으로 G1-2 대체: claude·codex CLI model Select, 증명할 목록이 없는 CLI는 OMP 전환(REQ-RVM-21–23, `evidence/probe-cli-model.md`). `evidence/review-rounds.md`)
**Domain**: REVMODEL
**Module**: autopus-adk
**Release**: v0.50.126(SPEC-MODELUX-003, SPEC-REA-001, SPEC-TEAMAUTO-001, redline capability와 함께). 작업 branch는 `feat/revmodel-001`이고 기준은 `feat/rea-001`(ee811e2e)이다. SPEC-REA-001이 구성 UI의 에이전트 수를 17개로 바꿨기 때문이다.
**Predecessor**: SPEC-MODELUX-002(폼 흐름, 확인 Select, Write Safety Contract), SPEC-MODELUX-003(OMP Autopus Tab의 증명된 model 목록과 thinking 규칙). 다르면 이 SPEC이 우선한다.
**Source PRD**: `prd.md`. 근거, probe, 결정은 `research.md`와 `plan.md` `## Risk-First Integration Probe`, review 반영은 `evidence/review-rounds.md`에 있다.

## 목적

`auto spec review`가 어느 provider를 쓰고, provider마다 어느 model·thinking으로 돌며, 누가 judge인지는 지금 `autopus.yaml`을 손으로 고쳐야만 바뀐다. init wizard에는 Review Gate 예/아니오 단계만 있다(`internal/cli/tui/wizard_steps.go` `buildReviewGateStep`). 2026-10-15 실측에서 OMP catalog는 `google-antigravity/gemini-3.1-pro`를 보였지만 한 번 호출하면 HTTP 404였다. 그날의 SPEC 리뷰는 gemini 없이 돌았고, quorum이 차서 receipt의 `degraded_reasons`는 비어 있었다. catalog에 있다는 것만으로는 model이 실제로 답한다고 볼 수 없다.

이 SPEC은 SPEC-MODELUX-002·003의 구성 UI(`auto init`·`auto update`·`auto quality`)에 `멀티프로바이더 리뷰` 탭을 더한다. 이 탭이 정하는 것은 다섯 가지다: (1) `auto spec review`에 참여하는 provider(`spec.review_gate.providers`), (2) provider마다 model과 effort·thinking(OMP backend는 OMP catalog, claude CLI는 Claude 탭 model, codex CLI는 run의 Codex catalog, agy는 `agy models` 목록이 증명한 것만, 모든 CLI는 OMP backend 전환도), (3) judge(`spec.review_gate.judge`, 참여 provider 또는 지금의 judge 없음), (4) 적용 전 미리보기, (5) 사용자가 고를 때만 도는 provider당 1회 확인 요청. 쓰는 key는 v0.50.125가 이미 읽는 기존 key뿐이다.

## Outcome Boundary

- Outcome Lock 원문은 `research.md` `## Outcome Lock`이다.
- Mandatory requirements: REQ-RVM-01 ~ 15, 19 ~ 26 (Must). REQ-RVM-16 ~ 18은 Should다.
- Explicit non-goals:
  - `orchestra.commands.*.providers`와 `orchestra.judge`·`orchestra.commands.*.judge` 쓰기. 명령 목록은 load할 때 `MigrateOrchestraConfig` Migration 3이 `orchestra.providers`의 모든 key로 다시 채운다(`pkg/config/migrate.go:75-88`, probe A2). 그래서 끄기는 orchestra 명령과 `--multi`에 효과가 없고, 미리보기가 그렇게 알린다
  - Migration 3 변경, `--multi`·`--providers` 의미 변경
  - 리뷰 정책 key(strategy, timeout, `enabled`, `max_revisions`, `min_providers`, `verdict_threshold`) 편집
  - CLI args 중 model·effort 값 밖의 항목 편집(그 항목은 bytes 그대로 남는다), OMP→CLI 되돌리기, `agy models` 밖의 agy model 원천
  - `orchestra.providers` 항목 추가·삭제, `backend`·`binary`·`tools`·`subprocess` 변경
  - review receipt `degraded_reasons` 규칙과 `auto doctor --provider-smoke` 동작 변경(Evolution Idea)
  - 확인 요청 자동 실행, 결과 저장(receipt·cache), 새 YAML key
- Completion evidence: Must S1–S12, S15–S27 PASS, `go test ./internal/cli/... ./pkg/...` 필터 없이 통과, 신규·변경 소스 파일 300줄 이하, `GOOS=windows GOARCH=amd64 go build ./...`, `evidence/manual-smoke-revmodel.txt`(S15), `auto spec validate --strict`. `research.md` `## Completion Debt`의 CD-1이 남아 있으면 완료가 아니다.

## Trust Model

- SPEC-MODELUX-001·002 Trust Model을 그대로 따른다. `autopus.yaml`은 반쯤 신뢰한다. writer는 loader가 환경 변수를 펼친 값을 쓰기 원천으로 쓰지 않고 raw 줄과 raw node를 쓴다.
- 새 비신뢰 입력은 넷이다. (0) `agy models` 출력은 Agy Catalog Contract의 엄격한 해석만 거친다. (1) OMP catalog 출력은 기존 `readInstalledOMPModelChoices`의 정규화만 거친다. (2) 확인 요청의 응답과 오류는 `classifyProviderSmokeResult`가 판정하고 결과 줄에는 `terminalSafe`를 거친 첫 줄 200 rune까지만 쓴다. (3) provider 이름과 model 문자열은 폼 라벨·결과 줄에 넣기 전에 `terminalSafe`를 거친다.
- 확인 요청 동안 터미널에 닿는 다른 출력도 같은 경계 안에 둔다. OMP review backend가 stderr에 직접 쓰는 두 줄은 provider·model·tools·사유를 `terminalSafe`로 감싼다(REQ-RVM-20). 이름이 `terminalSafe`로 바뀌는 provider가 하나라도 있으면 확인 요청을 시작하지 않는다(REQ-RVM-10).
- 미리보기·after 표의 CLI 값은 raw document 글자이고 바뀐 flag 값 밖의 args는 보이지 않는다. 대상 값이 placeholder면 편집하지 않는다(REQ-RVM-26).
- 확인 요청은 model 호출이라 비용이 든다. OMP backend는 일시 오류에 한 요청을 최대 3회 시도한다(`ompReviewMaxAttempts`). 그래서 기본값은 호출 없음이고 선택지 라벨이 최대 호출 수를 적는다. 새 secret, 승인 단계, 유료 CI는 없다.

## Requirements

### REQ-RVM-01 탭 진입
Priority: Must · EARS: EventDriven

WHEN the operator chooses 구성 편집 in auto init, auto update, or auto quality while `spec.review_gate.enabled` is true, the config holds at least two `orchestra.providers` entries, and `composeCITruthy(composeCIEnv())` is false, THEN THE SYSTEM SHALL list `멀티프로바이더 리뷰` as the last option of the tool tab select and run the Review Tab Contract when it is chosen, and SHALL otherwise not list it.

- 진입점은 `chooseTab`(`internal/cli/quality_compose_form_driver.go:155`)과 `edit`(`:133`)의 분기다. 290줄 파일이라 선택지 조립은 `[NEW] composeReviewTabOption`(`quality_compose_review.go`)이 하고 driver에는 호출 두 줄만 더한다.
- baseline 참여 수는 진입 조건이 아니다(REQ-RVM-05의 경고 줄과 `나가기`).
- CI 조건은 이 탭만의 규칙이다. update는 이미 CI에서 구성 단계를 열지 않고(`quality_compose_driver_gate.go:66-70`), quality는 CI가 아무것도 정하지 않으며(`quality_compose_entry.go:82-84`, REQ-MUX-16), init은 `!yes && composeTTYProbe()`만 본다(`init.go:100`). 그 gate들은 바꾸지 않는다.

### REQ-RVM-02 참여와 두 값
Priority: Must · EARS: EventDriven

WHEN the review tab starts, THEN THE SYSTEM SHALL take the baseline from the run's config and the raw document, start the working value equal to the baseline, and ask one participation select per `orchestra.providers` entry in `sortedProviderKeys` order whose default and 그대로 label come from the working value, where a provider takes part exactly when the working `spec.review_gate.providers` names it.

- baseline은 탭 진입 때의 값이고 변경 줄, `변경 없음`, 쓰기는 늘 baseline과 비교한다. working은 이번 편집 값이고 `다시 편집`은 working을 기본값으로 다시 묻는다. 검증 오류 뒤에는 working을 baseline으로 되돌린다. SPEC-MODELUX-002 planner의 시작 값·편집 값 분리와 같다.

### REQ-RVM-03 참여 쓰기
Priority: Must · EARS: EventDriven

WHEN the operator applies an edit whose participation differs from the baseline, THEN THE SYSTEM SHALL write `spec.review_gate.providers` as the baseline list without the turned-off providers followed by the turned-on providers in `sortedProviderKeys` order, and SHALL write no `orchestra.commands` key.

- 근거: `auto spec review`는 `--multi`가 없으면 이 목록만 쓰고, 이 목록은 실제 loader(`loadHarnessConfigForFlags`)와 update의 persist migration을 지나도 그대로다(probe A2). `auto doctor` readiness도 이 목록을 본다(`doctor_provider_readiness.go:36`).

### REQ-RVM-04 judge
Priority: Must · EARS: EventDriven

WHEN the participation passes the count check, THEN THE SYSTEM SHALL ask the judge select of the Review Tab Contract, offering 그대로 first when the working judge still takes part or is empty (`그대로 — judge 없음`), otherwise printing the judge-off note and offering only the participating providers with the first as default, and SHALL write `spec.review_gate.judge` only when the chosen judge differs from the baseline.

- 빈 judge는 judge 없는 리뷰라 지원 구성이다(`spec_review_loop.go:78` `NoJudge: p.gate.Judge == ""`). 탭은 judge가 있는 구성을 judge 없음으로 바꾸는 선택지를 내지 않는다.

### REQ-RVM-05 편집 검증
Priority: Must · EARS: Unwanted

IF the participation selects leave fewer than two participating providers, or a planned edit names a judge that does not take part, or a model, thinking, or effort that differs from the baseline and is not among the choices of REQ-RVM-06, REQ-RVM-07, or the CLI Model Contract, THEN THE SYSTEM SHALL print the exact error line of the Form Text table, reset the working value to the baseline, write nothing, and ask the after-error select, whose default 다시 고르기 asks the participation selects again and whose 나가기 returns to the menu.

- 수 검사는 참여 Select 바로 뒤에 하고 보이는 provider(`orchestra.providers` 항목)만 센다. 그래서 0개·1개에서도 judge Select가 열리지 않는다. baseline 참여가 이미 둘보다 적어도 탭에 들어갈 수 있다: 첫 참여 Select 앞에 경고 줄을 쓰고, 사용자는 provider를 켜서 둘 이상으로 만들거나 오류 뒤 `나가기`로 나간다. 그래서 고칠 수 없는 반복은 없다. 검증은 `[NEW] planReviewComposition`(순수 함수)이다. baseline과 같은 model(`그대로`)은 catalog와 대조하지 않는다. 두 개 이상 규칙은 init이 provider가 둘보다 적으면 Review Gate를 끄는 규칙(`init.go:154-157`)과 같다.

### REQ-RVM-06 증명된 model 선택지
Priority: Must · EARS: EventDriven

WHEN the tab asks the model of a participating provider whose entry has `backend: omp`, THEN THE SYSTEM SHALL offer 그대로 followed by exactly the attested installed models that report at least one thinking level and whose `orchestra.ModelFamilyForSelector` family equals the provider's `providerModelFamily` (every attested model when that family is empty), ordered by selector, mark 그대로 with `(catalog에 없음)` when the working model is not among them, and, when the catalog cannot be read, print the catalog note once and ask no model select.

- 증명 목록은 `installedOMPAttestedModelChoices`(`quality_omp_attested_choices.go:17-37`)를 탭 진입당 1회 부른 결과다. strict catalog는 thinking이 없는 model을 빼고(`quality_omp_custom.go:159`), 명시된 빈 thinking 배열은 catalog 전체를 `catalog_invalid`로 만든다. 그 경우는 읽지 못한 catalog와 같다.
- family는 런타임이 쓰는 두 함수로 맞춘다: `providerModelFamily`(`orchestra_brainstorm_judge.go:67-79`)와 `ModelFamilyForSelector`(`pkg/orchestra/model_family.go:31-43`).

### REQ-RVM-07 thinking과 model 쓰기
Priority: Must · EARS: EventDriven

WHEN the operator picks a model other than 그대로, THEN THE SYSTEM SHALL ask its thinking by the OMP Autopus tab rule, offering the levels in the order the normalized catalog reports them, defaulting to the working thinking when reported and otherwise to the first level, and skipping the select when only one level is reported, and SHALL write `orchestra.providers.<p>.model` as `<provider>/<model>:<thinking>`.

- 규칙은 `quality_compose_omp_autopus.go:159-171`과 같다. catalog 정규화는 thinking을 `sort.Strings`로 정렬한다(`pkg/adapter/omp/omp_model_catalog_normalize.go:188-200`). 그래서 `[low, medium, high]`는 `high`, `low`, `medium` 순서로 보인다.
- 근거: OMP backend 런타임은 이 값을 그대로 selector로 쓴다(`omp_review_backend.go:40-44`). codex 항목이 `backend: omp`면 `CodexOrchestraProfile`은 적용되지 않는다(probe A1). 이 key는 그 provider의 모든 실행(`auto spec review`, `--multi`, orchestra 명령)이 쓴다.

### REQ-RVM-08 CLI backend provider
Priority: Must · EARS: StateDriven

WHERE a participating provider entry has no `backend: omp`, THEN THE SYSTEM SHALL ask its model by the CLI Model Contract row that matches its binary, offer `OMP backend로 전환` as an extra choice whenever `omp` is installed and its catalog is readable, and leave every key and every args element of the entry that the chosen row does not name byte-identical.

- 2026-10-16 사용자 결정으로 G1-2(OMP backend만)를 대체한다. 새 `auto init`은 claude·codex·gemini를 모두 CLI backend로 만든다. 근거는 `evidence/probe-cli-model.md`: argv는 `Config.Args` 그대로이고 CLI는 `Model`을 쓰지 않는다.

### REQ-RVM-09 미리보기
Priority: Must · EARS: EventDriven

WHEN the tab's selects are answered and pass validation, THEN THE SYSTEM SHALL print, against the baseline, the after table, the change lines in the order of the Review Tab Contract, and its fixed notes, or the existing 변경 없음 line when the working value equals the baseline.

### REQ-RVM-10 확인 요청 (opt-in)
Priority: Must · EARS: EventDriven

WHEN the preview or the 변경 없음 line has been printed, THEN THE SYSTEM SHALL ask the live check select whose default is 확인 호출 없이 계속, and only when the operator picks 확인 호출 SHALL refuse with the name line if any participating provider name differs from its `terminalSafe` form and otherwise run the Live Check Contract on the candidate and print one result line per result.

### REQ-RVM-11 확인 요청 실패
Priority: Must · EARS: Unwanted

IF any live check result is not 통과, THEN THE SYSTEM SHALL print the live-check failure note and, for a changed edit, keep the confirmation default on 취소 and write nothing unless the operator explicitly picks 적용.

- 바뀐 것이 없으면 확인 Select 없이 메뉴로 돌아간다. 그래서 편집 없이도 지금 구성을 확인할 수 있다(S5(2), S15).

### REQ-RVM-12 쓰기 안전
Priority: Must · EARS: EventDriven

WHEN the operator applies a review tab edit in auto update or auto quality, THEN THE SYSTEM SHALL write it by the Line Splice Writer Contract in one atomic replace, refuse with the write-refused line and leave the file and the run's config unchanged when a target is unsafe, keep every byte outside the replaced spans identical, and then set the run's `cfg.Orchestra` and `cfg.Spec.ReviewGate` from the re-read file.

- 다시 읽기와 cfg 동기화는 OMP 탭의 `ompApply`(`quality_compose_omp.go:263`)와 같다. 그래서 같은 run의 뒤 저장이 이전 구성을 되쓰지 않는다.

### REQ-RVM-13 init 메모리 경로
Priority: Must · EARS: EventDriven

WHEN the operator applies a review tab edit during interactive auto init, THEN THE SYSTEM SHALL apply it to init's in-memory config, which init's save writes, and SHALL apply the wizard's Review Gate answer and the detected provider names before the compose step so that no later init step overwrites `spec.review_gate.providers` or `spec.review_gate.judge`.

- 지금 init은 compose 단계(`init.go:122`) 뒤에 `cfg.Spec.ReviewGate.Providers = providerNames`를 한다(`:126-129`). 이 대입을 compose 단계 앞으로 옮긴다. 그 뒤의 `MigrateOrchestraConfig`(`init.go:161`)는 review gate key를 건드리지 않는다(probe A2). 비대화형 init(`init.go:150-157`)은 그대로다.

### REQ-RVM-14 비대화형 실행
Priority: Must · EARS: Unwanted

IF an entry point runs without its interactive compose gate (auto update with `--yes`, `--check`, a non-terminal stdin, or a truthy CI; auto quality or auto init without a terminal; auto init with `--yes`) or `composeCITruthy(composeCIEnv())` is true, THEN THE SYSTEM SHALL list no review tab, ask no review select, read no OMP catalog for the tab, make no live check request, and write nothing to `autopus.yaml` through this tab.

- 폼은 지금도 `composeFormsTerminal`(`quality_compose_form.go:50-54`)이 참일 때만 열린다. 이 SPEC은 그 gate 밖에 새 입력 경로를 만들지 않고, CI만 REQ-RVM-01이 탭 단위로 더 막는다.

### REQ-RVM-15 옛 바이너리 호환
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL write only `orchestra.providers.<p>.model`, `.args`, `.model_policy`, `.backend`, `spec.review_gate.providers`, and `spec.review_gate.judge`, which v0.50.125 loads, add no key outside the v0.50.125 schema, and write a `model` value only together with `backend: omp` and only when it passes `validOMPModelSelector`.

- probe A3: v0.50.125(dd6d6ea2)로 만든 `auto`가 이 모양의 파일을 읽었고(`config get usage_profile` exit 0), 대조군인 모르는 key와 잘못된 selector는 거부했다. 3차: codex `pinned` + `-m` args 파일과 gemini `backend: omp` + `model` 전환 파일도 읽었다(exit 0).

### REQ-RVM-16 메뉴 요약
Priority: Should · EARS: StateDriven

WHERE the review tab is listed, THEN THE SYSTEM SHALL print the review summary line under the composition table each time the menu is drawn.

- ledger의 `brownfield_impact`가 assumed라 Should다. 표 칸은 늘리지 않고 줄 하나만 더한다.

### REQ-RVM-17 locale
Priority: Should · EARS: Ubiquitous

THE SYSTEM SHALL render every new title, option, note, and line of the Form Text table in Korean when tui.DetectSystemLang returns ko and in English otherwise, while written values stay locale-independent.

### REQ-RVM-18 문서
Priority: Should · EARS: Ubiquitous

THE SYSTEM SHALL document the review tab, the three keys it writes, the live check with its maximum call count, the CLI-backend limit, and that orchestra commands and `--multi` keep every provider, in init, update, and quality help, README, and the v0.50.126 CHANGELOG, without removing sentences existing tests check.

### REQ-RVM-19 test seam과 수동 smoke
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL run every new path through the existing seams (composeRunForm, composeFormsProbe, composeTTYProbe, composeCIEnv, ompCLIFakeRunner, providerSmokeBackendFactory, and the test-only `[NEW] agyModelsRunner`) and SHALL record the real-terminal smoke S15 in `evidence/manual-smoke-revmodel.txt`.

### REQ-RVM-20 OMP review backend 진단 줄
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL pass every config- or process-derived value through `terminalSafe` in the two stderr lines the OMP review backend writes (`omp_review_backend.go:52`, `:72-73`) and in the lines `reportCodexRuntimeFallback` writes for the model and effort, leaving their text unchanged for strings without control characters.

- 이 두 줄은 `auto spec review`와 `auto doctor --provider-smoke`에서도 같다. 제어 문자가 없는 이름에서는 출력이 지금과 같다.

### REQ-RVM-21 claude CLI model
Priority: Must · EARS: EventDriven

WHEN the tab asks the model of a participating CLI provider whose binary is `claude`, THEN THE SYSTEM SHALL offer 그대로 and `ClaudeModelForTier` of fable, opus, sonnet, and haiku, ask after a model choice the effort over `low`–`max` for every model but haiku (default the working `--effort` when valid, else `max`), and write by `editCLIArgs` only the values after `--model` and `--effort`, dropping the `--effort` pair for haiku.

### REQ-RVM-22 codex CLI model
Priority: Must · EARS: EventDriven

WHEN the tab asks the model of a participating CLI provider whose binary is `codex`, THEN THE SYSTEM SHALL offer 그대로 and, ordered by name, the run's Codex catalog snapshot models that report at least one of `low`–`max`, ask the effort over that model's levels among them (default the working effort when reported, else the highest), and on a change write `model_policy: pinned` and, by `editCLIArgs`, only the value after `-m` or `--model` and the `model_reasoning_effort=` value of its `-c` pair.

- `ApplyCodexProviderProfile`은 `--json`을 하나로 줄이는 등 args를 정규화하므로 쓰지 않는다(round 4 F-004). `pinned`면 load migration과 런타임이 덮지 않는다(probe A1 3차). 교집합이 빈 model은 보이지 않는다(F-005).

### REQ-RVM-23 agy model과 OMP 전환
Priority: Must · EARS: EventDriven

WHEN the tab asks the model of a participating CLI provider whose binary is neither `claude` nor `codex`, THEN THE SYSTEM SHALL offer 그대로, the agy models of REQ-RVM-24 when the binary is `agy` and its catalog is readable, and `OMP backend로 전환` when `omp` is installed and its catalog is readable, write an agy choice by `editCLIArgs` as the value after `--model` (inserting `--model <id>` at the start of args when absent), write a switch as `backend: omp` and an attested `model`, and print the no-model reason line when neither list is available.

- 사용자 결정 G1-3(2026-10-16). 전환은 `binary`·`args`·`prompt_via_args`·`subprocess`를 남긴다(probe: 조립 `backend="omp"`, read-only 정책 통과).

### REQ-RVM-24 agy model 목록
Priority: Must · EARS: EventDriven

WHEN the tab needs the models of an `agy` provider, THEN THE SYSTEM SHALL run `<binary> models` once per tab entry by the Agy Catalog Contract, treat any timeout, non-zero exit, oversized output, malformed line, or empty list as unreadable, and offer only listed ids whose family equals the provider's `providerModelFamily`.

### REQ-RVM-25 CLI baseline
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL decide the current model and effort of every CLI provider from a detached copy of the run's config after `config.MigrateOrchestraConfig`, so that `auto quality` (which reads through `LoadPreview` without that migration), `auto update`, and `auto init` decide on the args the runtime runs, and SHALL compute and apply the edit against the raw `orchestra.providers.<p>.args` node by the Line Splice Writer Contract, inserting the flag pair when absent, inserting the migrated args with only the target values changed when the raw args are absent or empty, and refusing alias, anchor, merge, block-scalar, and placeholder nodes.

- round 4 F-002, round 5 V-002: 결정은 migration 뒤 값, 쓰기는 raw node다. 그래서 args가 비었거나 옛 기본 argv인 codex도 쓴 `-m`이 load 뒤 남고 raw의 다른 bytes는 그대로다(S24).

### REQ-RVM-26 CLI 값 표시와 placeholder
Priority: Must · EARS: Unwanted

IF a CLI edit is previewed or its target flag value in the raw document is a `${` placeholder or differs from the loaded value, THEN THE SYSTEM SHALL show only the changed flag values taken from the raw document text and never any other args element, and for a placeholder or differing target SHALL print the CLI value refusal line, keep that provider unchanged, and write nothing for it.

- round 5 V-001. 미리보기는 env가 펼쳐진 run cfg에서 args 전체를 보이지 않는다. `terminalSafe`는 가림이 아니다.

## Contracts

### Review Tab Contract

| 단계 | 규칙 |
|---|---|
| 진입 | REQ-RVM-01의 세 조건을 모두 만족할 때 tool tab Select의 마지막 선택지 |
| 값 | baseline(탭 진입: run의 cfg, 그리고 raw document에서 `spec.review_gate.judge` key가 있는지와 그 raw 값)과 working(이번 편집). Select 기본값과 `그대로 — <값>`은 working, 변경 판정과 쓰기는 baseline 기준. 검증 오류 → working = baseline |
| 참여 | baseline 참여가 둘보다 적으면 첫 Select 앞에 경고 줄. provider마다 Select. 켜짐이면 `그대로 — 켜짐`·`끄기`, 꺼짐이면 `그대로 — 꺼짐`·`켜기`. `orchestra.providers` 항목이 없는 목록 이름은 보이지 않고 바뀌지 않는다 |
| 수 검사 | 참여 Select 바로 뒤, 보이는 provider만 센다. 둘보다 적으면 오류 줄 → 오류 뒤 Select(`다시 고르기` 기본 → 참여 Select부터 baseline 기본값으로, `나가기` → 쓰지 않고 메뉴로) |
| model | OMP backend는 REQ-RVM-06, CLI backend는 CLI Model Contract(REQ-RVM-08, 21–23). OMP·Codex catalog 읽기는 탭 진입당 각 1회 |
| thinking | model을 바꿨을 때만(REQ-RVM-07) |
| CLI 안내 | 증명할 목록이 없는 CLI provider에 OMP 전환을 줄 수 없으면 그 pass에 이유 1줄(REQ-RVM-23) |
| judge | working judge가 참여하면 `그대로 — <j>` 뒤 다른 참여 provider. working judge가 비었으면 `그대로 — judge 없음` 뒤 참여 provider. working judge가 꺼졌으면 안내 1줄 뒤 참여 provider만, 첫 provider가 기본 |
| 검증 | `planReviewComposition`(REQ-RVM-05). 오류 뒤 Select는 수 검사 행과 같다 |
| 미리보기 | CLI args는 바뀐 flag 값만 `orchestra.providers.<p>.args <flag> <old> → <new>`로 보이고 다른 항목은 보이지 않으며 값은 raw 글자다(REQ-RVM-26). after 표, 바뀌는 설정 줄(순서: `spec.review_gate.providers` → provider 순으로 그 provider의 `backend` → `model` → `model_policy` → `args` → `spec.review_gate.judge`, 모양 `<key> <before> → <after>`, 목록은 `[a, b]`, 빈 judge는 `judge 없음`), 고정 안내(provider를 끈 편집이면 명령 안내, model을 바꾼 편집이면 model 안내). 같으면 `변경 없음` |
| 확인 요청 | 미리보기나 `변경 없음` 뒤 `확인 호출 없이 계속`(기본) · `확인 호출 — provider <n>개에 요청 1회씩 (재시도 포함 최대 <3n>회 model 호출)`. `<n>`은 후보의 참여 provider 수 |
| 확인 | 바뀐 편집만: 기존 `review`(`quality_compose_form_driver.go:213`)와 같은 `취소`(기본) · `적용` · `다시 편집`. 바뀐 것이 없으면 확인 없이 메뉴로 |
| 적용 | update·quality는 Line Splice Writer Contract, init은 메모리 경로. 성공하면 기존 `적용했습니다` 줄 |

after 표(가로는 `terminalSafe` 뒤 rune 폭으로 맞추고 바뀐 칸 끝에 `*`, CLI backend의 model 칸은 raw 글자의 `<binary> <model> · <effort>`, model flag 값이 없으면 `<binary>`):

```text
provider  리뷰   model                                    judge
claude    켜짐   anthropic/claude-fable-5-1:max           judge
codex     켜짐   openai-codex/gpt-6-luna:high *
gemini    꺼짐 * google-antigravity/gemini-3.1-pro:high
```

### CLI Model Contract

| binary | 선택지 | 쓰기(`editCLIArgs`, 다른 args 항목은 bytes 그대로) | 근거 |
|---|---|---|---|
| `claude` | `그대로`, `claude-fable-5-1`·`claude-opus-5-5`·`claude-sonnet-5-5`·`claude-haiku-4-5`, effort low–max(haiku 없음) | `--model`·`--effort` 값, 없으면 print flag 뒤에 쌍 삽입, haiku는 `--effort` 쌍 삭제 | `claude_provider.go:36-47`, `quality_tier.go:8-11, 104-115` |
| `codex` | `그대로`, 단계 교집합이 비지 않은 Codex catalog model, 그 effort | `model_policy: pinned`, `-m`/`--model` 값, `-c model_reasoning_effort=` 값 | `codex_profile.go:66-75`, probe A1 3차 |
| `agy` | `그대로`, Agy Catalog Contract의 같은 family id(기본 이름 → effort 접미사) | `--model` 값, 없으면 args 맨 앞에 `--model <id>` | 사용자 관측 agy 1.4.0, G1-3 |
| 모든 CLI | OMP가 설치되고 catalog를 읽으면 `OMP backend로 전환` 추가 | `backend: omp` + `model` | REQ-RVM-23 |

- `editCLIArgs`(`[NEW]`)는 이름 붙은 flag 바로 뒤 값만 바꾸고, 없으면 정해진 자리에 쌍을 넣으며, 그 밖의 항목·순서·중복(`--json` 두 번 등)을 그대로 둔다. 쓰기 전 validation(REQ-RVM-05)과 미리보기 줄(REQ-RVM-09)은 이 행들의 선택지와 쓰기 값으로 계산한다(F-003).

### Agy Catalog Contract

| 항목 | 규칙 |
|---|---|
| 실행 | argv `<binary> models`, timeout 15초, stdout 64 KiB 상한, stderr는 버린다. test는 `agyModelsRunner` seam |
| 해석 | 첫 줄이 정확히 `Fetching available models...`면 건너뛴다. 나머지 비지 않은 줄은 모두 `^[a-z0-9][a-z0-9.-]*\t[^\t\x00-\x1f]+$`여야 한다. 하나라도 어긋나거나 model이 0개면 목록 전체를 읽지 못한 것으로 본다 |
| family | id 앞부분: `gemini-` → google, `claude-` → anthropic, `gpt-` → openai, 그 밖은 없음. provider family와 같은 id만 보인다 |
| 묶음 | 끝이 `-low`·`-medium`·`-high`인 id는 그 접미사를 뺀 기본 이름으로 묶는다. model Select는 기본 이름, effort Select는 그 묶음에 실제로 있는 접미사뿐이고 하나면 묻지 않는다. 쓰는 값은 늘 목록에 그대로 있던 id다(만든 id를 쓰지 않는다). 다른 탭의 model → effort 흐름과 같아 Select가 짧다 |
| 표시 | display name은 `terminalSafe` 뒤 effort Select 라벨에 쓴다 |
| 한계 | 목록에 있다고 호출이 되는 것은 아니다(2026-10-15 `gemini-3.1-pro:high` 404). 그래서 확인 요청을 권장한다 |

### Config Write Contract### Config Write Contract

| 선택 | 쓰는 key | 바뀌는 소비처 |
|---|---|---|
| 참여 | `spec.review_gate.providers` = baseline 순서에서 끈 provider를 뺀 것 + 켠 provider(`sortedProviderKeys` 순) | `auto spec review`(`--multi` 없음), `auto doctor` readiness |
| model (`backend: omp`) | `orchestra.providers.<p>.model` = `<provider>/<model>:<thinking>` | 그 provider의 모든 OMP backend 실행: `auto spec review`, `--multi`, orchestra 명령 |
| claude CLI model·effort | `orchestra.providers.<p>.args`의 `--model`·`--effort` 값 | 그 provider의 모든 CLI 실행 |
| codex CLI model·effort | `orchestra.providers.<p>.model_policy: pinned`, `args`의 `-m`·`-c model_reasoning_effort` 값 | 그 provider의 모든 CLI 실행. `CodexOrchestraProfile`은 더 적용되지 않는다 |
| OMP 전환 | `orchestra.providers.<p>.backend: omp`와 `model` | 그 provider가 OMP review backend로 돈다 |
| agy model | `orchestra.providers.<p>.args`의 `--model` 값(없으면 맨 앞에 쌍) | 그 provider의 모든 CLI 실행 |
| judge | `spec.review_gate.judge` = j. baseline에 key가 없으면 review_gate mapping 끝에 더하고, `""`면 그 값만 바꾼다 | `auto spec review` judge |
| 쓰지 않음 | `orchestra.commands.*`(Migration 3 소유), `orchestra.judge`, 리뷰 정책 key, `binary`, `tools`, `subprocess`, `prompt_via_args`, 위 행이 말하지 않는 args 항목, provider 항목 추가·삭제 | — |

- 바뀌지 않는 값은 쓰지 않는다. working이 baseline과 같으면 파일을 열어 쓰지도 않는다.

### Line Splice Writer Contract

| 항목 | 규칙 |
|---|---|
| 원천 | 디스크의 bytes. `saveQualityBlock`처럼 먼저 `pruneRetiredConfig`(`config_retired_keys.go:37`)를 거친다(SPEC-PANERM-001 REQ-10). 대상 위치는 raw document(`parseAutopusConfigDocument`, `platform_omp_config.go:45`, 환경 변수 펼침 없음)의 node `Line`·`Column`으로 찾는다 |
| 바꾸는 span | scalar(`judge`, `model`): 그 key 줄의 값 token(plain, `'…'`, `"…"`)만. 같은 줄의 inline 주석은 남는다. flow 목록: 같은 줄의 `[`부터 짝 `]`까지. block 목록: 항목 줄들을 같은 들여쓰기의 `- <name>` 줄로 바꾼다. 없는 `judge` key: review_gate mapping의 마지막 줄 뒤에 그 mapping 들여쓰기로 `judge: <j>` 한 줄. 없는 `model_policy`·`backend`·`model` key도 같은 방식으로 그 provider mapping 끝에 넣는다. CLI args는 목록 token 전체를 새 목록으로 바꾼다(model·effort 값 밖의 항목과 순서는 그대로). 새 값은 `OMPYAMLScalar`(`pkg/content/agent_transformer_omp.go:25`)로 encode한다 |
| 거부 | 대상 key·값 node나 span 안의 node가 anchor 또는 alias, 부모 mapping에 merge key, 대상 값에 `${`, 부모가 flow mapping, 값이 block scalar(`|`, `>`)이거나 목록 항목이 scalar가 아님, block 목록 span 안의 주석 줄. 문구는 Form Text의 쓰기 거부 줄이고 `<key>`는 첫 거부 경로다 |
| 검사 | `[NEW] verifyReviewLineEdit`: `verifyQualityLineEdit`(`quality_config_verify.go:20`)처럼 대상 경로를 뺀 두 문서가 같은 data(alias는 가리키는 data로)이고, 대상 경로가 계획 값이며, strict load와 `Validate`를 지나 load 결과가 baseline cfg에 계획 값만 바꾼 것과 같아야 한다. 하나라도 어긋나면 쓰지 않는다 |
| 쓰기 | `atomicWriteQualityConfig`(`quality_config.go:258`)와 같은 atomic replace. 파일 mode는 그대로다 |
| bytes | retired key가 없는 파일에서 span 밖 bytes는 원본과 같다. retired key가 있으면 `pruneRetiredConfig` 결과와 같다 |

### Live Check Contract

| 항목 | 규칙 |
|---|---|
| 언제 | 미리보기나 `변경 없음` 뒤 사용자가 `확인 호출`을 고를 때만(REQ-RVM-10). 비대화형·CI에서는 탭이 없다(REQ-RVM-14) |
| 엔진 | `runProviderTransportSmoke(ctx, candidate, 60초)`(`doctor_provider_smoke.go:87-121`)를 그대로 부른다. candidate는 run의 cfg 사본에 working을 넣은 것이고 파일은 아니다. 조립 오류나 provider 0개는 그 함수가 `review_gate` 실패 결과 하나로 바꾼다 |
| 대상 | 그 함수가 `assembleSpecReviewProviders`로 만든 집합(read-only 정책, 설치 확인). 실제 리뷰는 readiness 필터(`planSpecReviewProviders`, `spec_review_readiness.go:59`)를 더 거치므로 통과는 "그 provider가 답했다"이지 "모든 다음 리뷰에 들어간다"가 아니다 |
| 요청 수 | 집합의 provider마다 `Execute` 1회, 차례로. OMP backend는 일시 오류면 한 요청을 최대 3회 시도한다(`omp_review_backend_session.go:13, 57-63`). 그래서 model 호출은 최대 3n회이고 라벨이 그 수를 적는다 |
| prompt | `providerSmokePrompt()`의 글자 그대로: `Return ONLY this exact token and no other text: AUTOPUS_PROVIDER_SMOKE_OK`. 확인 요청이 더하는 flag는 없다. OMP provider는 `ompReviewSession.run`(`omp_review_backend_session.go:31`), CLI provider는 그 binary·args 그대로다 |
| timeout | 요청마다 60초(`[NEW] composeReviewCheckTimeout`). doctor 기본 30초(`doctor.go:106`)는 바꾸지 않는다 |
| 결과 줄 | `pass` → `확인 호출 <p> (<model>): 통과`, `warn` → `확인 호출 <p> (<model>): 응답에 확인 문자열이 없습니다`, `fail` → `확인 호출 <p> (<model>): 실패 — <detail>`. `<model>`은 candidate의 selector, CLI면 `CLI backend`, `review_gate` 결과면 ` (<model>)` 없이 쓴다. `<detail>`은 `terminalSafe`를 거친 첫 줄 200 rune까지 |
| 실패 뒤 | 안내 1줄. 바뀐 편집이면 확인 기본값 `취소`(REQ-RVM-11). 결과는 저장하지 않는다 |
| 터미널 | 이름 검사(REQ-RVM-10)와 backend 진단 줄(REQ-RVM-20). CLI subprocess가 provider stderr를 터미널에 넘기는지는 T6이 먼저 확인하고, 넘기면 그 출력도 `terminalSafeMultiline`을 거친다 |

### Forward-compat Contract

| 항목 | 규칙 |
|---|---|
| key | REQ-RVM-15의 세 경로뿐이다. 모두 v0.50.125 `pkg/config/schema_orchestra.go`·`schema_spec.go`에 같은 yaml tag로 있다(`git show v0.50.125:...`) |
| 값 | model은 `validOMPModelSelector`(`schema_provider_backend.go:56-75`)를 지나는 값만. 아니면 v0.50.125가 `provider_model_required: <p>`로 파일 전체를 거부한다(A3 대조군) |
| 검증 | S7이 이 branch가 쓴 파일을 v0.50.125 바이너리로 읽는다 |

### Form Text

| 자리 | ko | en |
|---|---|---|
| 탭 | `멀티프로바이더 리뷰` | `Multi-provider review` |
| 참여 | `<p> 리뷰 참여` · `그대로 — 켜짐` · `끄기` · `그대로 — 꺼짐` · `켜기` | `<p> takes part in review` · `keep — on` · `turn off` · `keep — off` · `turn on` |
| model | `<p> model` · `그대로 — <selector>` · ` (catalog에 없음)` | `<p> model` · `keep — <selector>` · ` (not in catalog)` |
| thinking | `<p> thinking (<provider>/<model>)` | 같다 |
| CLI model | `<p> model` · `그대로 — <model> · <effort>` · `<p> effort` | `<p> model` · `keep — <model> · <effort>` · `<p> effort` |
| Codex 안내 | `Codex catalog를 읽지 못해 codex model은 그대로입니다` | `The Codex catalog is unreadable; the codex model stays` |
| 전환 | `그대로 — CLI backend` · `OMP backend로 전환` | `keep — CLI backend` · `switch to the OMP backend` |
| model 없음 이유 | `<p>(<binary>)는 Autopus가 증명할 model 목록이 없어 model을 바꾸지 않습니다. OMP를 설치하면 OMP backend로 전환할 수 있습니다` | `<p> (<binary>) has no model list Autopus can attest, so its model stays; install OMP to switch it to the OMP backend` |
| catalog 안내 | `OMP catalog를 읽지 못해 model은 그대로입니다` | `The OMP catalog is unreadable; models stay unchanged` |
| judge | `judge` · `그대로 — judge 없음` · `judge <j>가 꺼져 judge를 다시 고릅니다` | `judge` · `keep — no judge` · `judge <j> was turned off; choose a new judge` |
| 오류 | `리뷰에는 provider가 두 개 이상 필요합니다 (지금 <n>개)` · `judge <j>는 리뷰에 참여하지 않는 provider라 고를 수 없습니다` · `<p> model <selector>는 설치된 OMP catalog가 증명하지 않습니다` · `<p> thinking <t>는 <selector>가 보고하지 않는 단계입니다` | `The review needs at least two providers (now <n>)` · `judge <j> does not take part in the review and cannot be chosen` · `The installed OMP catalog does not attest <p> model <selector>` · `<selector> reports no thinking level <t> for <p>` |
| 오류 뒤 | `다시 고르기` · `나가기` · `지금 리뷰 provider가 <n>개뿐입니다. 두 개 이상 켜야 적용할 수 있습니다` | `choose again` · `leave` · `Only <n> review providers now; turn on at least two to apply` |
| 명령 안내 | `꺼진 provider도 orchestra 명령(brainstorm·plan·review·secure)과 auto spec review --multi에서는 계속 돕니다` | `Turned-off providers still run in orchestra commands (brainstorm, plan, review, secure) and in auto spec review --multi` |
| model 안내 | `바뀐 model은 auto spec review와 orchestra 명령 모두에 쓰입니다` | `A changed model applies to auto spec review and to orchestra commands` |
| 확인 요청 | `확인 호출` · `확인 호출 없이 계속` · `확인 호출 — provider <n>개에 요청 1회씩 (재시도 포함 최대 <3n>회 model 호출)` · 결과 줄 · `확인 호출이 실패한 provider가 있습니다. 적용하면 그 provider는 리뷰에서 실패합니다.` · `확인 호출을 할 수 없습니다: provider 이름 <terminalSafe(p)>에 터미널 제어 문자가 있습니다` | `Live check` · `continue without a live check` · `live check — one request for each of <n> providers (at most <3n> model calls with retries)` · `live check <p> (<model>): passed` / `: the reply lacks the check token` / `: failed — <detail>` · `A live check failed; if you apply, that provider keeps failing in reviews.` · `Cannot run the live check: provider name <terminalSafe(p)> holds terminal control characters` |
| agy 안내 | `<p>(agy) model 목록을 읽지 못해 agy model은 그대로입니다` | `Could not read the agy model list for <p>; its agy model stays` |
| CLI 오류 | `<p> model <m>는 이 provider의 선택지에 없습니다` · `<p> effort <e>는 <m>이 받지 않는 단계입니다` | `<m> is not a choice for <p>` · `<m> does not take effort <e> for <p>` |
| CLI 값 거부 | `<p>의 <flag> 값이 환경 변수라 이 탭에서 바꾸지 않습니다: <raw>` | `<flag> of <p> comes from an environment variable; this tab leaves it: <raw>` |
| 쓰기 거부 | `autopus.yaml의 리뷰 설정을 안전하게 고칠 수 없어 아무것도 쓰지 않았습니다: <key>` | `The review settings in autopus.yaml cannot be edited safely; nothing was written: <key>` |
| 메뉴 요약 | `멀티프로바이더 리뷰: claude · codex · gemini(꺼짐) · judge claude` (빈 judge는 `judge 없음`) | `Multi-provider review: claude · codex · gemini (off) · judge claude` (`no judge`) |

## Existing Test Changes

1. tool tab Select가 선택지 하나라 건너뛰던 폼 test 중 fixture가 REQ-RVM-01 조건을 만족하는 것은 이제 그 Select가 열린다. 입력에 그 Select의 기본값을 고르는 Enter(`\n`) 하나를 더한다. 기존 줄 단언은 바꾸지 않는다.
2. tool tab Select의 선택지 목록 전체를 단언하는 test는 마지막에 `멀티프로바이더 리뷰`를 더한다.
3. init test 중 compose 단계 뒤 `spec.review_gate.providers`를 단언하는 것은 값이 같아 바뀌지 않는다(대입 순서만 바뀐다).
4. OMP review backend stderr 줄을 bytes로 단언하는 test가 있으면 제어 문자가 없는 입력이라 그대로다. 그 밖의 test는 바꾸지 않는다. T1이 파일마다 분류한다.

## 생성 파일 상세

- `internal/cli`: `[NEW] quality_compose_review.go`(+test, 탭 흐름·`composeReviewTabOption`), `[NEW] quality_compose_review_plan.go`(+test, baseline·working, `planReviewComposition`, 변경 줄, after 표), `[NEW] quality_compose_review_choices.go`(+test, family로 거른 증명 목록, thinking 규칙), `[NEW] quality_compose_review_write.go`(+test, `saveReviewComposition`·span·거부·init 메모리 적용), `[NEW] quality_compose_review_verify.go`(+test, `verifyReviewLineEdit`), `[NEW] quality_compose_review_check.go`(+test, 확인 요청), `[NEW] quality_compose_review_cli.go`(+test, CLI Model Contract, `editCLIArgs`, OMP 전환), `[NEW] quality_compose_review_agy.go`(+test, Agy Catalog Contract), `quality_compose_form_driver.go`(호출 두 줄), `quality_compose_form_text.go`(문구), `init.go`(대입 순서), `omp_review_backend.go`(두 줄에 `terminalSafe`). 메뉴 요약은 `f.draw()` 뒤 한 줄이고 driver가 300줄을 넘으면 `composeReviewTabOption`처럼 새 파일로 옮긴다.
- 문서: README, CHANGELOG(v0.50.126), help 상수(Should). `content/`·`templates/`를 고치면 `make generate-templates`로 생성 표면을 맞춘다.
- 증거: `evidence/review-rounds.md`, `[NEW] evidence/manual-smoke-revmodel.txt`. 신규·변경 소스는 300줄 이하다.

## Related SPECs

- SPEC-MODELUX-002(v0.50.125): 폼 흐름, tool tab, 확인 Select, Write Safety Contract, 줄 writer 검사 방식을 쓴다. 바뀌는 점은 tool tab 선택지가 하나 늘어난다는 것뿐이다.
- SPEC-MODELUX-003(v0.50.126): `installedOMPAttestedModelChoices`와 thinking 규칙을 그대로 쓴다. OMP 탭(에이전트 route)과 이 탭(리뷰 provider model)은 서로 다른 key를 쓴다.
- SPEC-REVIEWRO-001: OMP review backend와 `auto doctor --provider-smoke` 경로를 확인 요청이 그대로 쓴다. 진단 줄 두 개만 `terminalSafe`를 더한다.
- SPEC-PANERM-001: retired key prune 규칙(REQ-10)을 writer가 따른다.
- SPEC-ADK-REVIEW-INTEGRITY-001: quorum과 `degraded_reasons` 규칙은 바꾸지 않는다(Evolution Idea).
- SPEC-REA-001: 기준 branch. 이 SPEC은 에이전트 표를 바꾸지 않는다.
- Sibling: 없다(`research.md` `## Sibling SPEC Decision`).

## Open Issues

Gate 1 범위 결정 둘은 2026-10-16 사용자가 확정했다(근거는 `research.md` D2·D3).

- G1-1 | resolved 2026-10-16 | 사용자 결정: 참여 끄기는 `auto spec review`(`spec.review_gate.providers`)에만 적용하고, 미리보기의 명령 안내 줄(orchestra 명령과 `--multi`는 그대로)을 유지한다. 명령 목록은 Migration 3이 load마다 다시 채우므로 그것을 바꾸는 일은 scope expansion으로 남는다
- G1-2 | superseded 2026-10-16 | 처음 결정(OMP backend만)은 새 `auto init`의 세 CLI provider에서 쓸모가 없었다. 새 사용자 결정: claude CLI는 `--model`, codex CLI는 `-m`(`pinned`)으로 고르고, model 목록을 증명할 수 없는 CLI는 OMP backend 전환 또는 이유 줄이다(REQ-RVM-21–23)
- G1-3 | resolved 2026-10-16 | 사용자 결정: `agy models` 해석기를 더해 agy provider도 model Select를 갖는다(REQ-RVM-23, 24, Agy Catalog Contract). `OMP backend로 전환`은 다른 선택지로 남는다

## Traceability Matrix

| Requirement | Plan Task | Acceptance Scenario | Semantic Invariant |
|-------------|-----------|---------------------|--------------------|
| REQ-RVM-01 | T5 | S1, S6, S10, S18 | INV-RVM-08 |
| REQ-RVM-02 | T2, T5 | S1, S3, S17 | INV-RVM-01, INV-RVM-09 |
| REQ-RVM-03 | T2, T4 | S1, S11 | INV-RVM-01 |
| REQ-RVM-04 | T2, T5 | S3, S9 | INV-RVM-02 |
| REQ-RVM-05 | T2 | S3, S4, S18, S26 | INV-RVM-02, INV-RVM-03, INV-RVM-09 |
| REQ-RVM-06 | T3 | S4 | INV-RVM-03 |
| REQ-RVM-07 | T3, T4 | S2, S5 | INV-RVM-03, INV-RVM-04 |
| REQ-RVM-08 | T3, T5, T12, T13, T14 | S8, S19, S20, S21 | INV-RVM-11 |
| REQ-RVM-09 | T2, T5 | S1, S2, S11, S17 | INV-RVM-08, INV-RVM-09 |
| REQ-RVM-10 | T6 | S5, S16 | INV-RVM-05 |
| REQ-RVM-11 | T6 | S5 | INV-RVM-05 |
| REQ-RVM-12 | T4, T7 | S11, S12 | INV-RVM-07 |
| REQ-RVM-13 | T7 | S10 | INV-RVM-01 |
| REQ-RVM-14 | T8 | S6 | INV-RVM-05 |
| REQ-RVM-15 | T4, T8 | S7, S22 | INV-RVM-06 |
| REQ-RVM-16 | T9 | S13 | - |
| REQ-RVM-17 | T10 | S13 | - |
| REQ-RVM-18 | T10 | S14 | - |
| REQ-RVM-19 | T1, T11 | S15 | INV-RVM-05 |
| REQ-RVM-20 | T6 | S16 | INV-RVM-10 |
| REQ-RVM-21 | T12 | S19, S24 | INV-RVM-11 |
| REQ-RVM-22 | T13 | S20, S24, S25 | INV-RVM-11 |
| REQ-RVM-23 | T14 | S8, S21, S23 | INV-RVM-12 |
| REQ-RVM-24 | T14 | S23 | INV-RVM-12 |
| REQ-RVM-25 | T12 | S24 | INV-RVM-11 |
| REQ-RVM-26 | T12 | S26, S27 | INV-RVM-07 |
