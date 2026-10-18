# SPEC-TUIUX-001 리서치

branch `feat/tuiux-001`, 기준 8ec13935(`feat/revmodel-001` 위, SPEC-REA-001·SPEC-REVMODEL-001 UI 포함)에서 2026-10-16에 읽었다. huh 사실은 module cache의 huh v1.0.0 소스와 probe(`evidence/probe-nav-a1.md`)다. revision 2·3은 review round 1·2를 반영했다(`evidence/review-rounds.md`). 확인하지 못한 것은 `미확인`이다.

## 기존 코드 분석
- wizard: `internal/cli/tui/wizard_steps.go`(275줄). `RunInitWizard`(L43-72)는 `buildStepList`(L89-138)의 단계마다 따로 `huh.Form`을 `form.Run()`하고 오류가 나면 `Cancelled: true`를 돌려준다. 제목은 `[n/N] …`이고 값은 `Value(&r.…)`로 결과에 묶인다. `Accessible` 옵션은 비TTY·`--yes`의 "폼 없이 기본값"이다(L49-52). REA 단계는 `reaStepTextFor(lang)`로 ko·en이다. wizard에는 runner seam이 없다.
- 구성 단계: 모든 폼이 `composeSelect`·`composeConfirm` → `runComposeForm` → `composeRunForm(cmd, form)`을 지난다(`quality_compose_form.go:68-111`, 호출 27곳). 흐름은 중첩 함수다: `runForms` → `menu` → `edit` → `chooseTab` → `compose`(방식 → `groups`/`agents` → `PlanEdit` → `review`) 또는 `omp`·`reviewTab`(`quality_compose_form_driver.go`, 296줄). `적용할까요?`는 취소·적용·계속 편집, 기본값 취소다(L219-237).
- OMP 탭 `ompFailed`(`quality_compose_omp.go:130-138`)는 `errComposeInterrupted`만 넘기고 나머지는 경고로 바꾼다. native `ompPins`(L93)와 autopus 탭(`quality_compose_omp_autopus.go:43-93`)의 오류가 이 wrapper를 지난다. native 첫 폼의 기본값은 `돌아가기`("")다. 리뷰 탭의 `f.warn` 두 곳은 후보 config 오류(`quality_compose_review_check.go:27-31`)와 쓰기 오류(`quality_compose_review.go:228-241`)이고 폼 오류는 그대로 올라간다.
- init: `init.go:98-101`이 wizard 전에 기존 프로필을 `config.Load(dir)`로 읽는다. `Load`는 `MigratePlatformNames`가 이름을 고치면 autopus.yaml을 다시 쓰고(`pkg/config/loader.go:87-94`) `LoadPreview`는 쓰지 않는다(`:21-27`). 예: `gemini` → `antigravity-cli`(`migrate.go:185`). 그 뒤 `init.go:106-147`이 wizard → 답 적용 → `runInitComposeStepCancellable` → 방법론 → `rea.decideInit` 순이다. `decideInit`은 `cfg.Runtime.ReaMCP`(`yaml:"-"`)만 정하고 REA key는 생성 뒤 `reaRun.finish`의 `reaKeyAfterRun`이 정한다(`rea_run.go:112-127`, `rea_decision.go:153`). 이어서 `MigrateOrchestraConfig`, `applyStatusLineMode`(질문만), `warnParentRuleConflicts`(충돌이면 `config.Save`, `prompts.go:99-140`), Step 2 `DESIGN.md`, Step 3 생성, Step 5 `SummaryTable`(L224-232)이다. 리뷰 게이트를 꺼도 `DefaultFullConfig`의 provider 셋(claude·codex·gemini, `defaults.go:75-90`)은 남는다(`init.go:131-134`). 취소 줄은 `"  init cancelled."`(L120)다. `cloneHarnessConfig`가 있다(`update_preview.go:20`).
- 생성의 시간 의존: adapter manifest `generated_at`은 `time.Now()`(`pkg/adapter/manifest.go:55`, `:97`), backup 디렉터리 이름은 timestamp(`:179`)다. project 이름은 디렉터리 이름이다(`init.go:52`).
- 폼 gate: update는 `composeGate`(`--yes`·`--check`·TTY probe false·CI truthy면 묻지 않음, `quality_compose_driver_gate.go:64-70`), quality는 `qualityComposeInteractive`(CI를 보지 않음, `quality_compose_entry.go:83-86`), init wizard는 `!yes && composeTTYProbe()`(`init.go:106`)다.
- huh v1.0.0: 기본 KeyMap의 `Quit`은 ctrl+c뿐이다(`keymap.go:109`). Select에서 Esc는 filter에만 쓰인다: 입력 중이면 `SetFilter`, 끝난 뒤 글자가 남았으면 `ClearFilter`(`keymap.go:148-149`, `field_select.go:402-414`). `NewForm`은 `TERM=dumb`이면 `WithAccessible(true)`를 부르고(`form.go:124-127`) accessible 출력 기본은 os.Stdout이다(`:331`, `:672`). accessible 실행은 field 오류를 버리고(`:703-721`) Select는 `PromptInt`로 숫자만 받으며 EOF면 기본값이다. `WithProgramOptions`는 기존 option을 덮는다(`:353-356`). bubbletea v1.3.6은 높이를 넘는 위쪽 줄을 버린다(`standard_renderer.go:182-187`).
- 비test 코드에 `huh.NewInput`·`NewText`·`NewMultiSelect`가 없다. `WithAccessible(` 호출도 없지만 위의 `TERM=dumb` 경로는 production이다(review F-006로 고쳤다).

### 기존 test 분류
T1이 파일·줄 단위로 채운다.

T1 결과(2026-10-16, b08ea644 기준. 이 구현의 branch 기준은 8ec13935에 SPEC-REVMODEL-001 구현을 합친 b08ea644이고 base golden도 여기서 TUIUX 소스 변경 전에 만들었다):
- 분류 1(시그니처, 다섯 줄과 import): `quality_compose_form_helpers_test.go:44`(`useComposeFormRunner`의 stub), `quality_compose_form_failure_test.go:29`(`terminal gone` stub), `quality_compose_line_golden_test.go:25`(`countComposeFormRuns`), `quality_compose_review_gate_test.go:28`(`countReviewForms`), `quality_compose_form_test.go:94`(기본 runner 직접 호출). 다섯 파일 모두 `internal/cli/tui` import가 없어 import 한 줄이 더해진다. 이 밖에 `composeRunForm`에 값을 넣거나 부르는 test는 없다(grep `composeRunForm` 2026-10-16).
- 분류 2(init 확인 화면): wizard seam과 forms probe true로 init을 끝까지 도는 test는 `init_compose_form_test.go`(2개), `init_compose_review_test.go`(`TestInitComposeReview_S10TabEditSurvivesInit` 3개 subtest), `init_compose_omp_test.go`(`runComposeInitOMP`를 쓰는 5개), `quality_compose_form_effort_test.go:217`(`TestComposeEffortForms_InitSavesAndGeneratesTheEfforts`)다. `init_rea_test.go`와 `init_compose_test.go`는 `strings.Reader` 입력이라 기본 forms probe가 줄 driver를 고르고 확인 화면이 없다(분류 밖, 변경 없음). 파일별 (a)/(b)/(c)는 T9이 돌린 결과로 아래에 적는다.
- 분류 2 결과(T9, 2026-10-16): `init_compose_form_test.go`의 두 test는 그대로 통과한다(중단 test는 확인 화면 앞인 2번 폼에서 끝나고, Codex 편집 test는 유지 뒤 EOF가 적용을 고른다: (a)). `init_compose_review_test.go`의 세 subtest도 그대로다((a)·(a)·(b)). `quality_compose_form_effort_test.go:217`은 (a)다. `init_compose_omp_test.go`의 네 test는 그대로 통과하고, `TestInitComposeOMP_BlockerRefusesAndInitContinues`는 처음에 단언(`:105` `NotContains(out, "적용할까요?")`)이 확인 화면 제목(당시 `이대로 적용할까요?`)과 겹쳐 실패했다. 분류 밖이라 test를 고치지 않고 lead 결정으로 제목을 `이대로 진행할까요?`로 바꿨다(구현 중 개정 A-1). 이제 이 test도 고치지 않고 통과한다. 분류 2(c)에 해당하는 입력 수정은 0건이다.
- 분류 1 정정: `quality_compose_form_test.go`는 이미 `internal/cli/tui`를 import한다(L23). 그래서 import 줄은 네 파일에만 더해졌다.
- T10 확인(2026-10-16): `git diff b08ea644 -- '*_test.go'`에서 바뀐 기존 test 파일은 분류 1의 다섯 파일과 T5가 소유한 `internal/cli/tui/export_test.go`(test alias 두 줄 `TestWizardStreams`·`TestWizardStepKinds`, 단언 아님)뿐이다. 바뀐 `assert.`·`require.` 줄은 `quality_compose_form_test.go:94`의 호출 인자 한 줄이다.
- acceptance 기대 정정(코드 사실, Oracle Acceptance Notes 규칙): S11 (c)의 `quality.default =`는 프리셋 쓰기 줄로, S14 (c)의 `*` 칸은 `Claude Code (개별 설정: claude-custom)` 열 머리와 칸 값으로, S15 (f)는 리뷰 탭 자신의 CI 규칙이 지우는 리뷰 요약 줄 하나를 빼고 같다로, S21의 frame에 `Lang: "ko"`를 더했다(기대한 footer가 한국어라서).
- `runInitComposeStepCancellable`을 부르는 test는 없다(호출은 `init.go:136` 하나). T9은 이 함수를 `runInitComposeStage`로 바꿔도 test를 고치지 않는다.
- 폼 helper(`useAccessibleComposeForms`·`abortComposeFormAt`·`useComposeFormRunner`)를 쓰는 기존 test 파일은 18개다(새 `tuiux_noninteractive_golden_test.go` 제외).
- base golden은 `evidence/base-goldens/`의 `<흐름>.stdout`·`<흐름>.npd` 일곱 쌍이다: `init-yes`(S15 a), `update-yes`(b), `update-ci`(c), `quality-line-b`(d), `quality-no-tty`(e), `init-pipe`(g), `init-in`(S6 IN, NPD만 비교). `internal/cli/tuiux_noninteractive_golden_test.go`가 읽고, `TUIUX_WRITE_GOLDENS=1`이면 흐름마다 두 번 돌려 같을 때만 다시 쓴다. 2026-10-17 재기록: TUIUX 밖 변경이 폼이 열리지 않는 자리의 생성 파일을 바꿔 이 golden과 S6이 깨졌다. 출처는 SPEC-TEAMAUTO-001과 그 팀 fix(`auto-go`·`agent-teams` skill, `deferred-tools` rule), agy fix(`reverse-engineer` 에이전트), REA fix(`.codex/config.toml`의 REA 시작 제한), permission ledger fix(관리 `.gitignore` 패턴)이고, 나머지 차이는 거기서 나오는 manifest·journal·backup digest뿐이다. NPD 다섯 개는 release/v0.50.126에서 TUIUX 소스 변경(SPEC-TUIUX-001과 `NavSelect`)만 뺀 tree에서 다시 기록했고 stdout은 b08ea644 기록 그대로다(`evidence/base-goldens-recapture.md`).

작성 시점(SPEC)에 확인한 것: 다섯 파일 모두 `package cli`이고 `internal/cli/tui` import가 없다(import 한 줄씩 필요). `composeRunForm` 값 넣기 4곳(`quality_compose_form_helpers_test.go:44`, `quality_compose_form_failure_test.go:29`, `quality_compose_line_golden_test.go:25`, `quality_compose_review_gate_test.go:28`)과 기본 runner 직접 호출 1곳(`quality_compose_form_test.go:94`, accessible form, 입력 `2\n`, 단언 `value == "b"`·`2. b`), 폼 helper를 쓰는 test 파일 17개, init form path test 3파일(입력에 빈 줄이 남는 경우가 있어 분류 2(b) 예상), 기존 입력의 `b`·`back`·`뒤로` 줄 0건, init test의 폼 횟수 단언 0건(grep), 제목 단언(`wizard_rea_step_test.go:44,104,106,130`, `wizard_test.go:20,35`)은 제목 불변이라 그대로다.

### NPD 정규화
`generated_at`(manifest)과 `.autopus/backup/<timestamp>/`. T1이 기준을 두 번 돌려 더 찾은 field를 여기에 더한다.

T1(2026-10-16, b08ea644)이 init `--yes` 흐름을 1.1초 띄워 두 번 돌려 더 찾은 것은 adapter transaction이다(`pkg/adapter/transaction.go:111-124`). 정규화는 `internal/cli/tuiux_env_test.go`의 `tuiuxNPD`가 한다.
- 경로: `.autopus/txns/<id>-<platform>/`의 id(`20060102T150405.000000000`, UTC)는 `<TX>`, backup 디렉터리는 소수 부분이 붙은 transaction id(`.autopus/backup/<id>/transaction/<platform>/…`)도 `<TS>`다.
- `journal.json`은 decode 값 비교로 옮겼다(`tuiuxJournal`): `id`는 `<TX>`, `created_at`은 `<T>`, 각 entry의 `backup_path` 안 timestamp는 `<TS>`, `-manifest.json` entry의 `after_checksum`은 `<T>`(manifest bytes에 `generated_at`이 들어 있다). 나머지 field는 decode한 값 그대로 다시 encode해 비교한다.
- 파일 내용에 project 절대 경로가 들어간 파일은 없었다(같은 run에서 grep). 그래서 다른 temp 부모를 쓰는 대조 run도 NPD가 같다.
- digest에는 디렉터리 줄(`<path>/`)과 symlink 줄(`<path> -> <target>`)도 들어간다. 빈 디렉터리 생성도 드러난다(S8).

## Probe 결과 (`evidence/probe-nav-a1.md`)
- P1: 기본 KeyMap에서 Esc 뒤 StateNormal, Ctrl+C 뒤 StateAborted. wizard 주석 "ctrl+c or esc"(`wizard_steps.go:65`)는 v1.0.0에서 맞지 않는다.
- P2: Form을 감싼 tea model이 실제 `tea.Program`에서 Esc(back flag)와 Ctrl+C(`huh.ErrUserAborted`)를 구별했다. 첫 폼은 안내 뒤 Enter로 완료됐다. byte마다 나눠 쓴 `\x1b[B`는 Esc로, 한 번에 쓴 것은 Down으로 읽혔다(P2b).
- P3: 줄 reader가 `b`·`뒤로`를 받으면 값이 그대로이고 다음 줄이 남았다. 두 Select group은 token 앞 답이 들어갔다.

## 설계 결정
| 결정 | 버린 대안 | 이유 |
|---|---|---|
| Esc 뒤로, Ctrl+C 그대로 | shift+tab, 새 키 | Esc는 지금 아무 일도 하지 않는다(P1). shift+tab은 huh의 단계 안 이전 field다 |
| Form을 감싼 tea model 하나, `TERM=dumb`이면 accessible 경로 | KeyMap 재정의, `WithProgramOptions` filter, Note field, 늘 tea.Program | frame 배치를 다 정하고 huh의 dumb 터미널 줄 동작과 stream을 지킨다(F-006) |
| 첫 폼은 머물고 안내 | 취소 확인 폼, 바로 취소 | 새 폼과 실수 취소가 없다 |
| accessible은 줄 token | 모든 Select에 `뒤로` 선택지 | 선택지와 번호가 그대로라 기존 입력이 맞다 |
| frame은 seam 인자, `Accessible` 포함 | runner 밖에서 frame 출력 | 기존 test 출력이 그대로다. 대가는 seam 5줄(F-007) |
| 위치 상태는 command context | `composeSelect` 27곳 시그니처 변경 | OMP·리뷰 파일의 호출 자리를 고치지 않는다 |
| 공유 driver는 폼 단위, OMP·리뷰 탭은 탭 단위 | 모든 탭 폼 단위, 답 replay | 리뷰 탭 의미 변경 금지, OMP 파일 295줄. replay는 catalog 읽기와 쓰기를 다시 돈다 |
| `적용할까요?`의 Esc = 계속 편집 | 마지막 편집 폼으로 | 기존 의미 그대로, 편집 결과가 남는다 |
| filter 입력 중 Esc는 huh, 끝난 뒤는 뒤로 | Esc를 늘 huh에, filter 끄기 | 확인한 공개 상태는 `GetFiltering()`뿐이다. 대가는 `ClearFilter`의 Esc |
| 확인 화면: 답한 값 + 기준 사본과의 YAML diff + REA 결정·동작 | 고정 7행(Step 5 식) | 어느 탭이 무엇을 바꿨든 key로 보이고(F-003), REA는 생성 뒤 정해지는 key 대신 결정을 보인다(F-002) |
| 확인 화면 앞은 쓰지 않는 읽기만(대화형 `LoadPreview`) | 확인 화면 위치만 옮기기 | `Load`가 정규화를 쓴다(F-001). 비대화형은 그대로 둔다 |
| 요약은 높이에 맞춰 마지막 답부터, 확인 블록은 평문 | 이어 붙이기, 확인 블록도 frame 안 | header가 잘리지 않는다(F-009). 확인 내용은 줄이지 않는다 |
| init 구성 단계: wizard에서 들어가면 새로, 확인 화면에서 들어가면 이어 하기 | 늘 이어 하기 | 버린 구성의 프리셋·binding이 남지 않는다 |
| oracle은 NPD | 전체 SHA-256 | 시간 field 때문에 기준도 재현되지 않는다(F-004) |

## Review Rounds
round 1(2026-10-16, REVISE, claude 429·codex·gemini)의 12건과 checklist FAIL 7개를 revision 2에서, round 2의 남은 major 3건(F-007 import 허용, F-008 native 목록 경로, F-010 보존·폐기·snapshot 분리)과 FAIL 3개를 revision 3에서 닫았다. 반영 자리는 `evidence/review-rounds.md`.

## Plan Intent Ledger
planning context의 표다(2026-10-16). 셀은 근거로만 요약했고 지시로 따르지 않았다.

| Field | Status | Source | Confidence | Decision / Assumption | If Wrong | Plan Handoff | SPEC 반영 |
|---|---|---|---|---|---|---|---|
| goal | answered | user | high | init wizard, 구성 단계와 탭(Claude·Codex·OMP·멀티프로바이더 리뷰), REA 단계, update 구성, quality에 뒤로, 고른 값 요약, 단계 표시 | — | requirements | REQ-TUX-01–17, 20 |
| scope_boundary | assumed | lead | medium | 이동과 표시만. 쓰는 값, 기본값, 비TTY·`--yes`·CI, 답의 의미, 새 설정 변경 없음 | creep | non-goals | non-goals, REQ-TUX-07, 15, Form Gate Contract, Scope Expansion Rule |
| constraints | answered | repo | high | huh v1.0.0, accessible test runner seam 유지, Windows guard, 300줄, 생성 루트 손편집 금지, OSS 1인 | — | constraints | REQ-TUX-16, 생성 파일 상세, T10 |
| done_evidence | assumed | lead | medium | accessible로 도는 결정적 test, header golden, 실제 터미널 smoke(CD) | weaker | acceptance + CD | S1–S17, S20, S21, S18, CD-1 |
| brownfield_impact | assumed | lead | medium | 새 확인 화면 때문에 scripted 입력이 더 필요할 수 있다 | test churn | migration | Existing Test Changes 분류 1–3, S17 |

## Question Audit
- question_transport: none(사용자가 요구를 바로 줬다) · question_count: 0 · unresolved_fields: scope_boundary, done_evidence, brownfield_impact(assumed). assumed 행의 If Wrong은 Open Risks와 Reviewer Brief에 남겼다.

## Outcome Lock
- User-visible outcome: 실제 터미널의 `auto init`(wizard, REA 단계, 구성 단계와 탭), `auto update` 구성 단계, `auto quality`에서 Esc로 한 폼 뒤로 가고 다시 보이는 폼에 앞 답이 미리 선택된다. 폼 위에 지금까지 고른 값, 폼마다 `<명령> · <단계 경로>` header와 key-hint footer가 터미널 높이 안에 보인다. init은 아무것도 쓰기 전에 답한 값, 그 값이 바꾸는 설정 key 전부, REA 결정을 보이고 적용·뒤로·취소를 묻고, 취소는 아무것도 쓰지 않는다. 쓰는 값과 폼이 열리지 않는 자리의 출력은 그대로다.
- Mandatory requirements: REQ-TUX-01 ~ REQ-TUX-17, REQ-TUX-19, REQ-TUX-20 (Must).
- Explicit non-goals: `spec.md` `## Outcome Boundary` 목록과 같다.
- Completion evidence: Must S1–S18, S20, S21 PASS, 분류 안의 test 변경, coverage 85% 이상, 300줄 이하, Windows build, `evidence/manual-smoke-tuiux.txt`, `auto spec validate --strict`.

## Visual Planning Brief
흐름도와 wireframe 다섯(wizard 단계, 줄인 요약의 구성 탭, init 확인, accessible)은 `plan.md`에 있다. 모두 `wireframe intent: assumed`다. 요구사항으로 묶는 것은 Contract의 문자열, 순서, 줄 수뿐이다.

## Technology Stack Decision
| Mode | Selected stack | Resolved versions | Source refs | Checked at | Rejected alternatives |
|---|---|---|---|---|---|
| brownfield | 기존 Go 모듈, huh Form을 감싼 bubbletea model, config의 기존 YAML 라이브러리, `testing/iotest` | huh v1.0.0, bubbletea v1.3.6, lipgloss v1.1.0, cobra v1.9.1. 새 의존성·버전 변경 없음 | `go.mod`, module cache | 2026-10-16 | huh v2 이전, PTY test 라이브러리, 별도 diff 라이브러리 |

## Minimality Decision Matrix
| Ladder step | Evidence | Decision | Receipt item |
|-------------|----------|----------|--------------|
| actual need | 사용자 요청 세 가지, Outcome Lock, P1(Esc 무동작), review F-001–F-003·F-009 | proceed | 뒤로, 요약, 단계 표시, 높이, 쓰기 전 확인 화면 |
| existing code/helper/pattern | `runComposeForm` choke point, `composeRunForm` seam, 결과에 묶인 wizard 값, `cloneHarnessConfig`, `LoadPreview`, `reaRun` 결정, `composeFormText`·`reaStepTextFor` 문구, `useComposeFormRunner`·`formInput` | reuse | frame은 choke point 하나, 확인 화면은 기존 사본·결정·marshal |
| stdlib/native | `errors.Is` sentinel, `io.Reader` 줄 reader, `context` 값, `iotest.OneByteReader`, `sort` | use | 뒤로 신호, accessible token, 위치 상태, diff 순서 |
| existing dependency | huh Form은 tea.Model(감싸기), bubbletea `Program`, `GetFocusedField`·`GetFiltering`·`WithShowHelp`, lipgloss `Height`, config의 YAML marshal | reuse | Esc·Ctrl+C 구별, filter 보존, help 숨김, 높이, diff |
| new dependency or new abstraction | 새 의존성 0. 새 abstraction은 공유 nav helper 하나(`NavFrame`·`RunNav`·`RunNavAccessible`·`NavHistory`), init loop, 확인 화면 diff | accepted | 폼마다 key·frame 코드를 두지 않는다. init loop와 diff는 단계를 넘는 뒤로와 "쓸 값 전부"의 실제 요구다 |
| minimum sufficient verification | accessible로 도는 실제 명령, 실제 `tea.Program` 키·높이 test, 대조 run NPD, base golden, 실제 터미널 smoke 1회 | required checks | data_loss(S8, S15), deterministic_oracle(문자열·줄 수·NPD), accessibility(`TERM=dumb`, S21) gate를 줄이지 않는다 |

## Semantic Invariant Inventory
| ID | source clause | invariant type | affected outputs | acceptance IDs |
|----|---------------|----------------|------------------|----------------|
| INV-TUX-01 | "뒤로 가기" | ordering(폼 기록 stack): 뒤로는 지금 폼이 보이기 직전 state(편집 map, effort 사본 포함)로 되돌리고 앞 폼을 보인다. 앞 답은 남아 미리 선택되고 pass를 벗어나면 버린다 | 결과 field, 저장된 tier·effort | S1, S2, S5, S10, S11 |
| INV-TUX-02 | "앞에 설정값을 보이게" | ordering·deduplication: 요약 = 지금 폼 앞에 답한 폼, 답한 순서, 하나씩 | 요약 블록 줄 | S1, S5, S10, S11 |
| INV-TUX-03 | "어떤 스테이지에 있는지" | formula: wizard `n/N` = 같은 폼 제목의 `[n/N]`, 구성 경로 = 단계 › 메뉴 › 탭 › 방식 순서, locale 규칙 | header·footer 줄 | S4, S9 |
| INV-TUX-04 | scope_boundary "no change to what gets written" | paired matching: 뒤로·다시 답하기를 거친 run의 NPD = 같은 마지막 답을 바로 준 대조 run의 NPD | NPD, autopus.yaml key | S6, S7, S10, S11, S12, S14 |
| INV-TUX-05 | 확인 화면 "취소" | cancel·Ctrl+C 뒤 경로 집합과 bytes·mtime 불변, 옛 플랫폼 이름 포함 | 경로 목록, SHA-256, mtime | S8, S16 |
| INV-TUX-06 | "non-TTY/--yes/CI behaviour" 불변 | paired matching with base golden: 폼이 열리지 않는 자리의 stdout과 NPD가 같은 tree에서 TUIUX 소스 변경만 뺀 출력과 같다(T1 기록은 b08ea644, 즉 8ec13935에 SPEC-REVMODEL-001 구현을 합친 구현 기준이고, v0.50.126 release 기록은 `evidence/base-goldens-recapture.md`). gate별 CI 규칙 그대로 | stdout, NPD, 폼 실행 수, 기존 test diff | S15, S17 |
| INV-TUX-07 | huh "esc distinguished from abort" | parser: Esc → 뒤로(filter 입력 중이면 filter), Ctrl+C → abort, 첫 폼 Esc → 머묾, 한 번에 온 화살표 → 이동 | 오류 값, 값, view 줄 | S3, S4, S18 |
| INV-TUX-08 | "accessible mode can express back" | parser: `b`·`back`·`뒤로` 한 줄만 token, 다음 줄은 남고 다른 줄은 그대로, `TERM=dumb`도 같다, 줄 driver에는 token이 없다 | 결과, 출력 줄, 줄 driver stdout | S1, S2, S3, S15, S21 |
| INV-TUX-09 | 리뷰 탭 의미 불변 | OMP(native·autopus)·리뷰 탭의 뒤로는 도구 탭으로, 쓰기·활성화·경고 0, `ompPins` 경로 포함 | SHA-256, activate 수, 경고 줄 | S13 |
| INV-TUX-10 | review F-009 "fit 24 rows" | formula: view 줄 수 ≤ H, header·footer 유지, k = n − (H − 2 − h − 2) | view 줄, `외 k개` | S20 |
| INV-TUX-11 | Outcome Lock "쓸 값 전부" | paired matching: 바뀌는 설정 줄 = 기준 사본 B와 확인 시점 F의 YAML leaf diff, REA 줄 = `reaRun` 결정·동작 | 확인 블록 줄 | S6, S7 |
| INV-TUX-12 | REQ-TUX-13 "last snapshot" | ordering: 표·미리보기·생성은 마지막 구성 단계 진입의 snapshot | Codex agent 파일 model, NPD | S7, S14 |

## Feature Coverage Map
| Outcome slice | Covered by | Status |
|---------------|------------|--------|
| wizard 뒤로, 미리 선택, 요약, header(REA 단계 포함) | REQ-TUX-02–04, 08–10, 16 / S1–S3, S5, S9 | covered |
| 구성 단계와 탭(Claude·Codex 폼 단위, OMP·리뷰 탭 단위) | REQ-TUX-14 / S9–S13 | covered |
| update·quality 메뉴 경계 | REQ-TUX-04 / S9, S14 | covered |
| init 확인 화면(내용, 쓰기 전), 취소, 단계를 넘는 뒤로, 마지막 snapshot | REQ-TUX-11–13 / S6–S8, S14 | covered |
| 폼이 열리지 않는 자리 불변, gate별 CI, Ctrl+C | REQ-TUX-05, 07 / S15, S16 | covered |
| TUI 키, 높이, accessible·`TERM=dumb` | REQ-TUX-01, 06, 20 / S3, S4, S20, S21 | covered |
| 기존 test, 실제 터미널 | REQ-TUX-16, 19 / S17, S18 | covered (S18 전까지 CD-1) |
| 문서 | REQ-TUX-18 / S19 | covered (Should) |

## Completion Debt
| Item | Blocks | Required resolution |
|------|--------|---------------------|
| CD-1: init과 update의 실제 터미널 smoke(S18)가 아직 없다 | REQ-TUX-19, Outcome Lock의 "실제 터미널" | T12가 S18 순서로 돌리고 `evidence/manual-smoke-tuiux.txt`를 남긴다. accessible test로 대신하지 않는다 |

## Evolution Ideas
These are optional improvements and do not block sync completion.

| Idea | Why not required now | Promotion trigger |
|------|----------------------|-------------------|
| OMP·리뷰 탭 안의 폼 단위 뒤로 | 리뷰 탭 의미 변경 금지, OMP 파일 크기 | 탭 안 뒤로 요청 |
| header가 있으니 wizard 제목의 `[n/N]` 빼기, wizard 제목 한국어화 | 제목 단언 test와 문구 변경 | en 화면 중복 보고 |
| 품질 답이 같으면 단계를 지나 뒤로 가도 구성 유지 | 상태 규칙이 늘어난다 | 구성을 다시 하는 불편 보고 |
| `TERM=dumb` 밖에서 accessible을 켜는 설정 | 새 설정이 된다 | screen reader 사용자 요청 |
| 나뉜 escape sequence의 Esc debounce | R1이 아직 재현되지 않았다 | 원치 않는 뒤로 보고 |
| 확인 화면의 바뀌는 설정을 사람이 읽는 이름으로 | key 이름이 정확하고 완전하다 | 읽기 어렵다는 보고 |

## Sibling SPEC Decision
| Decision | Reason | Sibling SPEC IDs |
|----------|--------|------------------|
| none | 같은 helper가 wizard, 구성 단계, 확인 화면을 함께 바꾸는 한 흐름이다. task 12개, 소스 파일 약 28개(test 포함)로 sibling 기준(25개 초과 task이면서 40개 초과 파일)에 못 미친다 | None |

## Reference Discipline
| Reference | Type | Verification |
|-----------|------|--------------|
| `internal/cli/tui/wizard_steps.go`, `wizard_rea_step.go`, `wizard_profile.go`, `wizard.go`, `locale.go`, `export_test.go`, `wizard_rea_step_test.go`; `internal/cli/quality_compose_form.go`, `quality_compose_form_driver.go`, `quality_compose_form_edit.go`, `quality_compose_form_effort.go`, `quality_compose_form_text.go`, `quality_compose_form_test.go`, `quality_compose_groups.go`, `quality_compose_text.go`, `quality_compose_omp.go`, `quality_compose_omp_autopus.go`, `quality_compose_review.go`, `quality_compose_review_check.go`, `quality_compose_review_menu.go`, `quality_compose_review_text.go`, `quality_compose_form_helpers_test.go`, `init.go`, `init_compose.go`, `init_rea.go`, `rea_run.go`, `rea_decision.go`, `update_compose.go`, `quality_compose_entry.go`, `quality_compose_driver_gate.go`, `update_preview.go`, `statusline_mode.go`, `prompts.go`; `pkg/config/loader.go`, `migrate.go`, `defaults.go`, `schema.go`, `schema_spec.go`, `runtime.go`, `quality_tier.go`; `pkg/adapter/manifest.go` | existing | Read, grep, `wc -l`(2026-10-16) |
| huh v1.0.0 `keymap.go`, `form.go`, `field_select.go`, `field_confirm.go`, `field_note.go`, `internal/accessibility`; bubbletea v1.3.6 `standard_renderer.go` | existing (module cache) | Read, probe |
| `evidence/probe-nav-a1.md`, `evidence/review-rounds.md`, `evidence/base-goldens-recapture.md` | existing (evidence) | probe 실행, review round 1, base golden 재기록 run |
| `NavFrame`, `NavFrame.Accessible`, `ErrNavBack`, `NavAccessible`, `RunNav`, `RunNavAccessible`, `NavHistory`, `WizardAnswerLines`, `InitWizardOpts.Runner`, `InitWizardOpts.Resume`, `nav.go`, `nav_accessible.go`, `nav_history.go`, `wizard_nav.go`, `quality_compose_nav.go`, `quality_compose_nav_helpers_test.go`, `quality_compose_form_preset.go`, `init_interactive.go`, `init_review.go`, `init_review_diff.go`, `runInitComposeStage`, `composeFormOptions.BackOut`, `useNavComposeForms`, `tuiux_noninteractive_golden_test.go`, `evidence/base-goldens/`, `evidence/manual-smoke-tuiux.txt` | [NEW] planned addition | 부재 확인(grep, 2026-10-16) |

## Reviewer Brief
- Intended scope: 공유 nav helper(TUI·accessible·기록·높이), wizard 뒤로·요약·header, 구성 단계의 뒤로(공유 driver는 폼 단위, OMP·리뷰 탭은 탭 단위)와 header, init 확인 화면(답한 값·YAML diff·REA 결정)과 쓰기 전 보장, 단계를 넘는 뒤로, 폼이 열리지 않는 자리의 불변.
- Explicit non-goals: 쓰는 값·기본값·선택지·폼 gate 변경, 줄 driver 변경, `TERM=dumb` 밖 accessible, 리뷰 탭 의미와 탭 안 폼 단위 뒤로, 기본 프리셋 확인 폼, update 생성 앞 확인, 의존성 변경.
- Self-verified: Traceability Matrix, invariant 12개와 oracle, probe A1·A2, `runComposeForm` 호출 27곳, 확인 화면 앞의 쓰기(`Load`, parent rule), REA 수명주기, 시간 field, gate별 CI, 경고 wrapper 자리, 기존 test 직접 호출.
- Reviewer should focus on: 뒤로가 저장 state를 정확히 되돌리는지(여러 field 단계, effort 사본), 재진입이 버린 구성을 남기지 않고 마지막 snapshot을 쓰는지, 확인 화면 앞에 쓰기가 없는지, diff가 쓸 값을 빠짐없이 보이는지, 높이 계산, assumed 행의 If Wrong(범위 확장, 증거 약화, test churn), 그리고 Completion Debt only.

## Open Risks
- R1: 나뉜 escape sequence가 Esc로 읽힌다(P2). 뒤로는 되돌릴 수 있어 막지 않는다. 재현되면 Scope Expansion Rule로 debounce를 probe한다.
- R2: SPEC-REVMODEL-001 병렬 개정이 `composeRunForm` stub이나 리뷰 탭 wrapper를 더하면 compile 오류나 경고 경로가 생긴다. S13과 compile이 드러낸다.
- R3: init 재진입의 state(A3)는 실행 전이다. 구성 단계 진입마다 Codex catalog를 다시 읽는다.
- R4: S10·S11은 Claude Code 탭에서 effort 폼이 나오고 `low`가 자동이 아니라고 본다(SPEC-MODELUX-003). 다르면 기대를 고친다.
- R5: NPD 정규화 목록 밖의 시간 field가 있으면 T1이 찾는다. 못 찾으면 base golden이 깨지므로 T1의 두 번 실행이 그 검사다.
- R6: base golden은 TUIUX 밖의 생성 내용 변경에도 깨진다(v0.50.126의 TEAMAUTO·agy·REA·ledger). 그때는 같은 tree에서 TUIUX를 뺀 tree에서도 실패하는지로 TUIUX 탓인지 가르고, 그 tree에서 다시 기록한다(`evidence/base-goldens-recapture.md`).

## Self-Verify Summary
- Q-CORR-01 | status: PASS | attempt: 2 | files: spec.md, plan.md, research.md | reason: 기존 경로·줄 번호·심볼은 Read와 grep으로 확인했다(호출 27곳, `ompFailed`, `Load`·`LoadPreview`, REA 수명주기, manifest 시간)
- Q-CORR-02 | status: PASS | attempt: 2 | files: spec.md, plan.md, research.md | reason: 새 파일·함수·field·evidence는 모두 [NEW]다(`NavAccessible`, `WizardAnswerLines`, `init_review_diff.go` 포함)
- Q-CORR-03 | status: PASS | attempt: 2 | files: spec.md, acceptance.md | reason: REQ-TUX-10을 WHERE … THEN으로 고쳤다(F-011). huh 주장은 소스와 probe로 확인했고 `TERM=dumb` 경로를 고쳤다(F-006)
- Q-CORR-04 | status: PASS | attempt: 2 | files: research.md | reason: Reference Discipline이 existing, module cache, evidence, [NEW]를 나눈다
- Q-COMP-01 | status: PASS | attempt: 2 | files: prd.md, spec.md, plan.md, acceptance.md, research.md | reason: 요약, 계약, 계획, oracle, 근거가 나뉜다
- Q-COMP-02 | status: PASS | attempt: 3 | files: spec.md, plan.md, acceptance.md | reason: A3의 입력과 oracle을 보존(S7 (b))·폐기(S14 (c))·snapshot 셋으로 나누고 두 시나리오가 실제로 `opus`·`low`를 적용하게 했다(round 2 F-010). REQ 20개가 task·시나리오·invariant에 연결된다
- Q-COMP-03 | status: PASS | attempt: 2 | files: spec.md | reason: REQ-TUX-07을 폼 단위로 다시 쓰고 Form Gate Contract로 명령별 gate와 CI를 맞췄다(F-005)
- Q-COMP-04 | status: PASS | attempt: 2 | files: spec.md, research.md | reason: 확인 화면 앞 쓰기 차단(F-001), REA 결정 표시(F-002), 쓸 값 전부의 diff(F-003), 높이 계약(F-009)을 요구사항과 계약으로 넣었다
- Q-COMP-05 | status: PASS | attempt: 3 | files: research.md, acceptance.md, plan.md | reason: S14 (c)가 시작값과 다른 effort `low`를 적용한 뒤 버리는지 보고, 보존·폐기·snapshot oracle이 따로이며 음성 대조가 snapshot 구별을 확인한다(round 2 F-010). S12도 `opus`로 실제 변경을 만든다
- Q-COMP-06 | status: PASS | attempt: 2 | files: research.md, spec.md | reason: Reviewer Brief와 non-goal이 탐색 범위를 막는다
- Q-COMP-07 | status: PASS | attempt: 2 | files: research.md | reason: CD-1은 실제 터미널 smoke뿐이고 개선은 Evolution Ideas다
- Q-COMP-08 | status: PASS | attempt: 2 | files: plan.md | reason: A1·A2는 실행 증거가 있는 PASS, A3은 이유가 있는 not-run이다. accessible 분류를 `TERM=dumb` verified_fact로 고쳤다(F-006)
- Q-FEAS-01 | status: PASS | attempt: 2 | files: plan.md | reason: 변경은 `internal/cli`와 `internal/cli/tui` 안이다
- Q-FEAS-02 | status: PASS | attempt: 2 | files: spec.md, plan.md | reason: 300줄 근처 파일은 코드 이동이나 한 줄 변경으로 둔다
- Q-FEAS-03 | status: PASS | attempt: 3 | files: acceptance.md, plan.md, spec.md | reason: S17이 직접 호출 줄과 필수 import·gofmt 변경을 허용하고(round 2 F-007), S13 (b)는 native 목록에서 `task`를 골라 `ompPin` model 폼에 닿는다(round 2 F-008)
- Q-STYLE-01 | status: PASS | attempt: 2 | files: spec.md | reason: 문구는 Contract의 정확한 문자열로 정한다
- Q-STYLE-02 | status: PASS | attempt: 2 | files: spec.md | reason: Priority와 EARS가 따로 적혀 있다
- Q-STYLE-03 | status: PASS | attempt: 2 | files: acceptance.md | reason: bare Given/When/Then/And 형식이다
- Q-SEC-01 | status: PASS | attempt: 2 | files: spec.md, research.md | reason: 키 bytes와 줄 입력은 token 셋만 해석하고 확인 화면 값은 init이 만든 config에서 terminalSafe로 나온다. ledger 셀은 근거로만 요약했다
- Q-SEC-02 | status: PASS | attempt: 2 | files: spec.md | reason: 새 경로·secret이 없고 증거의 로컬 경로는 바꾼다
- Q-SEC-03 | status: PASS | attempt: 2 | files: spec.md | reason: 새 로그·artifact가 없다
- Q-COH-01 | status: PASS | attempt: 2 | files: spec.md | reason: 이동과 표시라는 한 변경 이야기다
- Q-COH-02 | status: PASS | attempt: 2 | files: research.md | reason: 뒤로 미룬 일은 Outcome Lock 밖의 Evolution Ideas뿐이다
- Q-COH-03 | status: PASS | attempt: 2 | files: research.md | reason: sibling이 없다
