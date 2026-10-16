# SPEC-REVMODEL-001 수락 기준
## Test Scenarios

fixture는 모두 temp project 안에 만든다. 폼 test는 SPEC-MODELUX-002처럼 `composeFormsProbe`를 true로, `composeRunForm`을 huh accessible runner로 두고 입력을 `testing/iotest.OneByteReader`로 준다. "X를 고른다"는 그 라벨의 번호 입력이고 Enter는 기본값이다. OMP catalog는 `ompCLIFakeRunner`, 확인 요청은 요청을 기록하는 fake `providerSmokeBackendFactory`(FB)다. 설치 확인은 PATH 앞에 둔 temp bin의 빈 실행 파일로 통과시킨다. "loader 값"은 `t.Chdir` 뒤 `loadHarnessConfigForFlags(globalFlags{})`의 결과다. 기준 바이너리는 `git archive v0.50.125`(dd6d6ea2)로 만든 `auto`다. 실제 터미널은 S15에서만 쓴다.

- R3: `platforms: [claude-code, codex, omp]`. `orchestra.judge: claude`. `orchestra.providers`의 claude·codex·gemini는 모두 `backend: omp`, `subprocess.timeout: 900`이고 model은 차례로 `anthropic/claude-fable-5-1:max`, `openai-codex/gpt-6-astra:max`, `google-antigravity/gemini-3.1-pro:high`다. `orchestra.commands`의 brainstorm·plan·review·secure는 모두 `providers: [claude, codex, gemini]`다. `spec.review_gate`는 `enabled: true`, `strategy: debate`, `providers: [claude, codex, gemini]`, `judge: claude`, `max_revisions: 2`다. 이 branch의 `auto update --yes`로 한 번 맞춘 bytes를 B0라 한다.
- RH: R3와 같은 값을 손으로 쓴 파일. 들여쓰기 2칸, `spec:` 바로 위에 주석 줄 `# review settings`, 구역 사이 빈 줄, `providers: [claude, codex, gemini] # reviewers`. retired key는 없다.
- RL: R3에서 `orchestra.providers`만 이 branch의 기본 항목(`DefaultClaudeProviderEntry`, `DefaultCodexProviderEntry`, gemini `binary: agy`)으로 바꾼 것. 셋 다 `backend`가 없다. RL을 이 branch의 `auto update --yes`로 한 번 맞춘 bytes를 B1이라 한다. CATC: Codex catalog(`codex debug models` JSON). `gpt-6-astra`·`gpt-6-luna`는 `low`–`max` 다섯 단계, `gpt-6.1-sol`은 `low`·`medium`·`high`·`xhigh`(SPEC-MODELUX-003 CAT-E와 같다). RB: R3의 `spec.review_gate.providers: [claude]`. RN: R3에서 `spec.review_gate.judge` 줄을 지운 것. RE: R3의 그 줄이 `judge: ""`인 것. RD: R3의 `enabled: false`.
- CATG: strict catalog. 보고 thinking은 `anthropic/claude-fable-5-1` [low, medium, high, max], `anthropic/claude-opus-5-5` [high, max], `openai-codex/gpt-6-astra` [low, medium, high, xhigh, max], `openai-codex/gpt-6-luna` [low, medium, high], `google-antigravity/gemini-3.1-pro` [low, high], `google-antigravity/gemini-3.8-flash` [low, medium, high], `example/foo-1` [high]. 정규화 뒤 순서는 정렬 순서다(예: luna는 `high`, `low`, `medium`).
- CATN: CATG에서 `google-antigravity/gemini-3.1-pro`를 뺀 것. CATE: CATG에 `thinking: []`인 `example/bar-1`을 더한 것(`catalog_invalid`). CATX: `omp models`가 오류를 내는 runner.
- FB 응답: claude·codex와 `google-antigravity/gemini-3.8-flash:high`인 gemini는 output `AUTOPUS_PROVIDER_SMOKE_OK`, `google-antigravity/gemini-3.1-pro:high`인 gemini는 error `HTTP 404 Not Found (Cloud Code Assist)`. FB는 요청마다 Provider, Prompt, Timeout, `Config.Model`을 기록한다.

### S1: gemini 끄기는 리뷰 목록만 바꾼다
Priority: Must
Given R3(B0), CATG가 있다
When `auto update`에서 구성 편집 → 멀티프로바이더 리뷰 → claude Enter → codex Enter → gemini `끄기` → claude·codex model Enter → judge Enter → 확인 호출 Enter → `적용` → 유지를 고른다
Then 바뀌는 설정은 `spec.review_gate.providers [claude, codex, gemini] → [claude, codex]` 한 줄이고, 그 뒤 `꺼진 provider도 orchestra 명령(brainstorm·plan·review·secure)과 auto spec review --multi에서는 계속 돕니다`가 있다
And B0와의 diff는 그 목록 줄 하나이고 `orchestra:` 구역 bytes는 B0와 같다. loader 값의 `spec.review_gate.providers`는 [claude, codex], `spec.review_gate.judge`는 claude, 네 명령의 `providers`는 [claude, codex, gemini]다
And loader 값으로 `selectSpecReviewProviderNames(cfg, nil, false)`는 [claude, codex]이고 FB 요청은 0회다

### S2: model과 thinking은 정확한 selector로 쓴다
Priority: Must
Given R3(B0), CATG가 있다
When 리뷰 탭에서 참여 Enter 셋, claude model `anthropic/claude-opus-5-5` → thinking Enter, codex model `openai-codex/gpt-6-luna` → thinking Enter, gemini model Enter, judge Enter, 확인 호출 Enter, `적용`을 고른다
Then codex model Select는 정확히 `그대로 — openai-codex/gpt-6-astra:max`, `openai-codex/gpt-6-astra`, `openai-codex/gpt-6-luna`이고 `example/foo-1`은 어느 Select에도 없다
And codex thinking Select는 `high`, `low`, `medium`이고 기본은 `high`다(working의 `max`를 luna가 보고하지 않는다). claude thinking Select는 `high`, `max`이고 기본은 `max`다
And 바뀌는 설정은 `orchestra.providers.claude.model anthropic/claude-fable-5-1:max → anthropic/claude-opus-5-5:max`, `orchestra.providers.codex.model openai-codex/gpt-6-astra:max → openai-codex/gpt-6-luna:high` 순서이고 `바뀐 model은 auto spec review와 orchestra 명령 모두에 쓰입니다`가 뒤따른다
And 적용 뒤 `orchestra.providers.codex.model`은 문자열 `openai-codex/gpt-6-luna:high`, claude는 `anthropic/claude-opus-5-5:max`, B0와의 diff는 두 줄이며 loader 값으로 조립한 리뷰 provider codex의 `Model`도 `openai-codex/gpt-6-luna:high`다

### S3: judge와 provider 수 검증
Priority: Must
Given R3(B0), CATG가 있다
When (a) `planReviewComposition`에 참여 {claude, codex}·judge gemini를 준다. (b) 리뷰 탭에서 claude `끄기`, 나머지 Enter, judge Enter, 확인 호출 Enter, `적용`. (c) 새 run에서 codex `끄기`, gemini `끄기`, 오류 뒤 `다시 고르기` Enter, 그 뒤 모든 Select Enter, 유지. (d) 새 run에서 셋 다 `끄기`, 오류 뒤 `나가기`, 유지
Then (a)는 `judge gemini는 리뷰에 참여하지 않는 provider라 고를 수 없습니다` 오류다
And (b)의 judge Select 앞에 `judge claude가 꺼져 judge를 다시 고릅니다`가 있고 선택지는 정확히 `codex`, `gemini`, 기본 `codex`다. 적용 뒤 `spec.review_gate.judge` = codex, `spec.review_gate.providers` = [codex, gemini], `orchestra.judge` = claude이고 B0와의 diff는 두 줄이다
And (c)·(d)는 참여 Select 바로 뒤 `리뷰에는 provider가 두 개 이상 필요합니다 (지금 1개)`·`(지금 0개)`를 쓰고 그 pass에 model·judge Select가 없으며, 오류 뒤 Select는 정확히 `다시 고르기`(기본)·`나가기`다. (c)에서 다시 묻는 참여 Select의 기본값은 baseline(`그대로 — 켜짐`)이라 다음 pass는 `변경 없음`이고, (d)는 `나가기` 뒤 바로 메뉴다. 두 run 뒤 파일 SHA-256은 B0와 같다

### S4: catalog 상태와 증명
Priority: Must
Given R3, CATN, CATX, CATE가 있다
When (a) CATN으로 리뷰 탭에서 gemini model Enter, codex model `openai-codex/gpt-6-luna` → Enter, judge Enter, 확인 호출 Enter, `적용`. (b) `planReviewComposition`에 ① CATN과 gemini 새 model `google-antigravity/gemini-3-pro:high`(baseline과 다르고 catalog에 없음), ② CATN과 baseline 그대로의 gemini model `google-antigravity/gemini-3.1-pro:high`, ③ CATG와 codex `openai-codex/gpt-6-luna` thinking `xhigh`를 준다. (c) CATX와 CATE로 각각 gemini `끄기` → `적용`
Then (a)의 gemini model Select는 정확히 `그대로 — google-antigravity/gemini-3.1-pro:high (catalog에 없음)`, `google-antigravity/gemini-3.8-flash`이고, 적용은 성공해 gemini model은 그대로, codex model은 `openai-codex/gpt-6-luna:high`다
And (b)의 ①은 `gemini model google-antigravity/gemini-3-pro:high는 설치된 OMP catalog가 증명하지 않습니다`, ③은 `codex thinking xhigh는 openai-codex/gpt-6-luna가 보고하지 않는 단계입니다` 오류이고, ②는 baseline과 같은 값이라 오류가 없다
And (c)는 둘 다 `OMP catalog를 읽지 못해 model은 그대로입니다`를 한 번 쓰고 model Select 없이 적용에 성공하며 `spec.review_gate.providers`는 [claude, codex]다

### S5: 확인 요청과 404
Priority: Must
Given R3(B0), CATG, FB가 있다
When 한 `auto update` run에서 (1) codex model `openai-codex/gpt-6-luna` → Enter, 나머지 Enter, `확인 호출 — provider 3개에 요청 1회씩 (재시도 포함 최대 9회 model 호출)` → 확인 Enter. (2) 리뷰 탭의 모든 Select Enter → 같은 확인 호출. (3) gemini model `google-antigravity/gemini-3.8-flash` → Enter, 확인 호출 → `적용`. (4) claude model `anthropic/claude-opus-5-5` → Enter, `확인 호출 없이 계속` → `취소`
Then (1)에서 FB 요청은 claude, codex, gemini 순서로 3회이고 Prompt는 모두 `Return ONLY this exact token and no other text: AUTOPUS_PROVIDER_SMOKE_OK`, Timeout은 모두 60초, codex `Config.Model`은 `openai-codex/gpt-6-luna:high`다
And (1)의 출력은 `확인 호출 claude (anthropic/claude-fable-5-1:max): 통과`, `확인 호출 codex (openai-codex/gpt-6-luna:high): 통과`, `확인 호출 gemini (google-antigravity/gemini-3.1-pro:high): 실패 — HTTP 404 Not Found (Cloud Code Assist)`, `확인 호출이 실패한 provider가 있습니다. 적용하면 그 provider는 리뷰에서 실패합니다.`이고, 확인 Enter(기본 `취소`) 뒤 파일 SHA-256이 B0와 같다
And (2)는 `변경 없음` 뒤 같은 세 결과 줄(codex는 `openai-codex/gpt-6-astra:max`)과 실패 안내를 쓰고 확인 Select 없이 메뉴로 돌아간다
And (3)의 gemini 줄은 `확인 호출 gemini (google-antigravity/gemini-3.8-flash:high): 통과`이고 실패 안내가 없으며 적용 뒤 gemini model은 `google-antigravity/gemini-3.8-flash:high`, codex는 `openai-codex/gpt-6-astra:max`다. (4) 뒤 FB 요청은 모두 9회다

### S6: 비대화형과 CI에서는 탭이 없다
Priority: Must
Given R3(B0), CATG, FB가 있다
When (a) `auto update --yes`, (b) stdin이 `strings.NewReader("\n\n\n")`이고 stdout이 buffer인 `auto update`, (c) 같은 stdin의 `auto quality --apply`를 실행하고, (d) `composeFormsProbe`·`composeTTYProbe`가 true이고 `composeCIEnv`가 `true`인 상태로 `auto update`, `auto quality --apply`, 빈 dir의 `auto init`을 tool tab Select까지 연다
Then (a)–(c) 출력 어디에도 `멀티프로바이더 리뷰`와 `리뷰 참여`가 없고 `composeRunForm` 기록에 리뷰 탭 title이 없으며, (d)의 update는 구성 단계를 열지 않고 quality·init의 tool tab 선택지에는 `멀티프로바이더 리뷰`가 없다
And 네 경우 모두 FB 요청과 탭의 catalog 읽기가 0회이고, (a)–(d)의 update·quality 뒤 파일 SHA-256은 B0와 같다

### S7: v0.50.125가 쓴 파일을 읽는다
Priority: Must
Given S1, S2, S3(b), S5(3), S9(b)가 남긴 파일과 대조군(S1 파일의 `spec.review_gate`에 `reviewers: [claude]`를 더한 것)이 있다
When 각 파일의 project에서 v0.50.125 바이너리로 `auto config get usage_profile`을 실행하고, 각 파일의 YAML key 경로 집합을 그 run의 시작 파일(S9는 RN, 나머지는 B0)과 비교한다
Then 다섯 파일은 exit 0이고 대조군은 exit 1과 `field reviewers not found in type config.ReviewGateConf`다
And 다섯 파일의 key 경로 집합은 시작 파일과 같고, 더해진 key는 S9(b)의 `spec.review_gate.judge` 하나뿐이다

### S8: OMP가 없으면 agy는 이유 줄이다
Priority: Must
Given RL, CATC, 그리고 PATH에 `omp`가 없는 temp bin이 있다
When 리뷰 탭에서 참여 Enter 셋, claude model Enter, codex model Enter, judge Enter, 확인 호출 Enter를 고르고, 다시 리뷰 탭에서 codex `끄기`, 나머지 Enter, 확인 호출 Enter, `적용`을 고른다
Then 첫 pass에 `gemini(agy)는 Autopus가 증명할 model 목록이 없어 model을 바꾸지 않습니다. OMP를 설치하면 OMP backend로 전환할 수 있습니다`가 한 번 있고 `OMP backend로 전환` 선택지와 gemini model Select는 없으며 claude·codex model Select는 있다. after 표의 model 칸은 `claude claude-fable-5-1 · max`, `codex gpt-6-astra · max`, `agy`다
And 첫 pass는 `변경 없음`이고, 두 번째 pass 뒤 `spec.review_gate.providers`는 [claude, gemini], `orchestra.providers` mapping은 적용 전과 deep-equal이다

### S9: judge 없음과 judge key 삽입
Priority: Must
Given RN, RE, CATG가 있다
When (a) RN에서 gemini `끄기`, judge Enter, `적용`. (b) RN에서 참여 Enter, judge `codex`, `적용`. (c) RE에서 judge `claude`, `적용`
Then (a)의 judge Select는 정확히 `그대로 — judge 없음`, `claude`, `codex`이고 적용 뒤 파일에 `spec.review_gate.judge` key가 없으며 diff는 목록 줄 하나다
And (b)는 바뀌는 설정 `spec.review_gate.judge judge 없음 → codex`를 쓰고, review_gate mapping의 마지막 줄 바로 뒤에 같은 들여쓰기의 `judge: codex` 한 줄만 더해지며 loader 값의 judge는 codex다
And (c)는 `judge: ""` 줄이 `judge: claude`로 바뀐 한 줄 diff다

### S10: init은 탭 편집을 덮지 않는다
Priority: Must
Given 빈 dir과, init이 감지할 provider가 claude·codex·gemini가 되도록 PATH를 정한 temp bin이 있다
When accessible runner로 (a) `auto init`에서 Review Gate `예`, 구성 편집 → 멀티프로바이더 리뷰 → gemini `끄기` → `적용` → 유지, 나머지 Enter를 고르고, (b) 같은 입력에서 리뷰 탭만 열지 않고, (c) Review Gate `아니오`로 구성 편집까지 연다
Then (a)의 `spec.review_gate.providers`는 [claude, codex]이고 `orchestra.commands`의 각 `providers`는 init의 migration(`init.go:161`) 결과대로 gemini를 담는다
And (b)의 `spec.review_gate.providers`는 [claude, codex, gemini]이고, (c)의 tool tab 선택지에는 `멀티프로바이더 리뷰`가 없다

### S11: update는 편집할 때만 쓰고 다음 update도 되돌리지 않는다
Priority: Must
Given R3(B0), CATG가 있다
When `auto update`에서 (1) 리뷰 탭의 모든 Select Enter → 확인 호출 Enter, (2) gemini `끄기` → `적용` → 유지를 고르고 update가 끝날 때까지 둔 뒤, (3) `auto update --yes`를 한 번 더 실행한다
Then (1)은 기존 `변경 없음` 줄과 B0와 같은 SHA-256이다
And (2)는 `적용했습니다` 줄 뒤 update가 끝난 파일에서 `spec.review_gate.providers`가 [claude, codex]이고, (3) 뒤 파일 SHA-256은 (2) 뒤와 같다

### S12: 안전하지 않은 대상은 쓰지 않고 손으로 쓴 파일은 span 밖을 지킨다
Priority: Must
Given R3의 사본 셋과 RH가 있다: (a) `orchestra.providers.codex.model: ${CODEX_REVIEW_MODEL}`과 환경 변수 `CODEX_REVIEW_MODEL=openai-codex/gpt-6-astra:max`, (b) `orchestra.commands.review.providers: &rp [claude, codex, gemini]`와 `spec.review_gate.providers: *rp`, (c) `spec.review_gate.providers: &rp [claude, codex, gemini]`와 `orchestra.commands.review.providers: *rp`
When (a)는 codex model `openai-codex/gpt-6-luna` → Enter → `적용`, (b)·(c)는 gemini `끄기` → `적용`, RH는 gemini `끄기`와 codex model `openai-codex/gpt-6-luna` → Enter → `적용`을 고른다
Then (a)는 `autopus.yaml의 리뷰 설정을 안전하게 고칠 수 없어 아무것도 쓰지 않았습니다: orchestra.providers.codex.model`, (b)·(c)는 같은 줄에 `spec.review_gate.providers`를 쓰고, 세 파일의 SHA-256과 run의 `cfg.Orchestra`·`cfg.Spec.ReviewGate`는 적용 전과 같다
And RH의 결과는 목록 줄이 `    providers: [claude, codex] # reviewers`(4칸, review_gate 아래), codex model 줄의 값만 `openai-codex/gpt-6-luna:high`로 바뀐 두 줄 diff이고, 그 두 span 밖 bytes(주석 줄, 빈 줄, 2칸 들여쓰기)는 RH와 같다

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
Then 리뷰 탭, 쓰는 key 셋, 확인 요청의 최대 호출 수, CLI backend 한계, orchestra 명령과 `--multi`가 모든 provider를 쓴다는 문장이 있고 기존 test가 확인하는 문장은 그대로다

### S15: 실제 터미널 smoke
Priority: Must
Given 이 repo와 실제 터미널, 사용자 동의(확인 요청 두 번, model 호출 보통 6회, 재시도 포함 최대 18회)가 있다. 사용자 결정(2026-10-16)대로 목표 gemini 리뷰 model은 `google-antigravity/gemini-3.8-flash:high`다
When `auto update`에서 리뷰 탭 → 모든 Select Enter → 확인 호출(편집 없음), 다시 구성 편집 → 리뷰 탭 → gemini model을 목표 값으로(이미 목표 값이면 다른 증명된 google model로) → Enter → 확인 호출 → `적용` → 유지를 하고(뒤의 경우는 같은 탭으로 목표 값에 되돌리고 확인 요청은 하지 않는다), `auto doctor`(model 호출 없음)를 실행한다
Then `evidence/manual-smoke-revmodel.txt`에 두 확인 요청의 결과 줄, 바뀐 `orchestra.providers.gemini.model` 줄, doctor의 리뷰 provider readiness 줄이 경로를 치환해 남는다
And `auto spec review`는 이 smoke에서 돌리지 않는다. 기록이 없으면 CD-1이 남고 완료가 아니다

### S16: 터미널 제어 문자
Priority: Must
Given OMP review backend와, 이름이 `gem\x1b[2Jini`(ESC 포함)인 provider를 담은 cfg가 있다
When (a) 그 이름과 유효한 selector로 `ompReviewBackend.Execute`를 부르고 os.Stderr를 pipe로 받는다. (b) 같은 cfg로 리뷰 탭의 확인 호출을 고른다
Then (a)의 stderr에는 0x1b byte가 없고 `provider=gem\x1b[2Jini`가 역슬래시 글자로 있다
And (b)는 `확인 호출을 할 수 없습니다: provider 이름 gem\x1b[2Jini에 터미널 제어 문자가 있습니다`를 쓰고 FB 요청은 0회이며, 제어 문자가 없는 이름의 (a) 출력은 지금과 bytes가 같다

### S17: 다시 편집은 working을 기본값으로, baseline과 비교한다
Priority: Must
Given R3(B0), CATG가 있다
When (1) 리뷰 탭에서 gemini `끄기`, 나머지 Enter, 확인 호출 Enter → `다시 편집` → 모든 Select Enter → 확인 호출 Enter → `적용`. (2) 새 run에서 gemini `끄기`, 나머지 Enter, 확인 호출 Enter → `다시 편집` → gemini `켜기`, 나머지 Enter
Then (1)의 두 번째 pass에서 gemini 참여 Select의 기본은 `그대로 — 꺼짐`이고 미리보기는 다시 `spec.review_gate.providers [claude, codex, gemini] → [claude, codex]`를 쓰며, 적용 뒤 그 목록은 [claude, codex]다
And (2)의 두 번째 pass는 `변경 없음`을 쓰고 확인 Select 없이 메뉴로 돌아가며 파일 SHA-256은 B0와 같다

### S18: baseline 참여가 둘보다 적어도 들어가고 나갈 수 있다
Priority: Must
Given RB, CATG가 있다
When (1) 리뷰 탭을 열고 참여 Select를 모두 Enter, 오류 뒤 `나가기`, 유지를 고른다. (2) 새 run에서 리뷰 탭의 codex `켜기`, 나머지 Enter, judge Enter, 확인 호출 Enter, `적용`, 유지를 고른다
Then (1)은 첫 참여 Select 앞에 `지금 리뷰 provider가 1개뿐입니다. 두 개 이상 켜야 적용할 수 있습니다`를 쓰고, 참여 Select 뒤 `리뷰에는 provider가 두 개 이상 필요합니다 (지금 1개)`와 정확히 `다시 고르기`(기본)·`나가기`인 Select를 보인 뒤 메뉴로 돌아가며 파일 SHA-256은 RB와 같다
And (2)의 바뀌는 설정은 `spec.review_gate.providers [claude] → [claude, codex]`이고 적용 뒤 loader 값의 그 목록은 [claude, codex], judge는 claude다

### S19: claude CLI model은 args `--model`로 쓴다
Priority: Must
Given RL(B1), CATC가 있다
When (1) 리뷰 탭에서 claude model `claude-opus-5-5` → effort `xhigh`, 나머지 Enter, 확인 호출 Enter, `적용`. (2) 다시 claude model `claude-haiku-4-5`, 나머지 Enter, 확인 호출 Enter, `적용`
Then (1)의 claude model Select는 정확히 `그대로 — claude-fable-5-1 · max`, `claude-fable-5-1`, `claude-opus-5-5`, `claude-sonnet-5-5`, `claude-haiku-4-5`이고 effort Select는 `low`, `medium`, `high`, `xhigh`, `max`, 기본 `max`다
And (1) 뒤 `orchestra.providers.claude.args`는 `[--print, --model, claude-opus-5-5, --effort, xhigh]`, B1과의 diff는 그 줄 하나이고, loader 값으로 조립한 claude argv(전역 `--effort` 없음)에 `--model claude-opus-5-5 --effort xhigh`가 이어져 있다
And (2)는 effort Select 없이 args를 `[--print, --model, claude-haiku-4-5]`로 쓴다

### S20: codex CLI model은 `pinned`와 `-m`으로 쓰고 profile이 덮지 않는다
Priority: Must
Given RL(B1), CATC가 있다
When 리뷰 탭에서 codex model `gpt-6.1-sol` → effort `high`, 나머지 Enter, 확인 호출 Enter, `적용`을 고른 뒤, 그 파일과 B1을 각각 loader 값 → `applyRuntimeHarnessOverrides(globalFlags{Effort: "low"})` → `assembleSpecReviewProviders`로 조립한다
Then codex model Select는 정확히 `그대로 — gpt-6-astra · max`, `gpt-6-astra`, `gpt-6-luna`, `gpt-6.1-sol`이고, `gpt-6.1-sol` effort Select는 `low`, `medium`, `high`, `xhigh`이며 `max`가 없고 기본은 `xhigh`다
And 적용 뒤 codex 항목은 `model_policy: pinned`, `args: [exec, --json, --sandbox, workspace-write, -m, gpt-6.1-sol, -c, model_reasoning_effort="high"]`이고 B1과의 diff는 그 두 줄이다
And 조립한 codex argv는 이 파일에서 `-m gpt-6.1-sol`과 `model_reasoning_effort="high"`를, B1에서는 `-m gpt-6-astra`와 `model_reasoning_effort="low"`를 담는다

### S21: agy는 OMP backend로 전환할 수 있다
Priority: Must
Given RL(B1), CATG, CATC, temp bin의 `omp`가 있다
When 리뷰 탭에서 gemini `OMP backend로 전환` → model `google-antigravity/gemini-3.8-flash` → thinking Enter, 나머지 Enter, 확인 호출 Enter, `적용`을 고른다
Then gemini 선택지는 정확히 `그대로 — CLI backend`, `OMP backend로 전환`이고, 전환 뒤 model Select는 정확히 `google-antigravity/gemini-3.1-pro`, `google-antigravity/gemini-3.8-flash`, thinking Select는 `high`, `low`, `medium`, 기본 `high`다
And 적용 뒤 gemini 항목에는 `backend: omp`와 `model: google-antigravity/gemini-3.8-flash:high` 두 줄만 더해지고 `binary: agy`, `args`, `prompt_via_args`, `subprocess` 줄은 bytes가 같다
And loader 값으로 조립한 gemini는 `backend="omp"`, `model="google-antigravity/gemini-3.8-flash:high"`이고 read-only 정책을 지난다

### S22: v0.50.125가 CLI·전환 파일을 읽는다
Priority: Must
Given S19(1), S19(2), S20, S21이 남긴 파일과 대조군(S20 파일의 codex에 `models: [x]` key를 더한 것)이 있다
When 각 파일의 project에서 v0.50.125 바이너리로 `auto config get usage_profile`을 실행하고 key 경로 집합을 B1과 비교한다
Then 네 파일은 exit 0이고 대조군은 exit 1과 `field models not found`를 담은 오류다
And 더해진 key는 S21의 `orchestra.providers.gemini.backend`·`orchestra.providers.gemini.model` 둘뿐이고 S20의 `model_policy`는 값만 바뀐다

## Oracle Acceptance Notes

- Must 시나리오는 concrete 값을 판정한다: Select 선택지 목록·순서·기본값, 바뀌는 설정 줄과 순서, 결과·안내·오류 문구, loader 값, 정확한 selector 문자열, FB 요청 수·Prompt·Timeout·`Config.Model`, diff 줄 수, span 밖 bytes, SHA-256, v0.50.125의 exit 코드와 오류 문구. 파일 존재나 종료 코드만으로 닫는 Must는 없다.
- 입력은 provider·backend·family·catalog 상태가 서로 다른 fixture다: 세 family의 OMP backend(R3), 손으로 쓴 파일(RH), CLI backend(RL), baseline 참여 1개(RB), judge 없음(RN·RE), 꺼진 gate(RD는 S10(c)와 같은 규칙), catalog 넷(CATG·CATN·CATE·CATX), 통과·404가 섞인 확인 요청(FB).
- CLI model 시나리오(S19–S21)는 조립된 argv와 backend를 판정한다. 실제 subprocess는 띄우지 않는다. argv는 `Config.Args` 그대로라 조립 결과가 실행 argv의 판정이다(`evidence/probe-cli-model.md`).
- FB는 `Execute` 요청을 센다. 한 요청 안의 재시도(최대 3회)는 backend 안이라 FB가 보지 않으므로 라벨의 최대 호출 수(3n)는 `ompReviewMaxAttempts`로 판정하고, S15의 동의 범위가 그 수를 덮는다.
- 기대 값은 2026-10-15에 이 branch 코드와 probe A1–A3에서 계산했다. thinking 순서는 catalog 정규화의 정렬을 따른다. 코드 사실과 다르면 test가 아니라 이 SPEC을 고친다.
