# SPEC-TUIUX-001: 설정 TUI의 뒤로 가기, 고른 값 요약, 단계 표시

**Status**: completed
**Created**: 2026-10-16
**Revision**: 3 (review round 1의 F-001–F-012, round 2의 F-007·F-008·F-010 반영. 반영표 `evidence/review-rounds.md`)
**Domain**: TUIUX
**Module**: autopus-adk
**Release**: v0.50.126 (SPEC-REA-001, SPEC-REVMODEL-001, TEAMAUTO 작업과 같은 release). branch `feat/tuiux-001`, 기준 b08ea644(`feat/revmodel-001` 위)
**Source PRD**: `prd.md`(같은 디렉터리). 코드 근거, probe, 결정은 `research.md`와 `evidence/probe-nav-a1.md`에 있다.

## 목적

사용자 말(2026-10-16): "TUI를 좀 더 UX적으로 사용자가 편하게 해줘. 엄청 많은 걸 설정해야 하니까 뒤로 가기라든지, 앞에 설정값을 보이게 하든지, 어떤 스테이지에 있는지… 이런 UX도 신경써."

지금은 세 가지가 없다. 뒤로 갈 수 없다. wizard는 단계마다 따로 `form.Run()`을 돌리고 오류가 나면 wizard 전체를 취소한다(`internal/cli/tui/wizard_steps.go:60-69`). huh v1.0.0의 기본 KeyMap에서 Esc는 아무 일도 하지 않는다(probe P1). 앞에서 고른 값이 보이지 않는다. 어디쯤인지는 wizard 제목의 `[n/N]`뿐이고 구성 단계와 그 탭에는 위치 표시가 없다.

이 SPEC은 표시와 이동만 바꾼다. 하나의 공유 navigation helper가 init wizard, 구성 단계(모든 탭), 새 init 확인 화면의 폼을 돌린다. Esc(accessible에서는 `b`)는 한 폼 뒤로, Ctrl+C는 지금처럼 전체 취소다. 폼마다 위치 header 한 줄, 지금까지 고른 값, key-hint footer가 터미널 높이 안에 보인다. init은 아무 파일도 쓰기 전에 확인 화면(적용·뒤로·취소)에서 답한 값과 그 값이 바꾸는 설정을 보인다. 무엇을 쓰는지, 기본값, 폼이 열리는 조건은 바뀌지 않는다.

## Outcome Boundary

- Outcome Lock 원문은 `research.md` `## Outcome Lock`에 있다. 요약: 실제 터미널의 `auto init`(wizard, REA 단계, 구성 단계와 그 탭), `auto update`의 구성 단계, `auto quality`에서 (1) Esc로 한 폼 뒤로 가고 다시 보이는 폼에는 앞 답이 미리 선택된다, (2) 지금까지 고른 값이 폼 위에 보이고, init은 쓰기 전 확인 화면에서 답한 값, 그 값이 바꾸는 설정 key 전부, REA 결정을 보인 뒤 적용·뒤로·취소를 묻는다(취소는 아무것도 쓰지 않는다), (3) 폼마다 `<명령> · <단계 경로>` header와 key-hint footer가 보인다.
- Mandatory requirements: REQ-TUX-01 ~ REQ-TUX-17, REQ-TUX-19, REQ-TUX-20 (Must). REQ-TUX-18은 Should다.
- Explicit non-goals:
  - 쓰는 값, key, 기본값, 선택지의 의미와 순서 변경. 새 설정, 새 flag, 새 YAML key
  - 폼이 열리는 조건(각 명령의 기존 gate, Input Mode Contract) 변경과 SPEC-MODELUX-001 줄 driver의 출력·키·경로 변경. 줄 driver에는 뒤로가 없다
  - `TERM=dumb` 밖에서 accessible mode를 켜는 새 경로나 설정
  - SPEC-REVMODEL-001 리뷰 탭의 의미 변경. 리뷰 탭과 OMP 탭 안의 폼 단위 뒤로(이 SPEC은 탭 단위 뒤로만 준다)
  - `기본 프리셋 고르기`에 확인 폼 추가, `auto update`의 플랫폼 생성 앞 확인 화면, statusLine·parent rule 줄 질문 변경
  - huh·bubbletea·lipgloss 버전 변경, wizard 단계 제목(`[n/N] …`) 문구 변경
- Completion evidence: `acceptance.md`의 Must 시나리오(S1–S18, S20, S21) PASS, 기존 test 변경은 `## Existing Test Changes`의 분류 안에서만, `go test ./internal/cli/... ./pkg/...`, 신규·변경 package coverage 85% 이상, 신규·변경 소스 파일 300줄 이하, `GOOS=windows GOARCH=amd64 go build ./...`, 실제 터미널 smoke 기록(`evidence/manual-smoke-tuiux.txt`, S18), `auto spec validate --strict` 통과.

## Trust Model

- 입력은 셋이다. TUI 키 bytes(bubbletea가 해석), accessible 줄 입력(`TERM=dumb` 터미널, scripted test), autopus.yaml과 catalog에서 온 라벨. 라벨은 지금처럼 `terminalSafe`를 거친다(SPEC-MODELUX-002 REQ-MUX2-02). header, 요약, 확인 화면의 key·값도 같다.
- 확인 화면의 값은 init이 새로 만든 config(`DefaultFullConfig`, flag, 감지, 답)의 YAML에서 온다. 사용자 autopus.yaml은 대화형 경로에서 쓰지 않는 읽기(`config.LoadPreview`)로만 읽는다.
- accessible 뒤로 token은 정확히 `b`, `back`, `뒤로` 한 줄뿐이다. 다른 줄은 지금처럼 huh에 그대로 간다. 이 표면들의 field는 Select와 Confirm뿐이라(비test 코드에 `huh.NewInput`·`NewText`·`NewMultiSelect` 0건) token은 유효한 답과 겹치지 않는다.
- 새 파일·로그·artifact는 없다. 증거 파일은 커밋 전에 `$HOME`, scratch 경로, 사용자 이름을 placeholder로 바꾼다.

## Requirements

### REQ-TUX-01 하나의 navigation helper
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL run every form of the init wizard, of the composition step in init, update, and quality, and of the init review through one shared navigation helper that renders the frame, maps the keys of the Navigation Contract, and keeps the state saved before each form.

- `[NEW] internal/cli/tui/nav.go`의 `NavFrame`, `ErrNavBack`, `NavAccessible`, `RunNav`, `[NEW] nav_accessible.go`의 `RunNavAccessible`, `[NEW] nav_history.go`의 `NavHistory[S]`가 그 helper다. wizard와 구성 단계는 이 helper만 부르고 따로 key 처리나 frame 출력을 하지 않는다.
- 구성 단계는 `composeRunForm`을 지나는 기존 choke point(`runComposeForm`, `quality_compose_form.go:102-111`) 하나에서 frame을 만든다. 위치 상태는 `runForms`가 command context에 한 번 싣고 `runComposeForm`이 읽으므로 `composeSelect`·`composeConfirm`의 시그니처와 27개 호출 자리는 그대로다.

### REQ-TUX-02 Esc는 한 폼 뒤로
Priority: Must · EARS: EventDriven

WHEN the operator presses Esc on a navigation form that is not the first form of its flow, THEN THE SYSTEM SHALL close the form, restore the state saved immediately before the form was shown, and show the form the Navigation Contract names for that form.

- Select가 filter 입력 중이면 Esc는 지금처럼 filter 입력을 끝낸다(huh `SetFilter`, 뒤로가 아니다). 판정은 focus된 field의 `GetFiltering()`이다(probe P2에서 compile 확인). filter 입력이 끝난 뒤의 Esc는 뒤로다. 그래서 남은 filter 글자를 지우던 huh의 Esc(`ClearFilter`)는 nav 폼에서 뒤로가 되고, 폼을 다시 열면 filter는 비어 있다.
- 같은 단계 안의 이전 field(언어 설정의 Select 3개)는 huh의 기존 shift+tab 그대로다.
- 한 escape sequence가 여러 read로 쪼개져 오면 bubbletea가 앞의 `\x1b`를 Esc로 읽는다(probe P2 `down+enter`). 뒤로는 값을 되돌리고 앞 답을 미리 선택하므로 되돌릴 수 있는 동작이다. Open Risk R1로 둔다.

### REQ-TUX-03 다시 보이는 폼의 미리 선택
Priority: Must · EARS: EventDriven

WHEN a form is shown again because the operator went back, or because the operator moved forward again after going back, THEN THE SYSTEM SHALL preselect the answer the state holds for that form, so that Enter keeps the earlier answer.

- wizard는 Select·Confirm이 결과 field에 묶여 있어 state 값이 곧 미리 선택이다(`wizard_steps.go:151`, `:179` 등 `Value(&r.…)`).
- 구성 단계의 등급·effort·에이전트 값 Select는 기본값 `그대로`가 state의 지금 값(앞 답 포함)을 유지하고 라벨이 그 값의 칸을 보인다. 뒤로로 다시 보이는 도구 탭·방식·목록 Select는 이 run에서 그 폼에 마지막으로 고른 선택지를 미리 선택한다. 앞으로 가서 새로 여는 폼의 기본값은 지금과 같다.

### REQ-TUX-04 첫 폼의 경계
Priority: Must · EARS: EventDriven

WHEN the operator goes back on the first form of a flow, THEN THE SYSTEM SHALL keep that form open with its values unchanged and show the first-step notice once for that key or line.

- 첫 폼은 wizard의 첫 단계, update·quality 구성 단계의 메뉴다. init 구성 단계의 메뉴는 첫 폼이 아니다(REQ-TUX-13).
- TUI는 footer 자리에 안내를 보이고 폼은 계속 돈다. accessible은 안내 한 줄을 쓰고 같은 질문의 다음 줄을 읽는다(probe P2 `esc-first-then-enter`).

### REQ-TUX-05 Ctrl+C는 그대로
Priority: Must · EARS: EventDriven

WHEN a navigation form receives Ctrl+C, THEN THE SYSTEM SHALL return huh.ErrUserAborted to its caller, so that init prints its existing cancel line and writes no file, update returns its existing non-zero interruption, and quality ends without error, as before this SPEC.

- SPEC-MODELUX-002 REQ-MUX2-14의 동작 그대로다. init 확인 화면의 Ctrl+C도 init 취소다.

### REQ-TUX-06 accessible 뒤로
Priority: Must · EARS: EventDriven

WHEN a navigation form runs in accessible mode and reads a line that equals b, back, or 뒤로 after trimming and ASCII case folding, THEN THE SYSTEM SHALL treat the line as Esc: the rest of the form keeps its values, the line after it stays unread for the next form, and the helper returns ErrNavBack, or shows the notice and reads the next line on a first form.

- production에도 accessible 경로가 있다. huh `NewForm`은 `TERM=dumb`이면 `WithAccessible(true)`를 부른다(huh `form.go:124-127`). `NavAccessible()`는 같은 규칙(`os.Getenv("TERM") == "dumb"`)으로 frame의 `Accessible`을 정하고, `RunNav`는 그때 accessible 경로로 간다. 그래서 `TERM=dumb`의 줄 기반 동작과 stream이 지금과 같다: wizard는 accessible이면 출력이 os.Stdout(huh의 accessible 기본, `form.go:331`, `:672`), 아니면 os.Stderr다. 구성 단계는 두 경우 모두 command의 stream이다.
- accessible huh에는 뒤로가 없다. `Form.runAccessible`이 field 오류를 버리고 `PromptInt`는 숫자가 아니면 다시 묻는다(huh `form.go:703-721`, `internal/accessibility`). helper는 입력을 줄 단위로 읽는 reader로 감싸 token 뒤로는 form이 끝날 때까지 EOF를 준다. EOF에서 huh는 기본값(지금 state 값)을 둔다(probe P3).
- 한 폼에 field가 여럿이면 token 앞의 답은 이미 state에 들어갔다(probe P3 group `a="ko"`). 그래서 호출자는 REQ-TUX-02의 저장 state로 되돌린다.

### REQ-TUX-07 폼이 열리지 않는 자리는 그대로
Priority: Must · EARS: Unwanted

IF a form did not open at a place before this SPEC, because the command's existing interactivity gate in the Form Gate Contract does not ask there or because the composition step uses the SPEC-MODELUX-001 line driver there, THEN THE SYSTEM SHALL open no navigation form at that place, print no frame, accept no back token, and keep that output byte-identical to b08ea644.

- navigation은 run 단위가 아니라 폼 단위다. 지금 폼이 열리는 자리는 모두 navigation 폼이 되고, 열리지 않는 자리는 그대로다. gate는 하나도 바꾸지 않는다(Form Gate Contract).
- 그래서 `auto init | tee log`(TTY stdin, pipe stdout)는 wizard가 지금처럼 stderr의 huh 폼이라 navigation을 받고, 구성 단계는 줄 driver이며 확인 화면은 없다. CI truthy는 지금처럼 update의 구성 단계만 막는다. init과 quality에서 CI는 판단에 들어가지 않는다.

### REQ-TUX-08 단계 header
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL show on every navigation form one header line in the Stage Indicator Contract format, built from the command name and the path of stages, menu choice, tab, and method the operator took to reach the form.

- wizard의 `<n>/<N>`은 같은 폼 제목의 `[n/N]`과 같은 값이다(`buildStepList`의 위치와 개수).

### REQ-TUX-09 key-hint footer
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL show the footer of the Stage Indicator Contract as the last line of every navigation form, without the back hint on a first form, and SHALL hide huh's own help line on those forms.

### REQ-TUX-10 지금까지 고른 값
Priority: Must · EARS: StateDriven

WHERE a wizard run or a composition edit pass has answered forms before the current form, THEN THE SYSTEM SHALL show above the current form a summary block that lists exactly those answers, one line each, in the order they were answered, in the Summary Contract format.

- 뒤로로 지나간 폼의 답은 다시 그 앞으로 오기 전까지 요약에 없다. 앞으로 가며 다시 답하면 다시 들어간다. 요약은 navigation 기록(`NavHistory`)에서 만든다.
- 구성 단계 메뉴, `적용할까요?`, OMP·리뷰 탭은 기존 표와 미리보기가 요약이다(Summary Contract). TUI에서 요약이 높이를 넘으면 Height Contract대로 줄인다.

### REQ-TUX-11 init 확인 화면
Priority: Must · EARS: EventDriven

WHEN interactive init on the form path has asked its last wizard or composition question, THEN THE SYSTEM SHALL, before any file is written, show the init review of the Init Review Contract with the answers, every setting key those answers change, and the REA decision, and ask 적용, 뒤로, or 취소, whose default is 적용.

- form path는 Form Gate Contract의 init 확인 행이다. 아니면 확인 화면 없이 지금처럼 진행한다.
- 확인 화면은 `rea.decideInit` 뒤, `config.MigrateOrchestraConfig`·statusLine 질문·parent rule 검사 앞이다. parent rule 검사는 충돌이 있으면 `config.Save`를 부르므로(`prompts.go:99-140`) 그 앞이어야 한다.
- 확인 화면 앞에서는 어떤 파일도 쓰지 않는다. 지금 init은 wizard 전에 기존 프로필을 `config.Load(dir)`(`init.go:99`)로 읽고, `Load`는 플랫폼 이름 정규화가 필요하면 autopus.yaml을 다시 쓴다(`pkg/config/loader.go:87-94`). 대화형 경로는 이 읽기를 `config.LoadPreview(dir)`(같은 정규화, 쓰기 없음)로 바꾼다. 비대화형 경로는 `config.Load` 그대로다(REQ-TUX-07). 확인 화면 앞의 다른 config 읽기(`prepareInitRea`)는 이미 `LoadPreview`다.
- 적용은 기존 흐름(Step 2–5)을 그대로 잇는다. Enter가 적용이라 확인 화면 앞에서 끝나던 scripted 입력은 EOF에서 적용을 받는다.

### REQ-TUX-12 취소는 아무것도 쓰지 않는다
Priority: Must · EARS: Unwanted

IF the operator chooses 취소 on the init review or presses Ctrl+C before it, THEN THE SYSTEM SHALL print the existing init cancel line, return without error, and leave every path under the project directory byte-identical with its mtime unchanged, creating none, also when the existing autopus.yaml holds a platform name that loading normalizes.

### REQ-TUX-13 init의 단계를 넘는 뒤로
Priority: Must · EARS: EventDriven

WHEN the operator goes back on the menu of init's composition stage, THEN THE SYSTEM SHALL restore the configuration to its state before the stage was entered and resume the wizard at its last step with every earlier answer preselected; and WHEN the operator chooses 뒤로 on the init review, THEN THE SYSTEM SHALL return to the composition menu with the configuration kept if the stage asked forms, and to the wizard's last step otherwise.

- 앞으로 다시 들어가는 구성 단계는 wizard 답을 적용한 새 cfg에서 새로 시작한다. 그래서 품질 답을 balanced로 바꾸면 버린 구성의 프리셋·binding이 남지 않는다. 기준 cfg 사본은 기존 `cloneHarnessConfig`(`update_preview.go:20`)로 만든다.
- 구성 단계에 들어갈 때마다 Codex catalog snapshot을 한 번 잡는다. 표, 미리보기, 생성은 마지막으로 들어간 구성 단계의 snapshot을 쓴다. SPEC-MODELUX-002 INV-MUX2-08("run당 1회")은 뒤로로 다시 들어간 init run에서 "들어갈 때마다 1회"가 된다. 표·미리보기 = 생성(INV-MUX2-04)은 그대로다.

### REQ-TUX-14 구성 단계 안의 뒤로
Priority: Must · EARS: EventDriven

WHEN the operator goes back inside the composition step, THEN THE SYSTEM SHALL move as the composition rows of the Navigation Contract say: Esc on 적용할까요? and on the override question acts as 계속 편집, Esc on 기본 프리셋 고르기 returns to the menu without a write, and Esc on any form of the OMP tab or the review tab returns to the tool tab with nothing written and no warning line.

- OMP 탭(native와 autopus 모두)의 실패 wrapper `ompFailed`(`quality_compose_omp.go:130-138`)는 지금 `errComposeInterrupted`만 넘기고 나머지를 경고로 바꾼다. native의 `ompPins` 오류(L93)와 autopus 탭의 폼 오류(`quality_compose_omp_autopus.go:43-93`)가 이 wrapper를 지난다. `ErrNavBack`도 넘기도록 같은 줄의 조건만 넓힌다.
- 리뷰 탭 코드는 바꾸지 않는다. 리뷰 탭의 경고 자리 두 곳은 폼 오류가 아니다(후보 config 오류 `quality_compose_review_check.go:27-31`, 쓰기 오류 `quality_compose_review.go:228-241`). 폼 오류는 `composeSelect`에서 그대로 올라간다.

### REQ-TUX-15 이동은 쓰는 값을 바꾸지 않는다
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL write, after any sequence of back and forward moves, project files whose normalized digest equals the one the same final answers write without those moves.

- normalized digest는 `acceptance.md` 머리의 NPD다(manifest `generated_at` 값과 timestamp backup 경로만 정규화). 그 밖의 bytes는 모두 같아야 한다.

### REQ-TUX-16 test seam
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL let tests drive every navigation form through the accessible helper over scripted input: composeRunForm takes the form's NavFrame, and InitWizardOpts carries a form runner and a resume result, whose nil values keep the terminal defaults.

- `composeRunForm func(cmd *cobra.Command, form *huh.Form, frame tui.NavFrame) error`의 기본값은 `tui.RunNav(form, frame, cmd.InOrStdin(), cmd.OutOrStdout())`다. `runComposeForm`이 `frame.Accessible = tui.NavAccessible()`를 넣는다. `frame.Accessible`이 true면 `RunNav`는 accessible 경로다. 그래서 accessible form을 기본 runner로 돌리던 기존 test(`TestComposeRunForm_DefaultBindsTheCommandStreams`)는 `tui.NavFrame{Accessible: true}`를 넘기면 지금과 같은 출력과 값을 얻는다.
- wizard의 `[NEW] InitWizardOpts.Runner`가 nil이면 `tui.RunNav(form, frame, os.Stdin, out)`이고 `out`은 accessible이면 os.Stdout, 아니면 os.Stderr다(지금 huh 기본 stream과 같다).
- `[NEW] InitWizardOpts.Resume`가 nil이 아니면 wizard는 그 답을 state로 마지막 단계에서 시작하고, 앞 단계를 모두 답한 것으로 요약에 보인다. `[NEW] tui.WizardAnswerLines(opts, result, lang)`는 `buildStepList(opts)`가 묻는 단계의 답 줄을 만든다(확인 화면의 답한 값).
- 새 test helper `[NEW] useNavComposeForms`는 새 test 파일에 두고 `tui.RunNavAccessible(form, frame, cmd.InOrStdin(), cmd.OutOrStdout())`를 runner로 건다. 기존 helper(`useComposeFormRunner`)는 frame을 쓰지 않는 지금 runner 그대로다.

### REQ-TUX-17 locale
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL print every frame and review text in Korean when the language rule of its surface yields ko and in English otherwise: the wizard and its REA step use DetectSystemLang as reaStepTextFor does, and the composition step and the init review use the session language as composeFormTextFor does.

### REQ-TUX-18 문서
Priority: Should · EARS: EventDriven

WHEN this SPEC ships, THEN THE SYSTEM SHALL describe Esc back, the accessible back token, the first-step notice, the frame, and the init review in the init, update, and quality help texts, README, and the v0.50.126 CHANGELOG.

### REQ-TUX-19 실제 터미널 smoke
Priority: Must · EARS: EventDriven

WHEN the implementation is complete, THEN THE SYSTEM SHALL have one recorded real-terminal smoke of auto init and auto update that exercises Esc back, the first-step notice, the frame, the init review's 뒤로 and 취소, and Ctrl+C, stored as evidence/manual-smoke-tuiux.txt with local paths replaced.

### REQ-TUX-20 터미널 높이
Priority: Must · EARS: Ubiquitous

THE SYSTEM SHALL fit every TUI navigation form into the terminal height by the Height Contract, keeping the header line and the footer line visible and shortening only the summary block.

## Contracts

### Form Gate Contract

폼이 열리는 조건은 지금 그대로다. 이 SPEC은 열리는 폼을 navigation 폼으로 바꿀 뿐이다.

| 자리 | 폼이 열리는 조건(기존) | CI |
|---|---|---|
| init wizard | `--yes`가 없고 `composeTTYProbe()` true(`init.go:106`), `FormsSupported`(`wizard_steps.go:54`) | 보지 않는다 |
| init 구성 단계 | wizard가 돈 run에서 품질 답이 역할별 직접 구성이고 `composeFormsProbe(cmd)` true. 아니면 줄 driver나 읽기 전용 표 | 보지 않는다 |
| init 확인 화면 | wizard가 돈 run에서 `composeFormsProbe(cmd)` true | 보지 않는다 |
| update 구성 단계 | `composeGate`가 묻는 행(`--yes`·`-y`·`--auto`·`--check`·`--workspace`·TTY probe false·CI truthy면 묻지 않는다, `quality_compose_driver_gate.go:64-70`)이고 `composeFormsProbe(cmd)` true | truthy면 묻지 않는다 |
| quality 구성 단계 | `qualityComposeInteractive`(아니면 기존 TTY 오류, `quality_compose_entry.go:83-86`)이고 `composeFormsProbe(cmd)` true | 보지 않는다 |

- 폼이 열리는 자리에서 `TERM=dumb`이면 huh처럼 accessible 경로다(REQ-TUX-06).

### Navigation Contract

| 입력 방식 | 뒤로 | 앞으로 | 취소 |
|---|---|---|---|
| TUI 폼(`RunNav`) | Esc. Select filter 입력 중이면 filter 입력을 끝낸다(huh 그대로) | Enter. 같은 단계 안의 다음·이전 field는 huh 그대로(Enter·Tab / shift+tab) | Ctrl+C → `huh.ErrUserAborted`(기존 의미) |
| accessible(`TERM=dumb`, test runner) | `b`·`back`·`뒤로` 한 줄 | 기존 답(번호, y/n, 빈 줄 = 기본값) | 폼 단위 중단 없음(huh 그대로). 프로세스 SIGINT는 지금처럼 프로세스를 끝낸다 |
| 줄 driver, 폼이 열리지 않는 자리 | 없음 | 그대로 | 그대로 |

| 표면 | 폼 | 뒤로 가면 | state |
|---|---|---|---|
| init wizard | 단계 k (k ≥ 2) | 단계 k-1 | k가 보이기 직전 값으로 되돌린다. k-1과 그 뒤 답은 남아 다시 보일 때 미리 선택된다 |
| init wizard | 단계 1 | 머문다, 안내 | 그대로 |
| init 구성 단계 | 메뉴 | wizard 마지막 단계 | 구성 단계에 들어가기 직전 cfg로 되돌린다(메모리 구성은 버린다) |
| init 확인 | 뒤로 또는 Esc | 구성 단계가 폼을 물었으면 그 메뉴, 아니면 wizard 마지막 단계 | 메뉴로 가면 cfg 그대로, wizard로 가면 위 행 |
| update·quality | 메뉴 | 머문다, 안내 | 그대로 |
| 구성 단계 | 도구 탭 | 메뉴 | 쓰지 않는다 |
| 구성 단계 | 방식 | 도구 탭. 도구 탭을 묻지 않았으면(선택지 하나, 갱신 뒤 계속 편집) 메뉴 | 쓰지 않는다 |
| 구성 단계 | 역할 그룹의 등급·effort | 이번 편집 pass에서 바로 앞에 물은 폼. pass의 첫 폼이면 방식 | 지금 폼이 보이기 직전 값(편집 map과 effort working 사본)으로 되돌린다. 방식으로 나가면 이번 pass는 버린다 |
| 구성 단계 | 에이전트별 목록·값·effort | 바로 앞에 물은 폼(값 → 목록, effort → 값, 두 번째 이후 목록 → 앞 에이전트의 마지막 폼). pass의 첫 목록이면 방식 | 위와 같다 |
| 구성 단계 | override 질문(모든 도구 같게) | 방식. 편집 결과는 남는다 | `계속 편집`과 같다 |
| 구성 단계 | `적용할까요?`(갱신 경로 포함) | `계속 편집`과 같다 | 기존 동작 |
| 구성 단계 | 기본 프리셋 고르기 | 메뉴 | 쓰지 않는다 |
| 구성 단계 | OMP 탭·멀티프로바이더 리뷰 탭의 모든 폼 | 도구 탭 | 그 탭의 이번 pass는 쓰지 않았고 버린다. 경고 줄 없음 |

- update·quality에서 `적용`으로 끝난 쓰기는 뒤로가 되돌리지 않는다(기존 규칙 "nothing undoes an earlier confirmed write").
- 안내 줄: ko `첫 단계입니다. 취소하려면 Ctrl+C를 누르세요.` / en `This is the first step; press Ctrl+C to cancel.`

### Summary Contract

| 자리 | 보이는 값 | 형식 |
|---|---|---|
| wizard 단계 k | 이번 run에서 k 앞에 답한 단계. flag나 기존 설정으로 건너뛴 단계는 없다 | 머리 줄 + `  <단계 이름>: <값>` |
| 구성 편집 pass(역할 그룹, 에이전트별) | 이번 pass에서 지금 폼 앞에 답한 등급·effort·값 폼. 목록 고르기는 줄이 아니다 | 머리 줄 + `  <주제>: <답>` |
| 메뉴 | 기존 구성 표, 프리셋 상태 줄, 리뷰 요약 줄 | 기존 그대로 |
| 도구 탭·방식 | 없음. header 경로가 앞 선택이다 | - |
| `적용할까요?`, OMP·리뷰 탭 | 기존 미리보기(after-table, 바뀌는 설정 줄, OMP route 미리보기, 리뷰 미리보기) | 기존 그대로 |
| init 확인 | Init Review Contract | 폼 앞에 평문 블록 |

- 머리 줄: ko `지금까지 고른 값` / en `Chosen so far`. 줄이 없으면 블록도 없다.
- wizard 값: Select는 고른 선택지 라벨의 첫 ` — ` 앞(`Fullstack`, `Ultra`, `역할별 직접 구성`, `Compose by role`, `TDD`, `None`). 언어 설정은 ko `주석 <c> · 커밋 <m> · AI 응답 <a>` / en `comments <c> · commits <m> · AI responses <a>`(값은 en·ko·ja·zh 코드). Confirm은 ko `켬`·`끔` / en `on`·`off`.
- 단계 이름: ko `사용 프로필`, `언어 설정`, `품질 모드`, `리뷰 게이트`, `REA 역공학 MCP`, `방법론` / en `Usage Profile`, `Language Settings`, `Quality Mode`, `Review Gate`, `REA reverse-engineering MCP`, `Methodology`(en은 지금 제목에서 `[n/N] `를 뗀 글자, REA는 `reaStepTextFor(lang).Title`).
- 구성 주제: 그룹 라벨(`composeRoleGroups`의 ko·en), 에이전트 이름, effort 폼은 `<주제> effort`(`composeEffortTitle`). 답은 저장 값(`fable`·`opus`·`sonnet`·`haiku`, effort 단계 이름) 또는 `그대로`/`keep`(`composeFormText.Keep`).

### Init Review Contract

확인 블록은 구성 단계의 표처럼 폼 앞에 평문으로 쓴다(TUI에서도 scrollback에 남아 Height Contract를 받지 않는다). 그 뒤 frame(header `init · 확인` / `init · Review`, 요약 없음, footer)과 Select가 온다.

| 묶음 | ko / en 머리 줄 | 줄 |
|---|---|---|
| 전체 | `적용할 설정` / `Settings to apply` | - |
| 답한 값 | ` 답한 값` / ` Answers` | `tui.WizardAnswerLines`: 이번 run의 `buildStepList(opts)` 단계마다 `  <단계 이름>: <값>`(Summary Contract의 wizard 값). 물은 단계가 없으면 묶음이 없다 |
| 바뀌는 설정 | ` 바뀌는 설정` / ` Changed settings` | 아래 diff 규칙. 바뀐 key가 없으면 `  (없음)` / `  (none)` |
| REA | ` REA` / ` REA` | REA 대상 플랫폼이 있을 때(`reaRun.active`) 한 줄 `  REA MCP: <결정> · <동작>`. 결정은 이 run의 REA 결정(`reaRun.state`, `rea_run.go:112-127`) ko `켬`·`끔` / en `on`·`off`. 동작은 `decide`가 정한 `cfg.Runtime.ReaMCP`: Add ko `등록` / en `add`, Keep `그대로` / `keep`, Remove `제거` / `remove` |

- diff 규칙: B는 wizard 답을 적용하기 직전의 cfg 사본(REQ-TUX-13의 기준 사본), F는 확인 화면 시점의 cfg다. 둘을 `config.Save`와 같은 struct tag로 YAML marshal 한다(`yaml:"-"` field, 예 `Runtime`은 없다). mapping을 leaf까지 펴고 list는 flow(`[a, b]`) 한 leaf로 본다. 더해졌거나 값이 다르거나 없어진 leaf가 바뀐 leaf다. 바뀐 leaf를 부모 경로로 묶어 부모마다 한 줄 `  <부모 경로>: <key>=<값>, …`(key 순)으로 쓰고, 최상위 leaf는 `  <key>: <값>`이다. 없어진 leaf의 값은 ko `(없음)` / en `(none)`이다. 줄은 경로 순이고 값은 `terminalSafe`를 거친다.
- 그래서 구성 단계의 프리셋 등급과 effort(`quality.presets.<p>.agents`·`.efforts`), binding(`quality.providers`), OMP `role_model_policy`, 리뷰 탭이 바꾼 key(`spec.review_gate` 등)가 모두 key 이름으로 보인다. 값이 같으면 줄이 없다(예: 리뷰 게이트를 끄면 `spec.review_gate: enabled=false`만 보이고 기본 provider 목록은 그대로라 줄이 없다).
- 확인 화면에 없는 것: 확인 뒤의 정규화(`MigrateOrchestraConfig`), statusLine·parent rule 줄 질문과 그 결과, 생성 뒤 `reaRun.finish`가 정하는 REA key(`reaKeyAfterRun`), 생성 파일. 확인 블록 머리 줄 아래에 이 설명을 두지 않고 help 문서(REQ-TUX-18)에 적는다.
- Select 제목 ko `이대로 진행할까요?` / en `Proceed with these settings?`, 선택지 ko `적용`·`뒤로`·`취소` / en `apply`·`back`·`cancel`, 기본값 적용. 제목은 구성 단계의 `적용할까요?`와 다르게 읽히도록 `적용할까요?`로 끝나지 않는다(구현 중 개정 A-1, `evidence/review-rounds.md`). 취소는 기존 줄 `  init cancelled.`(`init.go:120`과 같은 bytes)를 쓴다.
- 줄 하나만 묻고 바로 쓰는 `기본 프리셋 고르기`는 확인 화면 대상이 아니다(non-goal). update·quality의 구성 쓰기는 기존 `적용할까요?`가 확인 화면이다.

### Stage Indicator Contract

| 표면 | header 형식 | ko 예 | en 예 |
|---|---|---|---|
| wizard 단계 | `init · <n>/<N> <단계 이름>` | `init · 4/6 리뷰 게이트` | `init · 4/6 Review Gate` |
| 구성 메뉴 | `<명령> · 모델 구성` | `update · 모델 구성` | `update · Model composition` |
| 구성 하위 | `<명령> · 모델 구성 › <메뉴 선택>[ › <탭>][ › <방식>]` | `update · 모델 구성 › 구성 편집 › Codex › 역할 그룹` | `update · Model composition › edit the composition › Codex › role groups` |
| init 확인 | `init · 확인` | `init · 확인` | `init · Review` |

- `<명령>`은 `cmd.Name()`(init, update, quality)이다. 구분자는 ` · `(U+00B7)와 ` › `(U+203A), 앞뒤 공백 하나씩이다.
- 메뉴 선택은 `MenuEdit`·`MenuRefresh`·`MenuPreset`, 탭은 `TabAll`·`Claude Code`·`Codex`·`OMP`·리뷰 탭 라벨(`멀티프로바이더 리뷰`/`Multi-provider review`), 방식은 `MethodGroups`·`MethodAgents` 문구다(`quality_compose_form_text.go`, `quality_compose_review_text.go`). 갱신 뒤 계속 편집은 `… › 최신 기본값으로 갱신 › 모든 도구 같게 › 역할 그룹`이다.
- 도구 탭 폼의 경로는 `… › 구성 편집`, 방식 폼은 `… › <탭>`, 편집·override·`적용할까요?` 폼은 `… › <탭> › <방식>`, OMP·리뷰 탭 폼은 `… › <탭>`이다.

| footer | ko | en |
|---|---|---|
| TUI | `방향키 이동 · Enter 선택 · Esc 뒤로 · Ctrl+C 취소` | `Arrows move · Enter select · Esc back · Ctrl+C cancel` |
| TUI 첫 폼 | `방향키 이동 · Enter 선택 · Ctrl+C 취소` | `Arrows move · Enter select · Ctrl+C cancel` |
| accessible | `답을 입력하고 Enter를 누르세요 · 뒤로: b` | `Type an answer and press Enter · back: b` |
| accessible 첫 폼 | `답을 입력하고 Enter를 누르세요` | `Type an answer and press Enter` |

- TUI 배치: header 한 줄, 요약 블록, huh 폼, footer가 마지막 줄. huh help 줄은 숨긴다(`WithShowHelp(false)`). 첫 폼에서 Esc를 누르면 다음 키까지 footer 자리에 안내 줄이 보인다. 폼이 끝나면 frame도 사라진다(huh View가 끝나면 빈 문자열인 것과 같다). header·footer는 `MutedStyle`이다.
- accessible 배치: header(빈 문자열이면 없음), 요약 블록, footer를 ANSI 없이 한 줄씩 쓴 뒤 field를 돈다. 안내는 token 줄 다음에 한 줄이다. accessible은 줄을 줄이지 않는다.
- 요약 블록의 머리 줄은 호출자가 `NavFrame.Summary`의 첫 줄로 준다(wizard·편집 pass는 `지금까지 고른 값`). helper는 받은 줄을 그대로 쓴다.

### Height Contract

- H는 마지막 `tea.WindowSizeMsg`의 높이, 받기 전에는 24다. `RunNav`는 form에 높이 H−2의 `WindowSizeMsg`를 넘겨 form이 header·footer를 뺀 높이 안에서 그리게 한다(huh가 넘치는 선택지를 scroll 한다).
- Select는 첫 key 전의 모든 frame에서 기본값의 커서 줄을 보인다. form이 필요한 줄을 받으면 첫 선택지부터 모두, 받지 못하면 huh가 커서 줄로 scroll 한 줄들을 보인다. huh는 선택지를 넣는 순간의 field 값으로 scroll 위치를 정하므로 구성 단계의 Select는 `tui.NavSelect`로 만든다(v0.50.126 RC smoke: `true`를 먼저 둔 bool Select가 커서 줄 없이 둘째 선택지와 빈 줄로 열렸다).
- h는 그린 form의 줄 수(`lipgloss.Height(form.View())`), avail = H − 2 − h다. 요약 블록(머리 줄 1 + 답 n줄)이 avail 안이면 다 보인다. 아니고 avail ≥ 3이면 머리 줄, `  … 외 <k>개`(en `  … <k> more`), 마지막 avail − 2개 답을 보이고 k = n − (avail − 2)다. avail < 3이면 요약 블록을 보이지 않는다.
- header와 footer(또는 그 자리의 안내)는 늘 보인다. 그래서 그린 줄 수는 H 이하다. bubbletea v1.3.6이 높이를 넘는 위쪽 줄을 버리는 일(`standard_renderer.go:182-187`)이 생기지 않는다.

## Existing Test Changes

기존 test의 단언 의미는 바꾸지 않는다. 바뀌는 것은 아래 분류뿐이고, T1이 파일별 분류를 `research.md` `### 기존 test 분류`에 붙인다.

1. seam 시그니처, 다섯 곳: `composeRunForm`에 값을 넣는 4곳(`quality_compose_form_helpers_test.go:44`, `quality_compose_form_failure_test.go:29`, `quality_compose_line_golden_test.go:25`, `quality_compose_review_gate_test.go:28`)은 `tui.NavFrame` 인자 하나를 더한다. 기본 runner를 직접 부르는 1곳(`quality_compose_form_test.go:94`, `TestComposeRunForm_DefaultBindsTheCommandStreams`)은 `require.NoError(t, composeRunForm(cmd, form, tui.NavFrame{Accessible: true}))`가 된다. 이 줄은 `require.` 줄이지만 호출 인자만 늘어난 시그니처 변경으로 허용하고, 그 test의 기대 값(`value == "b"`, `2. b`)은 그대로다. 이 test가 기대는 accessible 실행, 숫자 입력, accessible 출력은 `Accessible`이 true인 `RunNav`가 지금 `form.Run()`과 같게 준다(REQ-TUX-16).
   - 다섯 파일은 모두 `package cli`이고 지금 `internal/cli/tui`를 import하지 않는다. 그래서 파일마다 `github.com/insajin/autopus-adk/internal/cli/tui` import 한 줄과 gofmt가 그 import block에 내는 정렬 변경을 허용한다. 그 밖의 줄은 바꾸지 않는다.
2. init 확인 화면: form path로 init을 돌리는 test(`init_compose_form_test.go`, `init_compose_review_test.go`, `init_compose_omp_test.go`, form path라면 `init_rea_test.go`)는 마지막에 폼 하나를 더 만난다. (a) 입력이 옛 흐름 끝에서 끝나면 EOF가 적용을 고른다: 변경 없음. (b) 남는 줄이 빈 줄뿐이면 Enter가 적용이다: 변경 없음. (c) 남는 줄에 빈 줄이 아닌 답이 있으면 그 줄이 확인 화면에 닿기 전에 끝나도록 입력 모양만 고친다. init test에는 폼 실행 횟수를 단언하는 줄이 없다(grep `Equal(t, <n>, *` 0건). 있게 되면 회귀로 보고 멈춘다.
3. 새 helper(`useNavComposeForms`)와 새 test는 새 파일에만 둔다(`[NEW] quality_compose_nav_helpers_test.go` 등). 기존 test 파일에 함수를 더하지 않는다.
4. wizard test(`internal/cli/tui/*_test.go`)는 step builder와 제목이 그대로라 기존 줄을 바꾸지 않는다. T5가 소유한 `export_test.go`에 test alias 두 줄(`TestWizardStreams`·`TestWizardStepKinds`)만 더한다(S17 (iv)). `RunInitWizard(Accessible: true)`는 지금처럼 폼 없이 기본값이다.
5. 줄 driver golden(`quality_compose_line_golden_test.go`, `quality_compose_table_golden_test.go`)은 stub 시그니처 줄 말고는 그대로 통과해야 한다.
6. 위 분류 밖에서 실패하면 회귀다. test가 아니라 코드를 고친다.
7. SPEC-REVMODEL-001 개정이 같은 release에서 `composeRunForm` stub을 더하면 새 시그니처를 써야 한다. compile 오류가 신호이고 나중에 들어오는 쪽이 맞춘다.

## 생성 파일 상세

- `[NEW] internal/cli/tui/nav.go`(+test): `NavFrame{Header, Summary, First, Lang, Accessible}`, `ErrNavBack`, `NavAccessible`, `RunNav`(accessible이면 `RunNavAccessible`로, 아니면 wrapper tea model: frame 그리기, Height Contract, Esc·filter·첫 폼 처리, Ctrl+C·완료 판정, help 숨김), frame 문구 ko·en.
- `[NEW] internal/cli/tui/nav_accessible.go`(+test): `RunNavAccessible`, 줄 단위 token reader(1 byte씩 읽어 다음 줄을 남긴다).
- `[NEW] internal/cli/tui/nav_select.go`(+test): `NavSelect`(첫 frame은 첫 선택지부터, 커서는 기본값). `composeSelect`가 쓴다(v0.50.126 RC smoke).
- `[NEW] internal/cli/tui/nav_history.go`(+test): `NavHistory[S]`(폼마다 보이기 직전 state와 답 라벨, 뒤로 pop, 요약 줄).
- `[NEW] internal/cli/tui/wizard_nav.go`(+test): wizard 단계 loop(history, 단계 이름, 요약 값, resume, mode별 stream), `WizardAnswerLines`. `wizard_steps.go`(275줄)는 loop를 이 파일로 옮기고 `InitWizardOpts`에 `Runner`·`Resume` 두 줄만 더한다.
- `[NEW] internal/cli/quality_compose_nav.go`(+test): 구성 단계 위치 상태(경로, 첫 폼, 요약), command context 싣기·읽기, header 문자열.
- `[NEW] internal/cli/quality_compose_form_preset.go`: `choosePreset`을 옮긴다(`quality_compose_form_driver.go`가 296줄이라 뒤로 처리를 더할 자리를 만든다). 동작 불변.
- `[NEW] internal/cli/init_interactive.go`(+test): init 대화형 loop(기준 cfg 사본, wizard → 구성 단계 → 확인 화면, resume). `init.go`의 L106-147은 이 함수 호출로 바뀌고, L98-101의 기존 프로필 읽기는 대화형 경로에서 `config.LoadPreview`다.
- `[NEW] internal/cli/init_review.go`(+test): 확인 블록 묶음, 문구, Select. `[NEW] internal/cli/init_review_diff.go`(+test): Init Review Contract의 YAML diff 규칙.
- 기존 파일 변경: `quality_compose_form.go`(seam 시그니처, frame), `quality_compose_form_driver.go`(menu·edit·chooseTab·compose·review의 뒤로), `quality_compose_form_edit.go`·`quality_compose_form_effort.go`(history와 요약), `quality_compose_omp.go`(`ompFailed` 조건 한 줄), `init_compose.go`(`[NEW] runInitComposeStage`: done·back·cancelled, `[NEW] composeFormOptions.BackOut`), `init.go`(호출 교체, `LoadPreview`), `wizard_steps.go`, test 5곳(분류 1).
- 문서(Should): `init_help.go`, `update_help.go`, `quality_help.go`, `README.md`, `CHANGELOG.md`(v0.50.126). 증거: `evidence/probe-nav-a1.md`(있음), `evidence/review-rounds.md`(있음), `[NEW] evidence/base-goldens/`(T1), `[NEW] evidence/manual-smoke-tuiux.txt`(S18).
- 모든 신규·변경 소스 파일은 300줄 이하다. 새 의존성은 없다(YAML marshal은 config가 이미 쓰는 YAML 라이브러리다).

## Related SPECs

- SPEC-MODELUX-002(implemented): 폼 흐름, `composeRunForm`·`composeFormsProbe` seam, Input Mode Contract, Ctrl+C를 그대로 쓴다. 바뀌는 점은 셋이다. (1) `composeRunForm`이 `tui.NavFrame`을 받는다. (2) 폼마다 frame이 보이고 Esc가 뒤로다. (3) INV-MUX2-08은 뒤로로 다시 들어간 init run에서 구성 단계 진입마다 1회다(REQ-TUX-13).
- SPEC-MODELUX-003: effort Select는 역할 그룹·에이전트별 history의 한 폼이다. OMP native·Autopus 탭은 탭 단위 뒤로만 받는다.
- SPEC-REVMODEL-001(같은 release, 병렬 개정 중): 리뷰 탭의 의미는 바꾸지 않는다. 공유 layer(frame, 탭 단위 뒤로)만 리뷰 탭에 닿는다. 확인 화면은 리뷰 탭이 바꾼 key를 diff로 보인다.
- SPEC-REA-001: REA 단계는 wizard 단계 하나로 뒤로·요약·header를 받는다. 확인 화면의 REA 줄은 `reaRun`의 결정과 동작을 읽기만 한다.
- SPEC-INITUX-001: wizard 테마, `FormsSupported` guard, `SummaryTable`을 그대로 쓴다. SPEC-MODELUX-001: 줄 driver는 그대로다.
- Sibling: 없다(`research.md` `## Sibling SPEC Decision`).

## Traceability Matrix

| Requirement | Plan Task | Acceptance Scenario | Semantic Invariant |
|-------------|-----------|---------------------|--------------------|
| REQ-TUX-01 | T2, T3, T4, T6 | S4, S9, S21 | INV-TUX-07 |
| REQ-TUX-02 | T2, T5, T7, T8 | S1, S2, S4, S10, S11 | INV-TUX-01, INV-TUX-07 |
| REQ-TUX-03 | T5, T8 | S1, S5, S10, S11 | INV-TUX-01 |
| REQ-TUX-04 | T2, T3, T5, T7 | S3, S14 | INV-TUX-07 |
| REQ-TUX-05 | T2, T9 | S4, S8, S16 | INV-TUX-05 |
| REQ-TUX-06 | T2, T3 | S1, S2, S3, S21 | INV-TUX-08 |
| REQ-TUX-07 | T1, T10 | S15 | INV-TUX-06 |
| REQ-TUX-08 | T5, T6 | S9 | INV-TUX-03 |
| REQ-TUX-09 | T2, T3 | S4, S9 | INV-TUX-03 |
| REQ-TUX-10 | T4, T5, T8 | S1, S5, S10, S11 | INV-TUX-02 |
| REQ-TUX-11 | T9 | S6, S7 | INV-TUX-04, INV-TUX-11 |
| REQ-TUX-12 | T9 | S8 | INV-TUX-05 |
| REQ-TUX-13 | T9 | S7, S14 | INV-TUX-04, INV-TUX-12 |
| REQ-TUX-14 | T7 | S12, S13 | INV-TUX-09 |
| REQ-TUX-15 | T7, T8, T9 | S6, S7, S10, S11, S12, S14 | INV-TUX-04 |
| REQ-TUX-16 | T5, T6 | S1, S17, S21 | INV-TUX-06 |
| REQ-TUX-17 | T2, T5, T6, T9 | S6, S9 | INV-TUX-03 |
| REQ-TUX-18 | T11 | S19 | - |
| REQ-TUX-19 | T12 | S18 | INV-TUX-07 |
| REQ-TUX-20 | T2 | S20 | INV-TUX-10 |
