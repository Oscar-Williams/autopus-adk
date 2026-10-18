# SPEC-TUIUX-001 review rounds

## Round 1 (2026-10-16, verdict REVISE)

- Providers: claude는 429로 빠졌고 codex와 gemini가 판정했다. finding 12건(major 10, minor 2), checklist FAIL 7개(Q-CORR-03, Q-COMP-02, Q-COMP-03, Q-COMP-04, Q-COMP-05, Q-COMP-08, Q-FEAS-03, 모두 codex).
- 모든 finding을 코드로 다시 확인한 뒤 revision 2에 반영했다. 확인에 쓴 자리는 표의 "근거"다.

| ID | 판정 | 근거(다시 확인) | 반영 |
|---|---|---|---|
| F-001 | 맞다 | `init.go:99` `config.Load(dir)`, `pkg/config/loader.go:87-94`의 `os.WriteFile`, `migrate.go:185`(`gemini` → `antigravity-cli`). init 경로의 다른 `config.Load`는 없다(grep) | REQ-TUX-11·12: 대화형 경로의 프로필 읽기는 `config.LoadPreview`, 비대화형은 그대로. S8 (b)에 옛 이름 `gemini` fixture. plan T9, Scope Expansion Rule |
| F-002 | 맞다 | `rea_run.go:112-127`(`decide`는 `cfg.Runtime.ReaMCP`만), `rea_decision.go:153`(`reaKeyAfterRun`), `runtime.go:15`(`yaml:"-"`), `defaults.go:75-90`(provider 셋), `init.go:131-134` | Init Review Contract: REA 줄은 `reaRun` 결정·동작, 리뷰 게이트는 답한 값과 diff(`spec.review_gate: enabled=false`). S6 기대 값을 다시 계산했다 |
| F-003 | 맞다 | 고정 7행은 프리셋 tier·effort, OMP policy, 리뷰 탭 key를 보이지 못한다 | Init Review Contract: 답한 값, 기준 사본과의 YAML leaf diff(부모마다 한 줄), REA. S6 정확한 줄, S7 (b)에 구성 key 줄 |
| F-004 | 맞다 | `pkg/adapter/manifest.go:55`, `:97`(`generated_at` = `time.Now()`), `:179`(backup timestamp), `init.go:52`(project 이름 = 디렉터리 이름) | oracle을 NPD(시간 field만 정규화)로 바꾸고 fixture 이름을 `p`로 고정. T1이 두 번 돌려 다른 field를 찾는다. REQ-TUX-15, acceptance 머리, S6·S7·S10–S15 |
| F-005 | 맞다 | `quality_compose_driver_gate.go:64-70`(update만 CI), `quality_compose_entry.go:83-86`(quality는 CI를 보지 않음), `init.go:106` | REQ-TUX-07을 폼 단위로 다시 쓰고 Form Gate Contract를 더했다. S15 (f)(g)(h) |
| F-006 | 맞다 | huh `form.go:124-127`(`TERM=dumb` → `WithAccessible(true)`), `:331`, `:672`(accessible 출력 기본 os.Stdout) | `NavFrame.Accessible`, `NavAccessible`, `RunNav`가 accessible이면 그 경로, wizard stream은 mode별. REQ-TUX-06·16, plan 분류표 정정, S21 |
| F-007 | 맞다 | `quality_compose_form_test.go:87-98`이 `composeRunForm(cmd, form)`을 직접 부르고 accessible·숫자 입력·`2. b`에 기댄다 | 분류 1을 다섯 곳으로, 그 줄은 `NavFrame{Accessible: true}`를 넘겨 단언 그대로. 새 helper는 새 파일. 폼 횟수 단언 0건이라 2(d) 삭제. S17 허용 목록 |
| F-008 | 맞다 | `quality_compose_omp.go:70-93`: 첫 폼 기본값 `돌아가기`, 첫 폼·family 오류는 바로 반환, `ompPins` 오류만 `ompFailed`(L93) | S13 (b): O3N(native)에서 에이전트별 지정 → 모델 계열 → 첫 model 폼 `b`. autopus 탭도 `ompFailed`를 지난다(`quality_compose_omp_autopus.go:43-93`) |
| F-009 | 맞다 | bubbletea `standard_renderer.go:182-187`(높이를 넘으면 위쪽 줄을 버린다) | REQ-TUX-20, Height Contract(H−2를 form에, 요약은 마지막 답부터 `외 k개`, avail < 3이면 숨김), 확인 블록은 평문. S20 |
| F-010 | 맞다 | S10·S11은 effort를 Enter로만 답했고 재진입 catalog가 같았다 | S10 (a)(b)(c): effort `low` 저장·복원·버림·pass 버리기. S11 (a) effort `low`와 목록 뒤로. S7 (b)·S14 (c): 첫 진입 CAT-A, 둘째 진입 CAT-B, `gpt-6.1-sol` 부재와 CAT-B 대조. A3 입력 |
| F-011 | 맞다 | house style과 파서는 StateDriven을 WHERE … THEN으로 읽는다(SPEC-MODELUX-002 REQ-MUX2-01) | REQ-TUX-10을 WHERE … THEN으로 고쳤다 |
| F-012 | F-007과 같다 | 같은 직접 호출 | F-007과 함께 닫았다 |

- 리뷰 탭은 다시 확인했다. `f.warn` 두 곳(`quality_compose_review_check.go:27-31`, `quality_compose_review.go:228-241`)은 폼 오류가 아니라 리뷰 탭 코드는 바꾸지 않는다(REQ-TUX-14).
- checklist FAIL 7개는 위 반영으로 닫았고 `research.md` `## Self-Verify Summary`에 attempt 2로 적었다.

## Round 2 (2026-10-16, verdict REVISE)

- Providers: claude는 429로 빠졌고 codex와 gemini가 판정했다. 12건 중 9건 resolved, major 3건(F-007, F-008, F-010)과 checklist FAIL 3개(Q-COMP-02, Q-COMP-05, Q-FEAS-03)가 남았다. revision 3에서 닫았다.

| ID | 판정 | 근거(다시 확인) | 반영 |
|---|---|---|---|
| F-007 | 맞다 | 다섯 test 파일(`quality_compose_form_helpers_test.go`, `_failure_test.go`, `_line_golden_test.go`, `_review_gate_test.go`, `quality_compose_form_test.go`)은 모두 `package cli`이고 `internal/cli/tui`를 import하지 않는다(grep) | 분류 1과 S17이 `quality_compose_form_test.go:94`의 `require.NoError(t, composeRunForm(cmd, form, tui.NavFrame{Accessible: true}))`(호출 인자만, 기대 값 그대로)와 파일마다 tui import 한 줄·gofmt 정렬을 명시적으로 허용한다. plan T6, Existing Test Changes, research 분류 |
| F-008 | 맞다 | `quality_compose_omp.go:161-200`: `ompPins`는 기본값 `완료`인 `에이전트를 고르세요` 목록을 먼저 묻고 에이전트를 고른 뒤에만 `ompPin`(`:208-222`, 제목 `<agent> (<capability>)`)을 부른다. native 에이전트는 scout·reviewer·security-reviewer·task·sonic(`omp_native_agents.go:6-7`) | S13 (b): `OMP 구성` 4번 `에이전트별 지정 (balanced 위에)` → `모델 계열` Enter → 목록 `task — 프로필 기본값` → `task (` model 폼 `b`. (b2): 목록에서 `b`. 두 경우 `ompFailed`(L93) 경로를 지나고, 조건이 넓혀지지 않으면 경고가 나와 도구 탭이 다시 안 나오므로 구별된다 |
| F-010 | 맞다 | balanced 설계·판단 넷은 `fable`(`role_model_policy_balanced.go:24-31`)이라 S14 (c)의 `fable`은 NoChange였다. S12 (a)(b)도 같은 문제였다(이번에 함께 찾았다) | S7 (b) 보존: Codex `opus` + effort `low` 적용 → 확인 뒤로 → 메뉴 유지 → `low` 유지, 대조 K. S14 (c) 폐기: Claude effort `low` 적용 → 메뉴 뒤로 → wizard → 새 진입에서 Codex `opus` 적용 → `claude-custom`·`efforts` 없음, 대조 X. snapshot은 둘 다 첫 진입 CAT-A·둘째 CAT-B이고 CAT-A만 쓴 음성 대조가 `gpt-6.1-sol`을 보인다. A3 oracle을 (1) 보존 (2) 폐기 (3) snapshot으로 나눴다. S12는 `opus`로 바꿨다 |

- checklist FAIL 3개는 위 반영으로 닫았고 `research.md` Self-Verify Summary에 attempt 3으로 적었다.

## 구현 중 개정 (2026-10-16, T9–T10, SPEC re-review 대상)

구현이 approved revision 3의 문장과 부딪힌 다섯 자리다. A-1은 lead 결정(안 B)이고, A-2–A-5는 `acceptance.md` Oracle Acceptance Notes의 "코드 사실과 다르면 test가 아니라 이 SPEC의 기대를 고친다" 규칙으로 고쳤다. 다섯 모두 코드가 아니라 SPEC 문장이 틀렸음을 아래 근거로 확인했고, 그래서 코드를 고치지 않았다.

| ID | 자리 | 옛 문장 | 새 문장 | SPEC이 틀린 이유(근거) |
|---|---|---|---|---|
| A-1 | `spec.md` Init Review Contract(L259), `acceptance.md` S6, `plan.md` 확인 wireframe | Select 제목 ko `이대로 적용할까요?` / en `Apply these settings?` | ko `이대로 진행할까요?` / en `Proceed with these settings?`(선택지 `적용`·`뒤로`·`취소`, 기본값 적용은 그대로) | 확인 화면 제목이 구성 단계와 OMP 탭의 확인 `적용할까요?`로 끝나 두 질문이 같은 말로 읽혔다. 같은 이유로 기존 단언 `init_compose_omp_test.go:105` `NotContains(out, "적용할까요?")`(뜻: blocker 뒤 OMP 확인을 묻지 않는다)가 실패했고, Existing Test Changes 규칙 6은 test가 아니라 코드를 고치라 한다. lead 결정(안 B)으로 제목만 바꿨고 그 test는 고치지 않은 채 통과한다 |
| A-2 | `acceptance.md` S11 (c) | 출력에 `quality.default =`가 없고 | 출력에 프리셋 쓰기 줄(`quality.default = <프리셋>` 한 줄, `choosePreset`)이 없고 … 표의 상태 줄은 이 판정에 들지 않는다 | 메뉴의 표마다 상태 줄 `quality.default = balanced   ultra: …`이 그려진다(`quality_compose_driver.go:76`, `composeDriverStatus` `quality_compose_driver_gate.go:120-121`, base golden `quality-line-b.stdout` 3번째 줄). SPEC-MODELUX-001부터의 출력이라 바꾸면 기존 golden과 REQ-TUX-07이 깨진다. 판정하려던 것은 프리셋 쓰기 줄(`quality_compose_form_preset.go:43`)이다 |
| A-3 | `acceptance.md` S14 (c) | 첫 진입의 적용 뒤 다시 그린 표에는 Claude Code `*` 칸이 있지만 둘째 진입의 첫 표에는 `*` 칸이 없다 | 다시 그린 표는 Claude Code 열이 `Claude Code (개별 설정: claude-custom)`이고 설계·판단 칸이 `fable · low`지만, 둘째 진입의 첫 표에는 `claude-custom`이 없고 그 칸이 `fable-5-1 · max`다 | `*`는 미리보기 표만 붙인다: 메뉴 표는 `composeSessionTable.Table`(`quality_compose_session.go:143-146`, 기준 표 nil)로, 미리보기만 `Preview`(`:150-154`)가 기준 표를 넘기고 `mark`가 바뀐 칸에 `*`를 단다(`quality_compose_table_render.go:40-45`). 메뉴 표에 `*`를 그리는 것은 이 SPEC 밖의 출력 변경이다. 새 문장도 같은 것(첫 진입의 적용은 표에 보이고 둘째 진입은 깨끗하게 시작한다)을 판정한다 |
| A-4 | `acceptance.md` S15 (f) | (f)는 `CI` 없이 같은 입력을 준 run과 출력이 같고 | … 다른 줄은 리뷰 탭 자신의 CI 규칙이 지우는 리뷰 요약 줄 `멀티프로바이더 리뷰: …` 하나뿐이다 | D는 리뷰 게이트가 켜져 있고 orchestra provider가 셋이라 리뷰 탭이 열린다. 리뷰 탭은 CI truthy에서 숨고 그 요약 줄도 없다(`quality_compose_review_menu.go:28`, SPEC-REVMODEL-001 REQ-RVM-14 "The CI rule is this tab's own"). 이 SPEC은 리뷰 탭 의미를 바꾸지 않으므로(non-goal) 코드가 아니라 기대가 틀렸다. quality gate가 CI를 보지 않는다는 판정(header와 메뉴가 같다)은 그대로다 |
| A-5 | `acceptance.md` S21 | `RunNav(form, NavFrame{Header: "h", Accessible: true}, in, out)` … out에는 `답을 입력하고 Enter를 누르세요 · 뒤로: b` | `NavFrame{Header: "h", Lang: "ko", Accessible: true}` | REQ-TUX-17은 frame 글을 언어 규칙이 ko일 때만 한국어로, 아니면 영어로 쓰라 한다. `Lang`이 비면 영어 footer `Type an answer and press Enter · back: b`가 맞고(`tui/nav.go:74-79`, `composeFormTextFor`와 같은 규칙 `quality_compose_form_text.go:82-87`), 한국어 footer를 기대하려면 frame에 `Lang: "ko"`가 있어야 한다. 빈 `Lang`을 한국어로 읽게 코드를 바꾸면 REQ-TUX-17을 어긴다 |
