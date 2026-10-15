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
