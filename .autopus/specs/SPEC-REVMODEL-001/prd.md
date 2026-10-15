# PRD: 멀티프로바이더 리뷰 모델 구성 탭

- **SPEC-ID**: SPEC-REVMODEL-001
- **Mode**: Minimal
- **Source**: planning context(2026-10-15)의 Plan Intent Ledger와 Question Audit, `feat/rea-001`(ee811e2e) 코드, probe A1–A3, review round 1(J-001–J-021)
- **Status**: Draft (revision 3)
- **Date**: 2026-10-15
- **Target module**: `autopus-adk`
- **Release**: v0.50.126, branch `feat/revmodel-001`

## 1. 문제

- `auto spec review`의 provider, provider별 model·thinking, judge는 `autopus.yaml`을 손으로 고쳐야만 바뀐다. init에는 Review Gate 예/아니오만 있다.
- catalog에 있는 model도 실제로는 실패할 수 있다. 2026-10-15에 `google-antigravity/gemini-3.1-pro:high`는 HTTP 404였고 그날 리뷰는 gemini 없이 돌았다. quorum이 차서 receipt에는 사유가 남지 않았다.

## 2. 목표

- `auto init`·`auto update`·`auto quality` 구성 UI에 `멀티프로바이더 리뷰` 탭을 더한다: `auto spec review` 참여 켜기·끄기, OMP backend provider의 model·thinking(같은 family의 증명된 설치 model만), 참여 provider 중 judge(지금 judge가 없으면 그대로 둘 수 있음), 적용 전 미리보기.
- 사용자가 고를 때만 provider마다 요청 1회(재시도 포함 최대 3회 model 호출)로 실제로 답하는지 보인다. 편집이 없어도 할 수 있고, 실패하면 기본값은 쓰지 않음이다.
- v0.50.125가 이미 읽는 세 key만 줄 단위로 쓴다. 비대화형·CI 실행은 탭을 내지 않는다.

## 3. 결정

| 항목 | 결정 | 이 SPEC에서 정한 세부 |
|---|---|---|
| 위치 (사용자) | 기존 구성 UI의 탭 | tool tab의 마지막 선택지, `enabled`·provider 둘 이상·CI 아님일 때(REQ-RVM-01) |
| 범위 (사용자) | 참여, model·thinking, judge | 참여는 `spec.review_gate.providers`만(G1-1), model은 OMP backend만(G1-2), judge는 `spec.review_gate.judge`만 |
| gemini 404 (사용자) | probe하고 바꾼다 | opt-in 확인 요청(`runProviderTransportSmoke`, 60초)과 실제 터미널 smoke에서 `gemini-3.8-flash:high`로 전환(REQ-RVM-10, 11, S15) |
| 출하 (사용자) | v0.50.126 | 새 key·secret·승인·의존성 없음, sibling 없음(REQ-RVM-15) |
| Gate 1 (사용자, 2026-10-16) | G1-1·G1-2 확정 | 참여 끄기는 `auto spec review`에만, 명령 안내 줄 유지. model 편집은 OMP backend만, CLI backend에는 이유 줄. 이 repo의 gemini 리뷰 model은 `google-antigravity/gemini-3.8-flash:high`로 바뀐다 |

## 4. 범위 밖

`orchestra.commands.*.providers`·`orchestra.judge`·Migration 3(명령 목록은 load마다 모든 provider로 다시 채워진다), 리뷰 정책 key, CLI backend provider의 model·args·`model_policy` 편집과 backend 전환, provider 항목 추가·삭제, `--multi`·`--providers` 의미, receipt `degraded_reasons` 규칙, doctor 변경, 확인 요청 자동 실행과 결과 저장.

## 5. 성공 기준과 위험

- 성공: gemini를 끄면 리뷰 목록에서 빠지고 실제 loader 뒤에도 그대로이며 명령 목록과 judge는 바뀌지 않는다(S1, S11). codex model을 고르면 `orchestra.providers.codex.model`이 정확히 `openai-codex/gpt-6-luna:high`다(S2). 꺼진 provider는 judge가 될 수 없고 provider가 둘보다 적으면 거부한다(S3). catalog에 없는 model은 새로 고를 수 없다(S4). 404 확인 요청 뒤 Enter는 파일을 바꾸지 않는다(S5). 비대화형·CI는 bytes를 바꾸지 않는다(S6). v0.50.125가 쓴 파일을 읽는다(S7). 손으로 쓴 파일은 span 밖 bytes가 그대로다(S12). 실제 터미널 smoke가 기록된다(S15, CD-1).
- 위험: 참여 끄기는 orchestra 명령과 `--multi`에 효과가 없다(G1-1 사용자 확정, 미리보기가 알림). CLI backend만 있는 새 init 프로젝트는 model을 고를 수 없다(G1-2). 줄 splice의 span 계산, CLI subprocess stderr, 60초 timeout은 구현 확인 전 가정이다. 탭이 더해져 tool tab Select를 건너뛰던 폼 test 일부에 Enter 하나가 더해진다.
