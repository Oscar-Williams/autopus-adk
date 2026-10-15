# SPEC-REVMODEL-001 수락 기준
## Test Scenarios

fixture는 모두 temp project 안에 만든다. 폼 test는 SPEC-MODELUX-002처럼 `composeFormsProbe`를 true로, `composeRunForm`을 huh accessible runner로 두고 입력을 `testing/iotest.OneByteReader`로 준다. "X를 고른다"는 그 라벨의 번호 입력이고 Enter는 기본값이다. OMP catalog는 `ompCLIFakeRunner`, 확인 호출은 요청을 기록하는 fake `providerSmokeBackendFactory`(FB)다. 설치 확인(`filterInstalledProviders`)은 PATH 앞에 둔 temp bin의 빈 실행 파일로 통과시킨다. 나머지는 실제 명령(`NewRootCmd`)이다. 기준 바이너리는 `git archive v0.50.125`(dd6d6ea2)로 만든 `auto`다. 실제 터미널은 S15에서만 쓴다.

- R3: `platforms: [claude-code, codex, omp]`. `orchestra.judge: claude`. `orchestra.providers`의 claude·codex·gemini는 모두 `backend: omp`, `subprocess.timeout: 900`이고 model은 차례로 `anthropic/claude-fable-5-1:max`, `openai-codex/gpt-6-astra:max`, `google-antigravity/gemini-3.1-pro:high`다. `orchestra.commands`의 brainstorm(debate)·plan(consensus)·review(debate)·secure(consensus)는 모두 `providers: [claude, codex, gemini]`다. `spec.review_gate`는 `enabled: true`, `strategy: debate`, `providers: [claude, codex, gemini]`, `judge: claude`, `max_revisions: 2`다. 이 branch의 `auto update --yes`로 한 번 맞춘 bytes를 B0라 한다.
- RC: R3 + `orchestra.judge: codex` + `orchestra.commands.brainstorm.providers: [claude, codex]` + `orchestra.commands.secure.judge: gemini`.
- RL: R3에서 `orchestra.providers`만 이 branch의 기본 항목(`DefaultClaudeProviderEntry`, `DefaultCodexProviderEntry`, gemini `binary: agy`)으로 바꾼 것. 셋 다 `backend`가 없다.
- CATG: strict catalog. provider/model과 보고 thinking은 `anthropic/claude-fable-5-1` [low, medium, high, max], `anthropic/claude-opus-5-5` [high, max], `openai-codex/gpt-6-astra` [low, medium, high, xhigh, max], `openai-codex/gpt-6-luna` [low, medium, high], `google-antigravity/gemini-3.1-pro` [low, high], `google-antigravity/gemini-3.8-flash` [low, medium, high], `example/foo-1` [].
- CATN: CATG에서 `google-antigravity/gemini-3.1-pro`를 뺀 것. CATX: `omp models`가 오류를 내는 runner.
- FB 응답: claude·codex와 `google-antigravity/gemini-3.8-flash:high`인 gemini는 output `AUTOPUS_PROVIDER_SMOKE_OK`, `google-antigravity/gemini-3.1-pro:high`인 gemini는 error `HTTP 404 Not Found (Cloud Code Assist)`. FB는 요청마다 Provider, Prompt, Timeout, `Config.Model`을 기록한다.
- "구역 밖 bytes"는 파일에서 top-level `orchestra:`와 `spec:` 구역을 잘라낸 나머지다.

### S1: gemini 끄기
Priority: Must
Given R3(B0), CATG가 있다
When `auto update`에서 구성 편집 → 멀티프로바이더 리뷰 → claude Enter → codex Enter → gemini `끄기` → claude·codex model Enter → judge Enter → 확인 호출 Enter → `적용` → 유지를 고른다
Then 미리보기의 바뀌는 설정은 이 순서의 다섯 줄뿐이다: `spec.review_gate.providers [claude, codex, gemini] → [claude, codex]`, `orchestra.commands.brainstorm.providers [claude, codex, gemini] → [claude, codex]`, 같은 모양의 plan·review·secure 줄. judge 줄은 없다
And 적용 뒤 `config.Load` 값은 `spec.review_gate.providers` = [claude, codex], 네 명령의 `providers` = [claude, codex], `spec.review_gate.judge` = claude, `orchestra.judge` = claude, `orchestra.providers.gemini.model` = `google-antigravity/gemini-3.1-pro:high`이고 strategy 넷은 그대로다
And B0와의 diff는 그 다섯 줄뿐이고 구역 밖 bytes의 SHA-256이 B0와 같으며, `resolveSpecReviewProviderNames(cfg, false)` = [claude, codex]이고 FB 호출은 0회다

### S2: model과 thinking은 정확한 selector로 쓴다
Priority: Must
Given R3(B0), CATG가 있다
When 리뷰 탭에서 참여는 모두 Enter, claude model `anthropic/claude-opus-5-5` → thinking Enter, codex model `openai-codex/gpt-6-luna` → thinking `high`, gemini model Enter, judge Enter, 확인 호출 Enter, `적용`을 고른다
Then codex model Select는 정확히 `그대로 — openai-codex/gpt-6-astra:max`, `openai-codex/gpt-6-astra`, `openai-codex/gpt-6-luna`이고, codex thinking Select는 `low`, `medium`, `high`이며 기본은 `low`다. claude thinking Select는 `high`, `max`이고 기본은 `max`다
And 바뀌는 설정은 `orchestra.providers.claude.model anthropic/claude-fable-5-1:max → anthropic/claude-opus-5-5:max`, `orchestra.providers.codex.model openai-codex/gpt-6-astra:max → openai-codex/gpt-6-luna:high` 두 줄이 이 순서로 있다
And 적용 뒤 `orchestra.providers.codex.model`은 문자열 `openai-codex/gpt-6-luna:high`, claude는 `anthropic/claude-opus-5-5:max`이고, B0와의 diff는 두 줄이며, `assembleSpecReviewProviders`의 codex `Model`도 `openai-codex/gpt-6-luna:high`다

### S3: judge와 provider 수 검증
Priority: Must
Given R3(B0), CATG가 있다
When (a) `planReviewComposition`에 참여 {claude, codex}·judge gemini와 참여 {claude}·judge claude를 준다. (b) 리뷰 탭에서 claude `끄기`, 나머지 Enter, judge Enter, 확인 호출 Enter, `적용`. (c) 새 run에서 codex `끄기`, gemini `끄기`, 나머지 Enter, 오류 뒤 모든 Select Enter, 유지
Then (a)는 `judge gemini는 리뷰에 참여하지 않는 provider라 고를 수 없습니다`와 `리뷰에는 provider가 두 개 이상 필요합니다 (지금 1개)` 오류다
And (b)의 judge Select 앞에 `judge claude가 꺼져 judge를 다시 고릅니다`가 있고 선택지는 정확히 `codex`, `gemini`이며 기본은 `codex`다. 적용 뒤 `spec.review_gate.judge` = codex, `orchestra.judge` = codex(편집 전 두 judge가 같았다), 다섯 목록 = [codex, gemini]다
And (c)는 `리뷰에는 provider가 두 개 이상 필요합니다 (지금 1개)` 뒤 참여 Select로 돌아가고, 다음 편집은 `변경 없음`이며 파일 SHA-256이 B0와 같다

### S4: catalog에 없는 model은 고를 수 없다
Priority: Must
Given R3, CATN, CATX가 있다
When CATN으로 리뷰 탭의 claude·gemini model Select를 열고, `planReviewComposition`에 CATN과 gemini 새 model `google-antigravity/gemini-3.1-pro:high`, CATG와 codex `openai-codex/gpt-6-luna` thinking `xhigh`를 주며, CATX로 리뷰 탭을 연다
Then gemini model Select는 정확히 `그대로 — google-antigravity/gemini-3.1-pro:high (catalog에 없음)`, `google-antigravity/gemini-3.8-flash`이고, claude model Select는 정확히 `그대로 — anthropic/claude-fable-5-1:max`, `anthropic/claude-fable-5-1`, `anthropic/claude-opus-5-5`이며 `example/foo-1`은 어느 Select에도 없다
And planner 오류는 `gemini model google-antigravity/gemini-3.1-pro:high는 설치된 OMP catalog가 증명하지 않습니다`와 `codex thinking xhigh는 openai-codex/gpt-6-luna가 보고하지 않는 단계입니다`다
And CATX run은 `OMP catalog를 읽지 못해 model은 그대로입니다`를 한 번 쓰고 model Select 없이 참여·judge Select를 묻는다

### S5: 확인 호출 404는 쓰지 않는다
Priority: Must
Given R3(B0), CATG, FB, temp bin의 `omp`가 있다
When (1) 리뷰 탭에서 codex model `openai-codex/gpt-6-luna` → `high`, 나머지 Enter, 확인 호출 `확인 호출 — 참여 provider 3개에 model 호출 1회씩` → 확인 Enter. (2) 구성 편집 → 리뷰 탭 → gemini model `google-antigravity/gemini-3.8-flash` → `high`, 나머지 Enter, 같은 확인 호출 → `적용`. (3) 다시 리뷰 탭에서 claude model `anthropic/claude-opus-5-5` → Enter, `확인 호출 없이 계속` → `취소`
Then (1)에서 FB는 claude, codex, gemini 순서로 3회 불리고 Prompt는 모두 `Return ONLY this exact token and no other text: AUTOPUS_PROVIDER_SMOKE_OK`, Timeout은 모두 60초, codex `Config.Model`은 `openai-codex/gpt-6-luna:high`다
And (1)의 출력은 `확인 호출 claude (anthropic/claude-fable-5-1:max): 통과`, `확인 호출 codex (openai-codex/gpt-6-luna:high): 통과`, `확인 호출 gemini (google-antigravity/gemini-3.1-pro:high): 실패 — HTTP 404 Not Found (Cloud Code Assist)`, `확인 호출이 실패한 provider가 있습니다. 적용하면 그 provider는 리뷰에서 실패합니다.`이고, Enter(기본 `취소`) 뒤 파일 SHA-256이 B0와 같다
And (2)의 gemini 줄은 `확인 호출 gemini (google-antigravity/gemini-3.8-flash:high): 통과`이고 실패 안내가 없으며, 적용 뒤 `orchestra.providers.gemini.model` = `google-antigravity/gemini-3.8-flash:high`, codex model = `openai-codex/gpt-6-astra:max`(1에서 취소)다. (3) 뒤 FB 호출 수는 6회 그대로다

### S6: 비대화형 실행은 묻지 않고 쓰지 않는다
Priority: Must
Given R3(B0), CATG, FB가 있다
When `auto update --yes`, stdin이 `strings.NewReader("\n\n\n")`이고 stdout이 buffer인 `auto update`, `auto quality --apply`를 비대화형으로 각각 실행한다
Then 세 출력 어디에도 `멀티프로바이더 리뷰`와 `리뷰 참여`가 없고, `composeRunForm` 기록에 리뷰 탭 title이 없다
And FB 호출은 0회이고 세 실행 뒤 파일 SHA-256이 모두 B0와 같다

### S7: v0.50.125가 쓴 파일을 읽는다
Priority: Must
Given S1, S2, S3(b), S5(2), S9가 남긴 파일과 대조군(S1 파일의 `spec.review_gate`에 `reviewers: [claude]`를 더한 것)이 있다
When 각 파일의 project에서 v0.50.125 바이너리로 `auto config get usage_profile`을 실행하고, 각 파일의 YAML key 경로 집합을 그 run의 시작 파일(S9는 RC, 나머지는 B0)의 집합과 비교한다
Then 다섯 파일은 exit 0이고, 대조군은 exit 1과 `field reviewers not found in type config.ReviewGateConf`다
And 다섯 파일의 key 경로는 모두 시작 파일에 있다. 빠지는 key는 S9 파일의 `orchestra.commands.secure.judge` 하나뿐이다

### S8: CLI backend provider는 model을 묻지 않는다
Priority: Must
Given RL, CATG가 있다
When 리뷰 탭에서 gemini `끄기`, 나머지 Enter, 확인 호출 Enter, `적용`을 고른다
Then 출력에 `claude는 CLI backend라 이 탭에서 model을 바꾸지 않습니다`, 같은 codex·gemini 줄이 하나씩 있고 model·thinking Select title이 없으며, after 표의 model 칸은 셋 다 `CLI backend`다
And 적용 뒤 목록 다섯 개는 gemini를 잃고 `orchestra.providers` mapping은 적용 전과 deep-equal이다(`args`, `model_policy`, `binary` 그대로)

### S9: 명령별 구성은 손댄 provider만 바뀐다
Priority: Must
Given RC, CATG가 있다
When (1) 리뷰 탭에서 gemini `끄기`, 나머지 Enter, `적용`. (2) gemini `켜기`, gemini model Enter, judge `gemini`, `적용`
Then (1)의 바뀌는 설정은 정확히 `spec.review_gate.providers [claude, codex, gemini] → [claude, codex]`, `orchestra.commands.plan.providers …`, `orchestra.commands.review.providers …`, `orchestra.commands.secure.providers [claude, codex, gemini] → [claude, codex]`, `orchestra.commands.secure.judge gemini 지움`이고 brainstorm·`orchestra.judge` 줄은 없다
And (1) 뒤 brainstorm은 [claude, codex], secure에는 judge key가 없고 `orchestra.judge`는 codex다
And (2)의 judge Select는 `그대로 — claude`, `codex`, `gemini`이고, 적용 뒤 다섯 목록은 모두 [claude, codex, gemini], `spec.review_gate.judge` = gemini, `orchestra.judge` = codex(편집 전 두 judge가 달랐다)다

### S10: init은 탭 편집을 덮지 않는다
Priority: Must
Given 빈 dir과, init이 감지할 provider가 claude·codex·gemini가 되도록 PATH를 정한 temp bin이 있다
When accessible runner로 `auto init`을 열어 Review Gate `예`, 구성 단계에서 구성 편집 → 멀티프로바이더 리뷰 → gemini `끄기` → `적용` → 유지, 나머지 wizard Enter를 고른다
Then 생성된 `autopus.yaml`의 `spec.review_gate.providers`는 [claude, codex]이고 `orchestra.commands`의 모든 `providers`에 gemini가 없다
And 같은 입력에서 리뷰 탭만 열지 않은 run의 `spec.review_gate.providers`는 [claude, codex, gemini]다

### S11: update는 편집할 때만 쓴다
Priority: Must
Given R3(B0), CATG가 있다
When `auto update`에서 (1) 리뷰 탭의 모든 Select Enter, (2) gemini `끄기` → `적용` → 유지를 고르고 update가 끝날 때까지 둔다
Then (1)은 기존 `변경 없음` 줄과 B0와 같은 SHA-256이다
And (2)는 `적용했습니다` 줄 뒤 update의 생성과 저장이 끝난 파일에서도 `spec.review_gate.providers`가 [claude, codex]다

### S12: 안전하지 않은 node는 쓰지 않는다
Priority: Must
Given R3의 두 사본이 있다: (a) `orchestra.providers.codex.model: ${CODEX_REVIEW_MODEL}`과 환경 변수 `CODEX_REVIEW_MODEL=openai-codex/gpt-6-astra:max`, (b) `orchestra.commands.review.providers: &rp [claude, codex, gemini]`와 `spec.review_gate.providers: *rp`
When (a)는 codex model `openai-codex/gpt-6-luna` → `high` → `적용`, (b)는 gemini `끄기` → `적용`을 고른다
Then (a)는 `autopus.yaml의 리뷰 설정을 안전하게 고칠 수 없어 아무것도 쓰지 않았습니다: orchestra.providers.codex.model`, (b)는 같은 줄에 `spec.review_gate.providers`를 쓴다
And 두 파일의 SHA-256과 run의 `cfg.Orchestra`·`cfg.Spec.ReviewGate`는 적용 전과 같다

### S13: 영어 문구와 메뉴 요약
Priority: Should
Given R3, CATG, `tui.DetectSystemLang`이 en인 환경과 ko인 환경이 있다
When 각 환경에서 구성 메뉴를 열고 S1과 같은 입력을 준다
Then en 출력에 `Multi-provider review`, `gemini takes part in review`, `Multi-provider review: claude · codex · gemini · judge claude`가 있다
And ko 출력의 메뉴에 `멀티프로바이더 리뷰: claude · codex · gemini · judge claude`가 있고 적용 뒤 `멀티프로바이더 리뷰: claude · codex · gemini(꺼짐) · judge claude`가 있으며, 두 환경이 쓴 파일 bytes는 같다

### S14: 문서
Priority: Should
Given 이 branch가 있다
When init·update·quality help, README, CHANGELOG를 읽는다
Then 리뷰 탭, 쓰는 key 여섯, 확인 호출과 비용, CLI backend 한계가 있고 기존 test가 확인하는 문장은 그대로다

### S15: 실제 터미널 smoke
Priority: Must
Given 이 repo(gemini가 `google-antigravity/gemini-3.1-pro:high`)와 실제 터미널, 사용자 동의(model 호출 최대 6회)가 있다
When `auto update`에서 리뷰 탭 → 확인 호출 → (실패가 있으면) 다시 편집 → gemini `google-antigravity/gemini-3.8-flash` → `high` → 확인 호출 → `적용`을 하고 `auto spec review`를 한 번 돌린다
Then `evidence/manual-smoke-revmodel.txt`에 두 확인 호출의 결과 줄, 바뀐 `orchestra.providers.gemini.model` 줄, 리뷰 receipt의 provider 상태가 경로를 치환해 남는다
And 기록이 없으면 CD-1이 남고 완료가 아니다

## Oracle Acceptance Notes

- Must 시나리오는 concrete 값을 판정한다: Select 선택지 목록과 순서·기본값, 바뀌는 설정 줄과 순서, 결과·안내·오류 문구, `config.Load` 값, 정확한 selector 문자열, FB 호출 수·Prompt·Timeout·`Config.Model`, diff 줄 수, SHA-256, v0.50.125의 exit 코드와 오류 문구. 파일 존재나 종료 코드만으로 닫는 Must는 없다(S7의 exit 0은 대조군의 오류 문구와 key 경로 비교가 함께 판정한다).
- 입력은 provider·backend·family·catalog 상태가 서로 다른 fixture다: 세 family의 OMP backend(R3), 명령별로 다른 구성(RC), CLI backend(RL), catalog 셋(CATG·CATN·CATX), 통과·404가 섞인 확인 호출(FB).
- 기대 값은 2026-10-15에 이 branch 코드와 probe A1–A3에서 계산했다. 실행으로 확인한 것은 A1–A3뿐이다. 코드 사실과 다르면 test가 아니라 이 SPEC을 고친다.
- S15는 model 호출 비용이 들어 사용자 동의 뒤에만 돈다. 기록 전에는 `research.md` CD-1이 남는다.
