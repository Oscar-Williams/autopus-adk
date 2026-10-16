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
