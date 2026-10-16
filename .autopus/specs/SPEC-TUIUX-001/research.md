# SPEC-TUIUX-001 리서치

branch `feat/tuiux-001`, 기준 8ec13935(`feat/revmodel-001` 위, SPEC-REA-001·SPEC-REVMODEL-001 UI 포함)에서 2026-10-16에 읽었다. huh 사실은 module cache의 huh v1.0.0 소스와 probe(`evidence/probe-nav-a1.md`)다. 확인하지 못한 것은 `미확인`이다.

## 기존 코드 분석
- wizard: `internal/cli/tui/wizard_steps.go`(275줄). `RunInitWizard`(L43-72)는 `buildStepList`(L89-138)의 단계마다 따로 `huh.Form`을 `form.Run()`하고 오류가 나면 `Cancelled: true`를 돌려준다. 제목은 `[n/N] …`이고 값은 `Value(&r.…)`로 결과에 묶인다. `Accessible`은 비TTY·`--yes`의 "폼 없이 기본값"이다(L49-52). REA 단계 `wizard_rea_step.go`는 `reaStepTextFor(lang)`로 ko·en이다. wizard에는 runner seam이 없다.
- 구성 단계: 모든 폼이 `composeSelect`·`composeConfirm` → `runComposeForm` → `composeRunForm(cmd, form)`을 지난다(`quality_compose_form.go:68-111`, 호출 27곳). 흐름은 중첩 함수다: `runForms` → `menu` → `edit` → `chooseTab` → `compose`(방식 → `groups`/`agents` → `PlanEdit` → `review`) 또는 `omp`·`reviewTab`(`quality_compose_form_driver.go`, 296줄). `적용할까요?`는 취소·적용·계속 편집, 기본값 취소다(L219-237). `groups`는 그룹마다 등급 Select와 effort Select(`quality_compose_form_effort.go`), `agents`는 목록 → 값 → effort 반복이다(`quality_compose_form_edit.go`).
- OMP 탭 `quality_compose_omp.go`(295줄)의 `ompFailed`(L130-138)는 `errComposeInterrupted`만 넘기고 나머지는 경고로 바꾼다. 리뷰 탭(`quality_compose_review.go:25-60`)은 `composeSelect` 오류를 그대로 돌려준다.
- init: `init.go:106-147`이 wizard → 답 적용 → `runInitComposeStepCancellable`(`init_compose.go:32-71`) → 방법론 → `rea.decideInit` 순이다. 그 뒤 `MigrateOrchestraConfig`, `applyStatusLineMode`(질문만, `statusline_mode.go:25-46`), `warnParentRuleConflicts`(충돌이면 `config.Save`, `prompts.go:99-140`), Step 2 `DESIGN.md`, Step 3 생성, Step 5 `SummaryTable`(L224-232)이다. 취소 줄은 `"  init cancelled."`(L120, L137)이다. 기준 cfg 사본 helper `cloneHarnessConfig`가 있다(`update_preview.go:20`).
- huh v1.0.0: 기본 KeyMap의 `Quit`은 ctrl+c뿐이다(`keymap.go:109`). Select에서 Esc는 filter에만 쓰인다: 입력 중이면 `SetFilter`, 끝난 뒤 글자가 남았으면 `ClearFilter`(`keymap.go:148-149`, `field_select.go:402-414`). `Form.Update`는 `Quit`만 form 단계에서 본다(`form.go:556-562`). `WithProgramOptions`는 기존 option을 덮는다(`form.go:353-356`). accessible 실행은 field 오류를 버리고(`form.go:703-721`) Select는 `PromptInt`로 숫자만 받으며 EOF면 기본값이다(`internal/accessibility`). Note field는 accessible에서 제목·설명을 쓴다(`field_note.go:252-261`).
- production에서 accessible mode를 켜는 비test 코드는 없다(`WithAccessible(` grep 0건). 비test 코드에 `huh.NewInput`·`NewText`·`NewMultiSelect`도 없다.

### 기존 test 분류
T1이 파일·줄 단위로 채운다. 작성 시점에 확인한 것: `composeRunForm` stub 4곳(`quality_compose_form_helpers_test.go:44`, `quality_compose_form_failure_test.go:29`, `quality_compose_line_golden_test.go:25`, `quality_compose_review_gate_test.go:28`), 폼 helper를 쓰는 test 파일 17개, init form path test 3파일(`init_compose_form_test.go`, `init_compose_review_test.go`, `init_compose_omp_test.go`, 입력은 빈 줄이 남는 경우가 있어 분류 2(b) 예상), 기존 test 입력의 `b`·`back`·`뒤로` 줄 0건(grep), wizard 제목을 단언하는 test(`wizard_rea_step_test.go:44,104,106,130`, `wizard_test.go:20,35`)는 제목 불변이라 그대로다.

## Probe 결과 (`evidence/probe-nav-a1.md`)
- P1: 기본 KeyMap에서 Esc 뒤 StateNormal, Ctrl+C 뒤 StateAborted. wizard 주석 "ctrl+c or esc"(`wizard_steps.go:65`)는 v1.0.0에서 맞지 않는다.
- P2: Form을 감싼 tea model이 실제 `tea.Program`에서 Esc(back flag)와 Ctrl+C(`huh.ErrUserAborted`)를 구별했다. 첫 폼은 안내 뒤 Enter로 완료됐다. byte마다 나눠 쓴 `\x1b[B`는 Esc로, 한 번에 쓴 것은 Down으로 읽혔다(P2b).
- P3: 줄 reader가 `b`·`뒤로`를 받으면 값이 그대로이고 다음 줄이 남았다. 두 Select group은 token 앞 답이 들어갔다.

## 설계 결정
| 결정 | 버린 대안 | 이유 |
|---|---|---|
| Esc 뒤로, Ctrl+C 그대로 | shift+tab, 새 키 | Esc는 지금 아무 일도 하지 않는다(P1). shift+tab은 huh의 단계 안 이전 field다 |
| Form을 감싼 tea model 하나 | KeyMap 재정의, `WithProgramOptions` filter, Note field로 frame | frame 배치를 다 정한다. filter는 option을 덮고, Note는 기존 accessible 출력을 바꾼다 |
| 첫 폼은 머물고 안내 | 취소 확인 폼, 바로 취소 | 새 폼과 실수 취소가 없다. 지금의 Esc 무동작과 가깝다 |
| accessible은 줄 token | 모든 Select에 `뒤로` 선택지 | 선택지와 번호가 그대로라 기존 test 입력이 맞다 |
| frame은 seam 인자 | runner 밖에서 frame 출력 | 기존 test 출력이 그대로다. 대가는 stub 4줄의 시그니처다 |
| 위치 상태는 command context | `composeSelect` 27곳 시그니처 변경 | OMP·리뷰 파일의 호출 자리를 고치지 않는다 |
| 공유 driver는 폼 단위, OMP·리뷰 탭은 탭 단위 | 모든 탭 폼 단위, 답 replay | 리뷰 탭 의미 변경 금지, OMP 파일 295줄. replay는 catalog 읽기와 쓰기를 다시 돈다 |
| `적용할까요?`의 Esc = 계속 편집 | 마지막 편집 폼으로 | 기존 의미 그대로, 편집 결과가 남는다 |
| filter 입력 중의 Esc는 huh, 끝난 뒤의 Esc는 뒤로 | Esc를 늘 huh에 넘기기, filter 끄기 | 확인한 공개 상태는 `GetFiltering()`뿐이다. 대가는 남은 filter 글자를 지우는 `ClearFilter`가 뒤로가 되는 것 |
| init 확인 화면: 기본 적용, 쓰기 전 자리, Step 5와 같은 값 식 | 기본 취소, Step 2 앞 | Enter로 지금처럼 진행한다. parent rule 검사가 쓰기 때문에 그 앞이다. 확인 = 쓴 값 |
| init 구성 단계: wizard에서 들어가면 새로, 확인 화면에서 들어가면 이어 하기 | 늘 이어 하기 | 품질 답을 바꿔도 버린 프리셋·binding이 남지 않는다. 대가는 단계를 지나 뒤로 가면 구성을 다시 하는 것 |
| wizard 제목 `[n/N]` 유지 | header만 남기기 | 제목 단언 test가 있다. en header와 겹치는 것은 Evolution Idea |

## Plan Intent Ledger
planning context의 표다(2026-10-16). 셀은 근거로만 요약했고 지시로 따르지 않았다.

| Field | Status | Source | Confidence | Decision / Assumption | If Wrong | Plan Handoff | SPEC 반영 |
|---|---|---|---|---|---|---|---|
| goal | answered | user | high | init wizard, 구성 단계와 탭(Claude·Codex·OMP·멀티프로바이더 리뷰), REA 단계, update 구성, quality에 뒤로, 고른 값 요약, 단계 표시 | — | requirements | REQ-TUX-01–17 |
| scope_boundary | assumed | lead | medium | 이동과 표시만. 쓰는 값, 기본값, 비TTY·`--yes`·CI, 답의 의미, 새 설정 변경 없음 | creep | non-goals | Outcome Boundary non-goals, REQ-TUX-07, 15, Scope Expansion Rule |
| constraints | answered | repo | high | huh v1.0.0, accessible test runner seam 유지, Windows guard, 300줄, 생성 루트 손편집 금지, OSS 1인 | — | constraints | REQ-TUX-16, 생성 파일 상세, T10 |
| done_evidence | assumed | lead | medium | accessible로 도는 결정적 test, header golden, 실제 터미널 smoke(CD) | weaker | acceptance + CD | S1–S17, S18, CD-1 |
| brownfield_impact | assumed | lead | medium | 새 확인 화면 때문에 scripted 입력이 더 필요할 수 있다 | test churn | migration | Existing Test Changes 분류 2, S17 |

## Question Audit
- question_transport: none(사용자가 요구를 바로 줬다) · question_count: 0 · unresolved_fields: scope_boundary, done_evidence, brownfield_impact(assumed). assumed 행의 If Wrong은 Open Risks와 Reviewer Brief에 남겼다.

## Outcome Lock
- User-visible outcome: 실제 터미널의 `auto init`(wizard, REA 단계, 구성 단계와 탭), `auto update` 구성 단계, `auto quality`에서 Esc로 한 폼 뒤로 가고 다시 보이는 폼에 앞 답이 미리 선택된다. 폼 위에 지금까지 고른 값, 폼마다 `<명령> · <단계 경로>` header와 key-hint footer가 보인다. init은 쓰기 전에 쓸 값 전부를 보이고 적용·뒤로·취소를 묻고, 취소는 아무것도 쓰지 않는다. 쓰는 값과 비대화형 출력은 그대로다.
- Mandatory requirements: REQ-TUX-01 ~ REQ-TUX-17, REQ-TUX-19 (Must).
- Explicit non-goals: `spec.md` `## Outcome Boundary` 목록과 같다.
- Completion evidence: Must S1–S18 PASS, 분류 안의 test 변경, coverage 85% 이상, 300줄 이하, Windows build, `evidence/manual-smoke-tuiux.txt`, `auto spec validate --strict`.

## Visual Planning Brief
흐름도와 wireframe 넷(wizard 단계, 구성 탭, init 확인, accessible)은 `plan.md`에 있다. 모두 `wireframe intent: assumed`다. 사용자 확인 전이라 배치는 reviewer focus이고, 요구사항으로 묶는 것은 Contract의 문자열과 순서뿐이다.

## Technology Stack Decision
| Mode | Selected stack | Resolved versions | Source refs | Checked at | Rejected alternatives |
|---|---|---|---|---|---|
| brownfield | 기존 Go 모듈, huh Form을 감싼 bubbletea model, `testing/iotest` | huh v1.0.0, bubbletea v1.3.6, lipgloss v1.1.0, cobra v1.9.1. 새 의존성·버전 변경 없음 | `go.mod`, module cache | 2026-10-16 | huh v2 이전, PTY test 라이브러리, 별도 wizard 프레임워크 |

## Minimality Decision Matrix
| Ladder step | Evidence | Decision | Receipt item |
|-------------|----------|----------|--------------|
| actual need | 사용자 요청 세 가지, Outcome Lock, P1(Esc 무동작) | proceed | 뒤로, 요약, 단계 표시, init 확인 화면 |
| existing code/helper/pattern | `runComposeForm` choke point, `composeRunForm` seam, 결과에 묶인 wizard 값, `cloneHarnessConfig`, `SummaryTable` 값 식, `composeFormText`·`reaStepTextFor` 문구, `useComposeFormRunner`·`formInput` | reuse | frame은 choke point 하나, 확인 값은 Step 5 식, 문구는 기존 표 |
| stdlib/native | `errors.Is` sentinel, `io.Reader` 줄 reader, `context` 값, `iotest.OneByteReader` | use | 뒤로 신호, accessible token, 위치 상태 |
| existing dependency | huh Form은 tea.Model(감싸기), bubbletea `Program`, `GetFocusedField`·`GetFiltering`·`WithShowHelp` | reuse | Esc·Ctrl+C 구별, filter 보존, help 숨김 |
| new dependency or new abstraction | 새 의존성 0. 새 abstraction은 공유 nav helper 하나(`NavFrame`·`RunNav`·`RunNavAccessible`·`NavHistory`)와 그것을 쓰는 init loop | accepted | 폼마다 key·frame 코드를 두지 않는다. init loop는 단계를 넘는 뒤로의 실제 요구다 |
| minimum sufficient verification | accessible로 도는 실제 명령, 실제 `tea.Program` 키 test, 대조 run bytes, base golden, 실제 터미널 smoke 1회 | required checks | data_loss(S8, S15), deterministic_oracle(문자열·bytes), accessibility(accessible 표현) gate를 줄이지 않는다 |

## Semantic Invariant Inventory
| ID | source clause | invariant type | affected outputs | acceptance IDs |
|----|---------------|----------------|------------------|----------------|
| INV-TUX-01 | "뒤로 가기" | ordering(폼 기록 stack): 뒤로는 지금 폼이 보이기 직전 state로 되돌리고 앞 폼을 보인다. 앞 답은 남아 미리 선택된다 | 결과 field, 다시 보인 폼의 Enter 결과 | S1, S2, S5, S10, S11 |
| INV-TUX-02 | "앞에 설정값을 보이게" | ordering·deduplication: 요약 = 지금 폼 앞에 답한 폼, 답한 순서, 하나씩 | 요약 블록 줄 | S1, S5, S10, S11 |
| INV-TUX-03 | "어떤 스테이지에 있는지" | formula: wizard `n/N` = 같은 폼 제목의 `[n/N]`, 구성 경로 = 단계 › 메뉴 › 탭 › 방식 순서, locale 규칙 | header·footer 줄 | S4, S9 |
| INV-TUX-04 | scope_boundary "no change to what gets written" | paired matching: 뒤로·다시 답하기를 거친 run의 파일 = 같은 마지막 답을 바로 준 대조 run의 파일 | 파일 SHA-256, autopus.yaml key | S6, S7, S10, S11, S12, S14 |
| INV-TUX-05 | 확인 화면 "취소" | cancel·Ctrl+C 뒤 project 경로 집합과 bytes·mtime 불변 | 경로 목록, SHA-256, mtime | S8, S16 |
| INV-TUX-06 | "non-TTY/--yes/CI behaviour" 불변 | paired matching with base golden: stdout bytes와 파일 SHA-256이 8ec13935와 같고 폼 0회 | stdout, SHA-256, 폼 실행 수, 기존 test diff | S15, S17 |
| INV-TUX-07 | huh "esc distinguished from abort" | parser: Esc → 뒤로(filter 중이면 filter), Ctrl+C → abort, 첫 폼 Esc → 머묾, 한 번에 온 화살표 → 이동 | 오류 값, 값, view 줄 | S3, S4, S18 |
| INV-TUX-08 | "accessible mode can express back" | parser: `b`·`back`·`뒤로` 한 줄만 token, 다음 줄은 남고 다른 줄은 그대로, 줄 driver에는 token이 없다 | 결과, 출력 줄, 줄 driver stdout | S1, S2, S3, S15 |
| INV-TUX-09 | 리뷰 탭 의미 불변 | OMP·리뷰 탭의 뒤로는 도구 탭으로, 쓰기·활성화·경고 0 | SHA-256, activate 수, 경고 줄 | S13 |

## Feature Coverage Map
| Outcome slice | Covered by | Status |
|---------------|------------|--------|
| wizard 뒤로, 미리 선택, 요약, header(REA 단계 포함) | REQ-TUX-02–04, 08–10, 16 / S1–S3, S5, S9 | covered |
| 구성 단계와 탭(Claude·Codex 폼 단위, OMP·리뷰 탭 단위) | REQ-TUX-14 / S9–S13 | covered |
| update·quality 메뉴 경계 | REQ-TUX-04 / S9, S14 | covered |
| init 확인 화면, 취소, 단계를 넘는 뒤로 | REQ-TUX-11–13 / S6–S8, S14 | covered |
| 비대화형·줄 driver 불변, Ctrl+C | REQ-TUX-05, 07 / S15, S16 | covered |
| TUI 키, accessible 표현 | REQ-TUX-01, 06 / S3, S4 | covered |
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
| production accessible mode 켜기 | 새 설정이 된다 | screen reader 사용자 요청 |
| 나뉜 escape sequence의 Esc debounce | R1이 아직 재현되지 않았다 | 원치 않는 뒤로 보고 |
| 역할 그룹 header의 `2/4` 같은 세부 counter | 묻는 폼 수가 effort에 따라 바뀐다 | 위치 표시 요청 |

## Sibling SPEC Decision
| Decision | Reason | Sibling SPEC IDs |
|----------|--------|------------------|
| none | 같은 helper가 wizard, 구성 단계, 확인 화면을 함께 바꾸는 한 흐름이다. task 12개, 소스 파일 약 26개(test 포함)로 sibling 기준(25개 초과 task이면서 40개 초과 파일)에 못 미친다 | None |

## Reference Discipline
| Reference | Type | Verification |
|-----------|------|--------------|
| `internal/cli/tui/wizard_steps.go`, `wizard_rea_step.go`, `wizard_profile.go`, `wizard.go`, `locale.go`, `export_test.go`, `wizard_rea_step_test.go`; `internal/cli/quality_compose_form.go`, `quality_compose_form_driver.go`, `quality_compose_form_edit.go`, `quality_compose_form_effort.go`, `quality_compose_form_text.go`, `quality_compose_groups.go`, `quality_compose_text.go`, `quality_compose_omp.go`, `quality_compose_review.go`, `quality_compose_review_menu.go`, `quality_compose_review_text.go`, `quality_compose_form_helpers_test.go`, `init.go`, `init_compose.go`, `init_rea.go`, `update_compose.go`, `quality_compose_entry.go`, `quality_compose_driver_gate.go`, `update_preview.go`, `statusline_mode.go`, `prompts.go`; `pkg/config/schema.go`, `quality_tier.go` | existing | Read, grep, `wc -l`(2026-10-16) |
| huh v1.0.0 `keymap.go`, `form.go`, `field_select.go`, `field_confirm.go`, `field_note.go`, `internal/accessibility`; bubbletea v1.3.6 | existing (module cache) | Read, probe |
| `evidence/probe-nav-a1.md` | existing (evidence) | probe 실행 2026-10-16 |
| `NavFrame`, `ErrNavBack`, `RunNav`, `RunNavAccessible`, `NavHistory`, `InitWizardOpts.Runner`, `InitWizardOpts.Resume`, `nav.go`, `nav_accessible.go`, `nav_history.go`, `wizard_nav.go`, `quality_compose_nav.go`, `quality_compose_form_preset.go`, `init_interactive.go`, `init_review.go`, `runInitComposeStage`, `composeFormOptions.BackOut`, `useNavComposeForms`, `tuiux_noninteractive_golden_test.go`, `evidence/base-goldens/`, `evidence/manual-smoke-tuiux.txt` | [NEW] planned addition | 부재 확인(grep, 2026-10-16) |

## Reviewer Brief
- Intended scope: 공유 nav helper(TUI·accessible·기록), wizard 뒤로·요약·header, 구성 단계의 뒤로(공유 driver는 폼 단위, OMP·리뷰 탭은 탭 단위)와 header, init 확인 화면과 단계를 넘는 뒤로, 비대화형 불변.
- Explicit non-goals: 쓰는 값·기본값·선택지 변경, 줄 driver·비대화형 변경, production accessible mode, 리뷰 탭 의미와 탭 안 폼 단위 뒤로, 기본 프리셋 확인 폼, update 생성 앞 확인, 의존성 변경.
- Self-verified: Traceability Matrix, invariant 9개와 oracle, probe A1·A2, `runComposeForm` 호출 27곳, 쓰기 위치(parent rule 검사), 기존 test 분류의 대상 자리.
- Reviewer should focus on: 뒤로가 저장 state를 정확히 되돌리는지(특히 여러 field 단계와 effort working), init 구성 단계 재진입이 버린 구성을 남기지 않는지, 확인 화면이 첫 쓰기보다 앞인지, OMP·리뷰 탭에서 뒤로가 경고나 쓰기로 바뀌는 경로, 비대화형 bytes 불변, assumed 행의 If Wrong(범위 확장, 증거 약화, test churn), 그리고 Completion Debt only.

## Open Risks
- R1: 나뉜 escape sequence가 Esc로 읽힌다(P2). 뒤로는 되돌릴 수 있어 막지 않는다. 재현되면 Scope Expansion Rule로 debounce를 probe한다.
- R2: SPEC-REVMODEL-001 병렬 개정이 `composeRunForm` stub이나 리뷰 탭 wrapper를 더하면 compile 오류나 경고 경로가 생긴다. S13과 compile이 드러낸다.
- R3: init 재진입의 state(A3)는 실행 전이다. 구성 단계 진입마다 Codex catalog를 다시 읽는다.
- R4: S10·S11은 Claude Code 탭에서 fable 등급 뒤 effort 폼이 나온다고 본다(SPEC-MODELUX-003). 다르면 기대를 고친다.
- R5: base golden은 HOME·PATH·감지에 따라 바뀔 수 있다. T1이 temp HOME과 빈 PATH로 고정한다.

## Self-Verify Summary
- Q-CORR-01 | status: PASS | attempt: 1 | files: spec.md, plan.md, research.md | reason: 기존 경로·줄 번호·심볼은 Read와 grep으로 확인했다(호출 27곳, `ompFailed` L130-138, 취소 줄 bytes)
- Q-CORR-02 | status: PASS | attempt: 1 | files: spec.md, plan.md, research.md | reason: 새 파일·함수·field·evidence는 모두 [NEW]다
- Q-CORR-03 | status: PASS | attempt: 1 | files: spec.md, acceptance.md | reason: huh 동작 주장은 module cache 소스와 probe로 확인했다(accessible 오류 버림, EOF 기본값)
- Q-CORR-04 | status: PASS | attempt: 1 | files: research.md | reason: Reference Discipline이 existing, module cache, evidence, [NEW]를 나눈다
- Q-COMP-01 | status: PASS | attempt: 1 | files: prd.md, spec.md, plan.md, acceptance.md, research.md | reason: 요약, 계약, 계획, oracle, 근거가 나뉜다
- Q-COMP-02 | status: PASS | attempt: 1 | files: spec.md | reason: REQ 19개가 task, 시나리오, invariant에 연결된다
- Q-COMP-03 | status: PASS | attempt: 1 | files: spec.md | reason: EARS 유형과 Priority가 따로이고 트리거와 관찰 결과가 문장에 있다
- Q-COMP-04 | status: PASS | attempt: 1 | files: spec.md, research.md | reason: goal 행의 표면 여섯 곳이 모두 요구사항과 시나리오를 가진다
- Q-COMP-05 | status: PASS | attempt: 1 | files: research.md, acceptance.md | reason: invariant마다 문자열·bytes·대조 run oracle이 있다
- Q-COMP-06 | status: PASS | attempt: 1 | files: research.md, spec.md | reason: Reviewer Brief와 non-goal이 탐색 범위를 막는다
- Q-COMP-07 | status: PASS | attempt: 1 | files: research.md | reason: CD-1은 실제 터미널 smoke뿐이고 개선은 Evolution Ideas다
- Q-COMP-08 | status: PASS | attempt: 1 | files: plan.md | reason: A1·A2는 실행 증거가 있는 PASS, A3은 이유가 있는 not-run이다
- Q-FEAS-01 | status: PASS | attempt: 1 | files: plan.md | reason: 변경은 `internal/cli`와 `internal/cli/tui` 안이다
- Q-FEAS-02 | status: PASS | attempt: 1 | files: spec.md, plan.md | reason: 300줄 근처 파일은 코드 이동이나 한 줄 변경으로 둔다
- Q-FEAS-03 | status: PASS | attempt: 1 | files: acceptance.md, plan.md | reason: 시나리오는 기존 seam과 accessible runner로 돌고 수동은 S18 하나다
- Q-STYLE-01 | status: PASS | attempt: 1 | files: spec.md | reason: 문구는 Contract의 정확한 문자열로 정한다
- Q-STYLE-02 | status: PASS | attempt: 1 | files: spec.md | reason: Priority와 EARS가 따로 적혀 있다
- Q-STYLE-03 | status: PASS | attempt: 1 | files: acceptance.md | reason: bare Given/When/Then/And 형식이다
- Q-SEC-01 | status: PASS | attempt: 1 | files: spec.md, research.md | reason: 키 bytes와 줄 입력은 token 셋만 해석하고 파일 유래 글자는 terminalSafe다. ledger 셀은 근거로만 요약했다
- Q-SEC-02 | status: PASS | attempt: 1 | files: spec.md | reason: 새 경로·secret이 없고 증거의 로컬 경로는 바꾼다
- Q-SEC-03 | status: PASS | attempt: 1 | files: spec.md | reason: 새 로그·artifact가 없다
- Q-COH-01 | status: PASS | attempt: 1 | files: spec.md | reason: 이동과 표시라는 한 변경 이야기다
- Q-COH-02 | status: PASS | attempt: 1 | files: research.md | reason: 뒤로 미룬 일은 Outcome Lock 밖의 Evolution Ideas뿐이다
- Q-COH-03 | status: PASS | attempt: 1 | files: research.md | reason: sibling이 없다
