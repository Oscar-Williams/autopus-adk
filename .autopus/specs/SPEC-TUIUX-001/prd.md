# PRD: 설정 TUI의 뒤로 가기, 고른 값 요약, 단계 표시

- **SPEC-ID**: SPEC-TUIUX-001
- **Mode**: Minimal
- **Source**: 사용자 요청(2026-10-16), planning context의 Plan Intent Ledger, 8ec13935 코드, huh v1.0.0 probe(`evidence/probe-nav-a1.md`)
- **Status**: Draft (revision 2, review round 1 반영)
- **Date**: 2026-10-16
- **Target module**: `autopus-adk`
- **Release**: v0.50.126(SPEC-REA-001, SPEC-REVMODEL-001, TEAMAUTO 작업과 함께), branch `feat/tuiux-001`

## 1. 문제

사용자 말: "TUI를 좀 더 UX적으로 사용자가 편하게 해줘. 엄청 많은 걸 설정해야 하니까 뒤로 가기라든지, 앞에 설정값을 보이게 하든지, 어떤 스테이지에 있는지… 이런 UX도 신경써."

1. 뒤로 갈 수 없다. wizard 단계는 따로 도는 폼이고, Esc는 huh v1.0.0 기본 KeyMap에서 아무 일도 하지 않으며 Ctrl+C는 전체 취소다. 한 답을 고치려면 처음부터 다시 해야 한다.
2. 앞에서 고른 값이 보이지 않는다. init은 다 쓴 뒤에야 Summary를 보인다.
3. 위치 표시는 wizard 제목의 `[n/6]`뿐이다. 구성 단계와 Claude·Codex·OMP·멀티프로바이더 리뷰 탭에는 어디쯤인지 표시가 없다.

## 2. 목표

- Esc로 한 폼 뒤로 간다(wizard 단계, 구성 단계의 메뉴·탭·방식·그룹·에이전트 폼, OMP·리뷰 탭은 도구 탭으로). 다시 보이는 폼에는 앞 답이 미리 선택된다. 첫 폼에서는 머물고 안내한다. Ctrl+C는 지금과 같다.
- 폼 위에 지금까지 고른 값이 한 줄씩 보이고, 터미널 높이를 넘으면 마지막 답부터 보이며 header는 잘리지 않는다. 구성 단계는 기존 표와 미리보기가 요약이다. init은 아무 파일도 쓰기 전에 답한 값, 그 답이 바꾸는 설정 key 전부(구성 등급·effort, OMP, 리뷰 탭 포함), REA 결정을 보이고 적용·뒤로·취소를 묻는다. 취소는 아무것도 쓰지 않는다.
- 폼마다 `init · 4/6 리뷰 게이트`, `update · 모델 구성 › 구성 편집 › Codex` 같은 header 한 줄과 `방향키 이동 · Enter 선택 · Esc 뒤로 · Ctrl+C 취소` footer가 ko·en으로 보인다.
- 이 모두를 공유 navigation helper 하나로 한다. accessible 폼(`TERM=dumb` 터미널 포함)에서는 `b` 한 줄이 뒤로다. 폼이 열리는 조건은 그대로다.

## 3. 사용자

- 실제 터미널에서 `auto init`, `auto update`, `auto quality`로 많은 설정을 고르는 사용자(주 사용자)
- scripted 입력·CI·`--yes`로 도는 사용자와 test(출력과 파일이 그대로여야 한다)
- 화면을 줄 단위로 읽는 사용자(`TERM=dumb`의 accessible 표현)

## 4. 범위 밖

쓰는 값·기본값·선택지 의미 변경과 새 설정, 폼 gate·비대화형 경로와 SPEC-MODELUX-001 줄 driver 변경, `TERM=dumb` 밖 accessible mode 추가, 리뷰 탭 의미 변경과 OMP·리뷰 탭 안의 폼 단위 뒤로, 기본 프리셋 고르기 확인 폼, update 생성 앞 확인, wizard 제목 문구 변경, 의존성 변경.

## 5. 성공 기준

- accessible runner로 돈 wizard에서 3단계 뒤로 뒤 2단계가 앞 답을 미리 선택하고 결과가 뒤로 없는 대조와 같다(S1). 여러 field 단계의 부분 답은 버려진다(S2).
- header·요약·footer 문자열이 Contract와 정확히 같다(S1, S5, S9, S10). 실제 `tea.Program`에서 Esc·Ctrl+C·filter가 구별된다(S4).
- init 확인 화면이 첫 쓰기 전에 답한 값, 바뀌는 설정 key, REA 결정을 정확한 줄로 보이고, 적용은 기준과 같은 파일을, 취소와 Ctrl+C는 옛 플랫폼 이름이 있어도 bytes·mtime 불변을 남긴다(S6, S8). 뒤로를 거친 run의 파일은 대조와 같고 재진입은 마지막 catalog를 쓴다(S7, S10–S14). 24줄 화면에서 header와 footer가 보인다(S20).
- 폼이 열리지 않는 자리의 출력과 파일이 8ec13935 golden과 같다(시간 field만 정규화, S15). `TERM=dumb`은 accessible 경로다(S21). 기존 test는 분류 안에서만 바뀐다(S17).
- 실제 터미널 smoke 기록(S18)이 남는다. 그 전에는 Completion Debt CD-1이다.
