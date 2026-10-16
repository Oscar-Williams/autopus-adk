# SPEC-REVMODEL-001 review rounds

## Round 1 (2026-10-15, REVISE: critical 1, major 13, minor 6, suggestion 1)

검증: 모든 finding을 이 branch 코드(ee811e2e)로 다시 확인했다. J-001은 `loadHarnessConfigForFlags` 경유 throwaway probe(`research.md` `## Probe 기록` 2차)로 재현했다. 설계를 바꾼 곳은 참여 범위, writer, 확인 요청 엔진, 값 모델이다.

| Finding | Severity | 판정 | 반영 |
|---|---|---|---|
| J-001 Migration 3이 명령 목록을 되돌림 | critical | 맞음(probe A2 재실행) | 참여 = `spec.review_gate.providers`만, 명령 목록·`orchestra.judge` 쓰지 않음, 미리보기 안내, G1-1(REQ-RVM-03, S1, S10, S11) |
| J-002 backend stderr가 terminalSafe 밖 | major | 맞음(`omp_review_backend.go:52, 72-73`) | REQ-RVM-20(두 줄 escape), 확인 요청 전 이름 검사, CLI stderr는 T6 확인(S16) |
| J-003 요청 1회 ≠ 호출 1회 | major | 맞음(`ompReviewMaxAttempts = 3`) | 라벨·Trust Model·S15 예산을 최대 3n회로(Live Check 요청 수 행) |
| J-004 편집 없이 확인 요청 불가 | major | 맞음 | `변경 없음` 뒤에도 확인 요청 Select, 확인 없이 메뉴로(REQ-RVM-10, 11, S5(2)) |
| J-005 S15 예산 모순 | major | 맞음 | 최대 18회, `auto spec review` 제거, doctor readiness(호출 없음)로 확인(S15) |
| J-006 전체 문서 재encode | major | 맞음(`encodeAutopusConfigRewrite`) | Line Splice Writer Contract, `pruneRetiredConfig`, `verifyReviewLineEdit`, 손으로 쓴 RH fixture(S12) |
| J-007 공유 anchor | major | 맞음 | anchor·alias가 span에 있으면 거부, 대상 밖 data 동일 검사(S12(c)) |
| J-008 빈 thinking 배열 | major | 맞음(`catalog_invalid`) | CATG 수정, CATE 분리, 단계 없는 model 경로 삭제(REQ-RVM-06, 07, S4(c)) |
| J-009 thinking 정렬과 기본값 | major | 맞음(`sort.Strings`) | OMP Autopus 탭 규칙 그대로, S2 순서·기본값 수정 |
| J-010 그대로·CATX와 검증 충돌 | major | 맞음 | 바뀐 model만 대조, catalog 없으면 model Select 없음, 오류 뒤 baseline(REQ-RVM-05, S4(a)(c), S3(c)) |
| J-011 0명 참여 경로 | major | 맞음 | 수 검사를 참여 직후로(S3(d)) |
| J-012 시작 값 이중 의미 | major | 맞음 | baseline·working 분리(REQ-RVM-02, INV-RVM-09, S17: 다시 편집 뒤 적용, 끄기→켜기 원복) |
| J-013 빈 명령 목록 fallback | minor | 맞음 | 명령 목록을 쓰지 않으므로 해당 없음(D2) |
| J-014 judge 생략과 `""` | major | 맞음 | `orchestra.judge`는 쓰지 않음, review gate judge는 raw key 유무를 baseline에(S9(b)(c)) |
| J-015 entry point별 CI | major | 맞음 | 탭 단위 CI 규칙, entry point별 gate 표(REQ-RVM-01, 14, S6(d)) |
| J-016 readiness 필터 | minor | 맞음 | 통과의 뜻을 Live Check 대상 행에 적음 |
| J-017 조립 실패 표시 | minor | 맞음 | `runProviderTransportSmoke` 전체 재사용, `review_gate` 결과 줄 |
| J-018 judge 없는 리뷰 | minor | 맞음(`NoJudge`) | `그대로 — judge 없음`(REQ-RVM-04, S9(a)) |
| J-019 CLI 안내와 참여 | minor | 맞음 | 참여하는 CLI provider만(REQ-RVM-08, S8) |
| J-020 꺼진 gate | minor | 맞음 | `enabled`일 때만 탭, init은 대입을 compose 앞으로(S10(c)) |
| J-021 300줄 여유 | suggestion | 맞음 | driver에는 호출 두 줄, 넘으면 새 파일로(T5) |

checklist FAIL(Q-COMP-02/04/05/08, Q-CORR-03, Q-FEAS-01/03, Q-STYLE-01, Q-SEC-01/03)은 위 반영으로 닫았고 `research.md` `## Self-Verify Summary`에 attempt 2로 남겼다.

## Round 2 (2026-10-16, 20 resolved, major 2 open)

| Finding | Severity | 판정 | 반영 |
|---|---|---|---|
| J-010 S4(b)가 baseline 값을 "새 model"로 줌 | major | 맞음(R3 gemini baseline = `gemini-3.1-pro:high`) | 음성 사례를 baseline과 다르고 catalog에 없는 `google-antigravity/gemini-3-pro:high`로 바꾸고, baseline 값 대조군(오류 없음)을 더함. INV-RVM-03에 "baseline과 같은 값은 통과"와 planner 오류 범위를 적음 |
| F-001 baseline 참여가 2개 미만이면 탈출 불가 | major | 맞음(진입은 막지 않고 오류 뒤 reset만 있었음) | 진입 허용, 첫 참여 Select 앞 경고 줄, 오류 뒤 Select `다시 고르기`(기본)·`나가기`, 보이는 provider만 셈(REQ-RVM-05, Review Tab Contract, Form Text, S3(c)(d), 새 S18) |

## Gate 1 (2026-10-16, 사용자 결정)

- G1-1 확정: 참여 끄기는 `auto spec review`(`spec.review_gate.providers`)에만 적용한다. 미리보기의 명령 안내 줄을 유지한다(`spec.md` Open Issues resolved).
- G1-2 확정: model 편집은 OMP backend provider만 한다. CLI backend provider에는 탭이 이유 줄을 보인다(Form Text CLI 안내, S8).
- 이 repo의 gemini 리뷰 model은 `google-antigravity/gemini-3.8-flash:high`로 바뀐다(live: `gemini-3.1-pro:high` 404, `gemini-3-pro` stream 끊김, `3.8-flash` ok). research 증거와 S15 목표 값으로만 쓰고 fixture는 기대지 않는다.

## Gate 1 개정 (2026-10-16, 실제 터미널 설정 뒤 사용자 결정)

- G1-2 superseded: 새 `auto init`은 claude·codex·gemini를 CLI backend로 만들어 "OMP backend만"이면 세 provider 모두 이유 줄만 보였다.
- 새 결정과 반영: claude CLI는 Claude 탭 model을 args `--model`로(REQ-RVM-21), codex CLI는 run의 Codex catalog model을 `model_policy: pinned` + `-m` args로 써서 `CodexOrchestraProfile`이 더 덮지 않게(REQ-RVM-22), model 목록을 증명할 수 없는 CLI는 OMP 설치 시 `backend: omp` + 증명된 model로 전환, 아니면 이유 줄(REQ-RVM-23). S8 개정, S19–S22 추가, T12–T14 추가, probe A1·A3 확장(행은 validator 상한 3이라 새 행 대신 두 행에 3차 증거를 더함).
- probe(`evidence/probe-cli-model.md`, model 호출 없음): argv는 `Config.Args` 그대로, CLI는 `Model`을 쓰지 않음, codex quality는 profile이 `-m gpt-6-astra`로 덮고 pinned는 쓴 `-m`을 지킴, claude `--model`은 남음, 설정된 gemini binary는 `agy`이고 `agy`는 `--model`·`models`가 있음(`gemini` CLI는 `-m`), v0.50.125는 pinned·전환 파일을 읽음.
- G1-3 open: agy는 model flag가 있지만 Autopus에 증명할 목록 해석기가 없어 OMP 전환만 준다. agy model Select는 scope expansion이다.

## Round 4 (2026-10-16, claude provider 429로 빠짐, codex·gemini 판정)

| Finding | 반영 |
|---|---|
| J-002 major: backend stderr와 `reportCodexRuntimeFallback`의 model·effort가 escape 없음 | REQ-RVM-20을 두 경로 모두로 넓힘, 미룸 없음(T13, S16 And) |
| F-002 major: `auto quality`는 `LoadPreview`라 migration 없이 읽음 | REQ-RVM-25: CLI baseline은 `MigrateOrchestraConfig` 뒤 사본, stored와 다르면 migration args에 값만 바꿔 씀(S24) |
| F-003: CLI 값의 공통 검증·미리보기 | REQ-RVM-05·09에 CLI Model Contract 포함, 미리보기 순서에 `backend`·`model`·`model_policy`·`args`(S26) |
| F-004: Codex helper가 다른 args를 정규화 | `ApplyCodexProviderProfile` 대신 `editCLIArgs`(값만 교체, 중복 `--json`도 보존, S24) |
| F-005: 빈 effort 교집합 | 그런 Codex model은 보이지 않음, planner 오류(S25) |
| F-006: non-goal·INV-RVM-06이 CLI 편집과 충돌 | Reviewer Brief·non-goal 정리, INV-RVM-06에 전환·pinned key 포함 |

## G1-3 (2026-10-16, 사용자 결정)

- `agy models` 해석기를 더한다(REQ-RVM-24, Agy Catalog Contract): 15초·64 KiB, 머리 줄 건너뜀, `<id>\t<display name>` 엄격 해석, 어긋나면 목록 없음과 이유 줄, provider family 같은 id만, 기본 이름 → 실제 있는 effort 접미사의 두 단계(쓰는 값은 늘 목록의 id). `OMP backend로 전환`은 다른 선택지로 남는다. 목록에 있다고 호출되는 것은 아니라 확인 요청을 권장한다(gemini-3.1-pro 404). 관측 sample은 S23에 그대로 쓴다.

## Round 5 (2026-10-16, REJECT, claude 429, codex·gemini 판정)

| Finding | 반영 |
|---|---|
| V-001/F-003 security: CLI 미리보기가 env가 펼쳐진 args 전체를 보임 | REQ-RVM-26: 바뀐 flag 값만 raw 글자로, 다른 args는 안 보임, 대상 값이 placeholder거나 raw와 load가 다르면 거부(S26 미리보기 줄, 새 S27: `${REVIEW_TOKEN}` 값·이름·다른 flag가 출력에 없음, `-m ${CODEX_MODEL}`은 거부·bytes 불변) |
| V-002: CLI baseline 하나로 | REQ-RVM-25: 현재 값 결정은 migration 뒤 사본, 편집은 raw `args` node(쌍 삽입, raw args 없으면 migration args 삽입, alias·anchor·merge·block·placeholder 거부). S24가 결정과 raw bytes 보존을 함께 판정 |
| V-003: Outcome·Reviewer Brief 범위 | research Outcome Lock·Reviewer Brief와 spec Outcome Boundary에 REQ-RVM-21–26, agy 해석, S15–S27 반영 |
| acceptance 200줄 | fixture와 AGY sample을 `evidence/acceptance-fixtures.md`로 옮기고 시나리오 사이 빈 줄을 줄임 |
