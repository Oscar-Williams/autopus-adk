# PRD: 멀티프로바이더 리뷰 모델 구성 탭

- **SPEC-ID**: SPEC-REVMODEL-001
- **Mode**: Minimal
- **Source**: planning context(2026-10-15)의 Plan Intent Ledger와 Question Audit, `feat/rea-001`(ee811e2e) 코드, probe A1–A3
- **Status**: Draft
- **Date**: 2026-10-15
- **Target module**: `autopus-adk`
- **Release**: v0.50.126, branch `feat/revmodel-001`

## 1. 문제

- 멀티프로바이더 리뷰의 provider, provider별 model·thinking, judge는 `autopus.yaml`을 손으로 고쳐야만 바뀐다. init에는 Review Gate 예/아니오만 있다.
- catalog에 있는 model도 실제로는 실패할 수 있다. 2026-10-15에 `google-antigravity/gemini-3.1-pro:high`는 HTTP 404였고 그날 리뷰는 gemini 없이 돌았다. quorum이 차서 receipt에는 사유가 남지 않았다.

## 2. 목표

- `auto init`·`auto update`·`auto quality` 구성 UI에 `멀티프로바이더 리뷰` 탭을 더한다: 참여 켜기·끄기, OMP backend provider의 model·thinking(같은 family의 증명된 설치 model만), 참여 provider 중 judge, 적용 전 미리보기.
- 사용자가 고를 때만 provider마다 1회 확인 호출을 돌려 실제로 답하는지 보인다. 실패하면 기본값은 쓰지 않음이다.
- v0.50.125가 이미 읽는 key만 쓴다. 비대화형 실행은 묻지 않고 쓰지 않는다.

## 3. 결정 (사용자와 확정, 다시 묻지 않음)

| 항목 | 결정 | 이 SPEC에서 정한 세부 |
|---|---|---|
| 위치 | 기존 구성 UI의 탭 | tool tab의 마지막 선택지, 기존 미리보기·확인 흐름(REQ-RVM-01, 09) |
| 범위 | 참여, model·thinking, judge | 끄기·켜기는 `spec.review_gate.providers`와 모든 `orchestra.commands.*.providers`에 delta로, model은 OMP backend만, judge는 참여 provider만(REQ-RVM-03, 04, 06, 08) |
| gemini 404 | probe하고 바꾼다 | opt-in 확인 호출(doctor smoke 경로 재사용, 60초)과 실제 터미널 smoke에서 `gemini-3.8-flash:high`로 전환(REQ-RVM-10, 11, S15) |
| 출하 | v0.50.126 | 새 key·secret·승인·의존성 없음, sibling 없음(REQ-RVM-15) |

## 4. 범위 밖

리뷰 정책 key(strategy, timeout, `enabled`, `max_revisions`, `min_providers`), 참여에서 나오지 않는 명령별 편집, CLI backend provider의 model·args·`model_policy` 편집과 backend 전환, provider 항목 추가·삭제, `--multi`·`--providers` 의미, receipt `degraded_reasons` 규칙, doctor 변경, 확인 호출 자동 실행과 결과 저장.

## 5. 성공 기준과 위험

- 성공: gemini를 끄면 리뷰 목록과 orchestra 명령 목록에서 빠지고 judge는 claude로 남는다(S1). codex model을 고르면 `orchestra.providers.codex.model`이 정확히 `openai-codex/gpt-6-luna:high`다(S2). 꺼진 provider는 judge가 될 수 없다(S3). catalog에 없는 model은 보이지 않는다(S4). 404 확인 호출 뒤 Enter는 파일을 바꾸지 않는다(S5). 비대화형 update는 bytes를 바꾸지 않는다(S6). v0.50.125가 쓴 파일을 읽는다(S7). 실제 터미널 smoke가 기록된다(S15, CD-1).
- 위험: CLI backend만 있는 새 init 프로젝트는 model을 고를 수 없다(D3, Evolution Idea). node writer의 bytes 보존과 60초 timeout은 구현 확인 전 가정이다. 탭이 더해져 tool tab Select를 건너뛰던 폼 test 일부에 Enter 하나가 더해진다.
