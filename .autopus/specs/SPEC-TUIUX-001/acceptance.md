# SPEC-TUIUX-001 수락 기준
## Test Scenarios

fixture는 모두 temp 부모 디렉터리 아래 이름이 `p`인 project 안에 만든다(project 이름이 디렉터리 이름이라서, `init.go:52`). 폼 test는 accessible helper로 돈다. 구성 단계와 init 확인 화면은 `useNavComposeForms(t)`(`composeFormsProbe` true, runner `tui.RunNavAccessible(form, frame, cmd.InOrStdin(), cmd.OutOrStdout())`), wizard는 `InitWizardOpts.Runner`에 같은 helper를 건다. 입력은 `formInput`(`testing/iotest.OneByteReader`)이고 한 항목이 한 줄이다. "X를 고른다"는 그 라벨의 번호를 입력한다는 뜻, Enter는 빈 줄, `b`는 뒤로 token 줄이다. 대신하는 seam은 TTY·CI probe, 폼 판정과 실행, `composeLanguage`, Codex catalog probe, OMP runner·activate, REA Node probe(pass, SPEC-REA-001 test seam), init의 `runInitWizard`뿐이다. 나머지는 실제 명령(`NewRootCmd`)과 실제 adapter `Generate`로 돈다. TUI 키는 S4·S20에서 실제 `tea.Program`과 `io.Pipe`로, 실제 터미널은 S18에서만 쓴다.

- ko는 `LANG=ko_KR.UTF-8`, en은 `LANG=en_US.UTF-8`이고 `LC_ALL`·`LC_MESSAGES`·`LANGUAGE`는 빈 값이다. 구성 단계는 `composeLanguage`를 `"ko"`나 `"en"`으로 고정한다. `TERM`은 S21 밖에서 `xterm-256color`다.
- WZ6: `InitWizardOpts{Platforms: [claude-code, codex], Providers: [claude, codex]}`. 단계는 사용 프로필, 언어 설정, 품질 모드, 리뷰 게이트, REA 역공학 MCP, 방법론의 6개다. WZ5: WZ6에 `ExistingProfile: "developer"`(5단계).
- D, P, O3: SPEC-MODELUX-002 `acceptance.md`와 같다. D는 `config.DefaultFullConfig("p")`를 `config.Save`한 파일(platforms `claude-code, codex`, balanced current), P는 D에 `quality.providers.codex: codex-custom`, O3은 D에 platform `omp`와 기존 OMP fake runner·호출을 세는 activate stub이다(OMP 탭은 SPEC-MODELUX-003의 autopus 탭). O3N은 O3에 `role_model_policy.agent_source: native`를 넣은 파일(SPEC-MODELUX-003 S13의 native fixture)이다. RV: D에 orchestra provider 둘 이상을 둔 리뷰 탭 fixture(SPEC-REVMODEL-001 acceptance, `CI` 빈 값).
- CAT-A, CAT-B: SPEC-MODELUX-001 `acceptance.md`와 같다(CAT-B는 CAT-A에서 `gpt-6.1-sol`을 빼고 `gpt-6-sol`을 넣었다). 따로 적지 않으면 CAT-A다. "첫 진입 CAT-A, 둘째 진입 CAT-B"는 Codex catalog probe seam이 첫 호출에 CAT-A, 둘째 호출에 CAT-B를 돌려준다는 뜻이다.
- IN: 빈 project `p`에서 `auto init --dir <DIR> --platforms claude-code,codex`, `composeTTYProbe` true, PATH에 provider CLI 없음, HOME·XDG는 temp. `runInitWizard` seam은 호출과 opts를 기록하고 W를 돌려준다. W = {UsageProfile developer, CommentsLang en, CommitsLang ko, AILang ko, Quality balanced, ReviewGate false, Rea true, Methodology tdd}. W_rg는 W의 Quality가 `role-groups`, W_none은 Methodology가 `none`인 값이다. IR은 seam이 W_rg를 돌려주는 IN이다.
- NPD(normalized project digest): project 아래 모든 파일의 경로와 SHA-256 목록이다. 계산 전에 adapter manifest(`pkg/adapter/manifest.go:75`, `:103`)의 `generated_at` 값을 `<T>`로, `.autopus/backup/<timestamp>/` 경로의 timestamp를 `<TS>`로 바꾼다. T1이 기준을 두 번 돌려 더 찾은 시간 field가 있으면 research `### NPD 정규화`에 적힌 대로 함께 바꾼다. 그 밖의 bytes는 그대로 비교한다.
- 대조 run: 같은 fixture에 마지막 답만 바로 준 run이다(뒤로와 다시 답하기를 뺀 입력). "대조와 같다"는 NPD가 같다는 뜻이다. base golden은 T1이 8ec13935에서 만든 `evidence/base-goldens/`(stdout은 `<DIR>` 치환, NPD)다.

### S1: wizard 뒤로, 미리 선택, 값
Priority: Must
Given WZ6, ko, accessible wizard runner
When 입력 `2`, `2`, `1`, `3`, `b`, Enter, Enter, Enter, `1`, `n`, Enter, `2`로 wizard를 돌린다
Then 결과는 {UsageProfile fullstack, CommentsLang ko, CommitsLang en, AILang ja, Quality ultra, ReviewGate false, Rea true, Methodology none, Cancelled false}이고 같은 답을 뒤로 없이 준 대조 결과와 같다
And header `init · 1/6 사용 프로필`은 1번, `init · 2/6 언어 설정`은 2번, `init · 3/6 품질 모드`는 2번, `init · 4/6 리뷰 게이트`는 1번 나온다
And 두 번째 `init · 2/6 언어 설정` 바로 뒤 줄은 정확히 `지금까지 고른 값`, `  사용 프로필: Fullstack`, `답을 입력하고 Enter를 누르세요 · 뒤로: b`다
And `init · 4/6 리뷰 게이트` 바로 뒤 줄은 정확히 `지금까지 고른 값`, `  사용 프로필: Fullstack`, `  언어 설정: 주석 ko · 커밋 en · AI 응답 ja`, `  품질 모드: Ultra`, `답을 입력하고 Enter를 누르세요 · 뒤로: b`다

### S2: 뒤로는 지금 단계의 부분 답을 버린다
Priority: Must
Given WZ6, ko, accessible wizard runner
When 입력 `2`, `4`, `b`, Enter를 주고 나머지 폼에는 모두 Enter를 준다
Then 결과는 {UsageProfile fullstack, CommentsLang en, CommitsLang ko, AILang ko, Quality balanced, ReviewGate true, Rea true, Methodology tdd}다. 언어 설정의 첫 답 zh는 남지 않는다
And `init · 1/6 사용 프로필`은 2번 나오고 두 번 다 요약 블록이 없으며 바로 뒤 줄이 `답을 입력하고 Enter를 누르세요`다
And 두 번째 사용 프로필 폼의 Enter는 Fullstack을 유지한다

### S3: 첫 폼의 경계
Priority: Must
Given WZ6과 ko·en 각각, 그리고 `NavFrame{Header: "init · 1/6 사용 프로필", First: true, Lang: "ko"}`로 도는 `RunNav`와 Select(A, B, C, 값 b)
When accessible wizard에 `b`, `1`과 나머지 Enter를 주고, `RunNav`에는 `io.Pipe`로 `\x1b`를 쓴 뒤 view를 읽고 `\r`를 쓴다
Then ko 출력에 `첫 단계입니다. 취소하려면 Ctrl+C를 누르세요.`가 정확히 1번, header `init · 1/6 사용 프로필`이 1번 있고 결과는 {UsageProfile developer, CommentsLang en, CommitsLang ko, AILang ko, Quality balanced, ReviewGate true, Rea true, Methodology tdd}다
And en 출력의 안내는 `This is the first step; press Ctrl+C to cancel.` 1번이다
And `RunNav`는 nil을 돌려주고 값은 b이며, Esc 뒤 view(ANSI 제거)의 마지막 줄은 안내 문장이고 첫 줄은 header다

### S4: TUI 키 대응
Priority: Must
Given Select(A, B, C, 값 b)와 Confirm(값 true) 폼, `NavFrame{Header: "h", Summary: ["  s: 1"], First: false, Lang: "ko"}`, 폭 80·높이 24의 `tea.WindowSizeMsg`
When `RunNav`를 실제 `tea.Program`과 `io.Pipe`로 돌리고 한 번의 write마다 (1) `\x1b`, (2) `\x03`, (3) `\r`, (4) `\x1b[B` 뒤 `\r`, (5) `/`, `c`, `\x1b`, `\x1b`, (6) Confirm에 `\x1b`를 쓴다
Then (1) `tui.ErrNavBack`, 값 b (2) `huh.ErrUserAborted`, 값 b (3) nil, 값 b (4) nil, 값 c (5) 첫 Esc 뒤에도 실행 중이고(filter 입력 끝) 둘째 Esc에 `tui.ErrNavBack`이다. 값은 호출자가 저장 state로 되돌리므로 여기서 보지 않는다 (6) `tui.ErrNavBack`, 값 true다
And 실행 중 view(ANSI 제거)의 첫 줄은 `h`, 둘째 줄은 `  s: 1`, 마지막 줄은 `방향키 이동 · Enter 선택 · Esc 뒤로 · Ctrl+C 취소`이고 `enter submit`과 `filter`가 없다

### S5: wizard resume과 요약 전체
Priority: Must
Given WZ6, ko, `Resume` = R {UsageProfile fullstack, CommentsLang ko, CommitsLang en, AILang ja, Quality ultra, ReviewGate false, Rea true, Methodology none}
When (a) 입력 Enter로 wizard를 돌리고 (b) 입력 `b`, Enter, Enter로 다시 돌린다
Then (a)의 첫 header는 `init · 6/6 방법론`이고 바로 뒤 줄은 정확히 `지금까지 고른 값`, `  사용 프로필: Fullstack`, `  언어 설정: 주석 ko · 커밋 en · AI 응답 ja`, `  품질 모드: Ultra`, `  리뷰 게이트: 끔`, `  REA 역공학 MCP: 켬`, footer이며 결과는 R과 같다
And (b)는 `init · 5/6 REA 역공학 MCP` 뒤 `init · 6/6 방법론`을 다시 보이고 결과가 R과 같다

### S6: init 확인 화면의 내용과 적용
Priority: Must
Given IN, ko, `useNavComposeForms`, 확인 폼이 돌 때 `DESIGN.md`와 `autopus.yaml`의 존재를 기록하는 runner hook, 같은 IN 흐름을 8ec13935에서 기록한 base golden
When init을 입력 Enter 하나로 돌린다
Then 출력에 다음 줄이 이 순서로 붙어 있다: `적용할 설정`, ` 답한 값`, `  사용 프로필: Developer`, `  언어 설정: 주석 en · 커밋 ko · AI 응답 ko`, `  품질 모드: Balanced`, `  리뷰 게이트: 끔`, `  REA 역공학 MCP: 켬`, `  방법론: TDD`, ` 바뀌는 설정`, `  language: ai_responses=ko, comments=en, commits=ko`, `  spec.review_gate: enabled=false`, `  usage_profile: developer`, ` REA`, `  REA MCP: 켬 · 등록`, `init · 확인`, `답을 입력하고 Enter를 누르세요 · 뒤로: b`
And 그 뒤는 huh accessible Select의 제목 줄(`이대로 진행할까요?`와 공백 하나), `1. 적용`, `2. 뒤로`, `3. 취소`이고, 확인 폼이 돌 때 두 파일은 없었으며 확인 블록은 `Design Context` 단계 줄보다 앞이다
And run 뒤 NPD가 base golden과 같다

### S7: 확인 화면의 뒤로는 구성을 남기고 마지막 snapshot을 쓴다
Priority: Must
Given IN과 IR, ko, `useNavComposeForms`
When (a) IN에서 seam이 첫 호출에 W, resume 호출에 W_none을 돌려주게 하고 확인 화면에 `2`, 다시 나온 확인 화면에 Enter를 준다
When (b) IR에서 첫 진입 CAT-A, 둘째 진입 CAT-B로 구성 편집 → Codex → 역할 그룹 → 설계·판단 `6.1-sol · xhigh  [opus]` 선택지(SPEC-MODELUX-002 S2와 같은 답, balanced의 `fable`과 다른 값) → 설계·판단 effort `low` 선택지 → 그 뒤 폼 모두 Enter → 적용 → 유지를 고르고, 확인 화면에 `2`, 다시 나온 메뉴(이어 하기)에서 유지, 다시 나온 확인 화면에 Enter를 준다
When (b′) 음성 대조로 (b)와 같은 입력을 두 진입 모두 CAT-A로 돌린다
Then (a) seam은 2번 불리고 둘째 호출의 `Resume`은 W와 같으며, 둘째 확인 블록에 `  방법론: None`이 있고 NPD는 seam이 W_none을 바로 돌려준 대조와 같다
And (b) 보존: seam은 1번 불리고 `구성을 바꿀까요?`는 3번, `init · 확인`은 2번이다. 두 확인 블록 모두 `  quality.providers: codex=codex-custom`과 `  quality.presets.codex-custom.efforts: architect=low, planner=low, security-auditor=low, spec-writer=low` 줄을 가진다. autopus.yaml의 `quality.presets.codex-custom`에서 planner·architect·spec-writer·security-auditor는 `opus`, `efforts`는 넷 모두 `low`다. NPD는 같은 답(`opus`의 CAT-B 선택지, effort `low`)을 CAT-B 한 번 진입으로 준 대조 K와 같다
And (b) 마지막 snapshot: Codex planner agent 파일의 model은 `gpt-6.1-sol`이 아니고 둘째 진입 표의 Codex planner 칸과 같으며 effort는 `low`다. (b′)의 같은 파일 model은 `gpt-6.1-sol`이다(oracle이 두 snapshot을 구별한다는 확인)
And 모든 경우 `Design Context` 단계 줄은 1번이다

### S8: 취소와 Ctrl+C는 아무것도 쓰지 않는다
Priority: Must
Given (a) 빈 IN, (b) `auto init --yes --dir <DIR> --platforms claude-code,codex`로 만든 뒤 autopus.yaml의 platforms 한 항목을 옛 이름 `gemini`로 바꾸고(`config.Load`가 `antigravity-cli`로 고쳐 쓰는 입력, `loader.go:87-94`) 모든 경로의 SHA-256과 mtime을 기록한 IN, (c) 확인 폼 번호에서 `huh.ErrUserAborted`를 돌려주는 runner
When (a)·(b)의 확인 화면에 `3`을 주고, (c)로 (a)·(b)를 다시 돈다
Then 네 run 모두 출력의 마지막 줄이 `  init cancelled.`이고 명령은 오류 없이 끝난다
And (a) 뒤 project 디렉터리에는 아무 경로도 없고, (b) 뒤 모든 경로의 SHA-256과 mtime이 기록과 같으며(`gemini` 줄 그대로) 새 경로가 없다

### S9: header와 footer 문자열
Priority: Must
Given WZ6·WZ5와 ko·en, D·O3·RV의 `auto quality`, D의 `auto update`
When wizard를 4단계까지 돌리고, D quality에서 구성 편집 → Codex → 역할 그룹 → `b`, `b`, `b` → 유지를, D update에서 메뉴 Enter를, O3에서 구성 편집 → OMP → `b` → `b` → 유지를, RV에서 구성 편집 → 멀티프로바이더 리뷰 → `b` → `b` → 유지를 고른다
Then wizard header는 ko `init · 4/6 리뷰 게이트`, en `init · 4/6 Review Gate`, WZ5 ko `init · 3/5 리뷰 게이트`이고 그 폼 제목은 각각 `[4/6] Review Gate`, `[3/5] Review Gate`다
And D quality ko의 header는 차례로 `quality · 모델 구성`, `quality · 모델 구성 › 구성 편집`, `quality · 모델 구성 › 구성 편집 › Codex`, `quality · 모델 구성 › 구성 편집 › Codex › 역할 그룹`, 그다음 뒤로 순서대로 다시 `quality · 모델 구성 › 구성 편집 › Codex`, `quality · 모델 구성 › 구성 편집`, `quality · 모델 구성`이다. en의 넷째 header는 `quality · Model composition › edit the composition › Codex › role groups`다
And D update의 메뉴 header는 `update · 모델 구성`, O3의 OMP 첫 폼은 `quality · 모델 구성 › 구성 편집 › OMP`, RV의 리뷰 첫 폼은 `quality · 모델 구성 › 구성 편집 › 멀티프로바이더 리뷰`, en 확인 화면은 `init · Review`다
And 메뉴와 wizard 1단계 뒤의 footer는 `답을 입력하고 Enter를 누르세요`, 나머지 폼 뒤는 `답을 입력하고 Enter를 누르세요 · 뒤로: b`, en은 `Type an answer and press Enter` / `Type an answer and press Enter · back: b`다

### S10: 역할 그룹 안의 뒤로, effort 복원, pass 버리기
Priority: Must
Given D, ko, `auto quality`, `useNavComposeForms`, Claude Code 탭
When (a) 구성 편집 → Claude Code → 역할 그룹 → 설계·판단 `fable` 선택지 → 설계·판단 effort `low` 선택지 → 구현 `b` → 설계·판단 effort Enter → 구현과 그 뒤 폼 모두 Enter → 적용 → 유지
When (b) 같은 시작에서 설계·판단 `fable` → effort `low` → 구현 `b` → effort `b` → 설계·판단 `opus` 선택지 → effort Enter → 그 뒤 폼 모두 Enter → 적용 → 유지
When (c) 같은 시작에서 설계·판단 `fable` → effort `low` → 구현 `b` → effort `b` → 설계·판단 `b` → 방식 역할 그룹 → 모든 폼 Enter → 유지
Then (a)의 첫 `구현 — executor, …` 폼 앞 블록은 정확히 `지금까지 고른 값`, `  설계·판단: fable`, `  설계·판단 effort: low`이고 두 번째 `설계·판단 effort` 폼 앞 블록은 정확히 `지금까지 고른 값`, `  설계·판단: fable`이다
And (a) autopus.yaml의 `quality.presets.claude-custom.agents`에서 planner·architect·spec-writer·security-auditor가 `fable`, `quality.presets.claude-custom.efforts`에서 넷 모두 `low`이고 NPD는 대조와 같다
And (b) 넷은 `opus`이고 `efforts`에 넷 중 어느 것도 없으며(`low`는 버려졌다) NPD는 opus와 effort Enter를 바로 준 대조와 같다
And (c) 출력에 변경 없음 줄(`composeText.NoChange`)이 1번이고 autopus.yaml SHA-256은 실행 전과 같다

### S11: 에이전트별, 도구 탭, 방식, 기본 프리셋의 뒤로
Priority: Must
Given D, ko, `auto quality`, `useNavComposeForms`
When (a) 구성 편집 → Claude Code → 에이전트별 → 목록 planner → 값 `opus` 선택지 → effort `low` 선택지 → 목록 `b` → effort Enter → 목록 완료 → 적용 → 유지
When (b) 구성 편집 → Codex → 방식 `b` → 도구 탭 Enter → 방식 `b` → 도구 탭 `b` → 유지
When (c) 기본 프리셋 고르기 → `b` → 유지
Then (a) 다시 보인 effort 폼 앞 블록은 정확히 `지금까지 고른 값`, `  planner: opus`이고, autopus.yaml에서 planner는 `opus`, `efforts.planner`는 `low`이며 NPD는 대조와 같다
And (b) 다시 보인 도구 탭의 Enter는 Codex를 고르고(다음 header `quality · 모델 구성 › 구성 편집 › Codex`), `구성을 바꿀까요?`는 2번이며 autopus.yaml SHA-256은 실행 전과 같다
And (c) 출력에 프리셋 쓰기 줄(`quality.default = <프리셋>` 한 줄, `choosePreset`)이 없고 autopus.yaml SHA-256은 실행 전과 같다. 표의 상태 줄 `quality.default = balanced   ultra: …`은 메뉴마다 그려지므로 이 판정에 들지 않는다(T7에서 코드 사실로 고쳤다)

### S12: 적용할까요?와 override 질문의 뒤로
Priority: Must
Given D와 P, ko, `auto quality`, `useNavComposeForms`
When (a) D에서 구성 편집 → Claude Code → 역할 그룹 → 설계·판단 `opus` 선택지(balanced의 `fable`과 다른 값) → 그 뒤 폼 모두 Enter → 적용할까요? `b` → 방식 역할 그룹 → 모든 폼 Enter → 적용할까요? 적용 → 유지
When (b) P에서 구성 편집 → 모든 도구 같게 → 역할 그룹 → 설계·판단 `opus` 선택지 → 그 뒤 폼 모두 Enter → override 질문 `b` → 방식 역할 그룹 → 모든 폼 Enter → override 질문 `n` → 적용 → 유지
Then (a) `어떻게 고칠까요?`와 `적용할까요?`는 2번씩이고 확인의 취소 줄(`composeText.Cancelled`)은 0번이며 NPD는 대조와 같다
And (b) NPD는 override 질문에 처음부터 `n`을 준 대조와 같다

### S13: OMP 탭과 리뷰 탭은 탭 단위로 돌아간다
Priority: Must
Given O3·O3N(activate 호출 수), RV, ko, `auto quality`, `useNavComposeForms`
When (a) O3에서 구성 편집 → OMP → 첫 폼 `b` (b) O3N에서 구성 편집 → OMP → `OMP 구성` 폼의 4번 `에이전트별 지정 (balanced 위에)`(기본값 1번 `돌아가기`가 아닌 선택지) → `모델 계열` Enter(저장된 계열) → `에이전트를 고르세요` 목록의 `task — 프로필 기본값` 선택지(기본값 `완료`가 아닌 선택지) → `task (<capability>)` model 폼 `b` (b2) (b)와 같되 목록에서 `b` (c) RV에서 구성 편집 → 멀티프로바이더 리뷰 → 첫 폼 `b` (d) RV에서 첫 폼 Enter → 둘째 폼 `b`를 주고, 각각 도구 탭 `b` → 유지를 고른다
Then 다섯 run 모두 출력에 `경고:`가 없고 `어느 도구를 바꿀까요?`가 2번이며 다시 보인 도구 탭의 기본 선택은 그 탭이다
And autopus.yaml SHA-256, `.omp/` 아래 파일 bytes가 실행 전과 같고 activate 호출은 0번이다
And (b)의 출력에는 `에이전트를 고르세요` 목록과 `task (`로 시작하는 model 폼 제목이 차례로 있고, 그 뒤 바로 도구 탭 `어느 도구를 바꿀까요?`가 온다. 이 `b`는 `ompPin` → `ompPins` → `ompFailed`(`quality_compose_omp.go:93`)를 지난다. (b2)의 `b`도 `ompPins` → `ompFailed`를 지난다. 두 경우 `ompFailed`의 조건이 넓혀지지 않았으면 `  경고:` 줄이 나오고 도구 탭이 다시 나오지 않으므로 이 단언이 그 경로를 구별한다

### S14: 메뉴의 경계와 init 구성 단계 밖으로 뒤로
Priority: Must
Given D(quality·update), IR, ko, `useNavComposeForms`
When (a) D quality 메뉴에 `b` → 유지 (b) D update 메뉴에 `b` → Enter (c) IR에서 첫 진입 CAT-A, 둘째 진입 CAT-B로, 첫 진입에서 구성 편집 → Claude Code → 역할 그룹 → 설계·판단 Enter(그대로) → 설계·판단 effort `low` 선택지 → 그 뒤 폼 모두 Enter → 적용 → 메뉴 `b`를 고르고, seam의 resume 호출은 W_rg를 돌려주며, 새로 시작한 둘째 진입에서 구성 편집 → Codex → 역할 그룹 → 설계·판단 `opus`의 CAT-B 선택지 → 그 뒤 폼 모두 Enter → 적용 → 유지, 확인 화면 Enter (c′) 음성 대조로 (c)와 같은 입력을 두 진입 모두 CAT-A로 돌린다
Then (a) 안내 줄이 1번이고 구성 표는 1번만 그려지며 SHA-256이 같다
And (b) 출력은 `b` 없이 Enter만 준 대조 출력에 안내 한 줄을 더한 것과 같다
And (c) 폐기: seam의 둘째 호출 `Resume`은 W_rg와 같다. 첫 진입의 적용 뒤 다시 그린 표는 Claude Code 열이 `Claude Code (개별 설정: claude-custom)`이고 설계·판단 칸이 `fable · low`지만, 둘째 진입의 첫 표에는 `claude-custom`이 없고 그 칸이 `fable-5-1 · max`다(다시 그린 표에는 `*` 표시가 없어 T9에서 코드 사실로 고쳤다). autopus.yaml에 `claude-custom`과 `efforts` key가 하나도 없고 `quality.providers`는 `codex: codex-custom` 하나다. 확인 블록에 `claude-custom`·`efforts`가 든 줄이 없고 `  quality.providers: codex=codex-custom`이 있다. NPD는 seam이 W_rg를 한 번 돌려주고 둘째 진입의 답만 CAT-B 한 번 진입으로 준 대조 X와 같다
And (c) 마지막 snapshot: Codex planner agent 파일의 model은 `gpt-6.1-sol`이 아니고 둘째 진입 표의 Codex planner 칸과 같다. (c′)의 같은 파일 model은 `gpt-6.1-sol`이다

### S15: 폼이 열리지 않는 자리는 bytes가 같다
Priority: Must
Given base golden, 폼 실행을 세는 runner
When (a) `auto init --yes --dir <DIR> --platforms claude-code,codex` (b) 그 project의 `auto update --yes` (c) `CI=true`인 `auto update`(TTY probe true) (d) `cmd.SetIn`으로 `b\n`을 준 D의 `auto quality` 줄 driver (e) TTY probe false인 `auto quality` (f) `CI=true`이고 forms probe true인 D `auto quality`에 메뉴 Enter (g) TTY probe true, forms probe false(pipe stdout)인 IN (h) `CI=true`인 IN을 돌리고, 기존 줄 driver golden test를 돌린다
Then (a)–(e)와 (g)의 stdout은 base golden과 bytes가 같고 NPD가 같으며 폼 실행은 (a)–(e)에서 0번, (g)에서 0번(wizard seam만 불린다)이다
And (a)–(e), (g)의 출력에 `지금까지 고른 값`, `Chosen so far`, `방향키 이동`, `Arrows move`, `뒤로: b`, `back: b`, ` · 모델 구성`, ` · Model composition`, `적용할 설정`이 없다
And (f)는 `CI` 없이 같은 입력을 준 run과 출력이 같고 header `quality · 모델 구성`이 있다(quality의 gate는 CI를 보지 않는다). 다른 줄은 리뷰 탭 자신의 CI 규칙(SPEC-REVMODEL-001 REQ-RVM-14)이 지우는 리뷰 요약 줄 `멀티프로바이더 리뷰: …` 하나뿐이다(T10에서 코드 사실로 고쳤다). (h)는 wizard seam이 불리고 `init · 확인`이 있다(init의 gate는 CI를 보지 않는다)
And (e)의 오류는 `interactive quality selection requires a TTY; use an explicit quality or profile command`이고 기존 줄 driver golden test는 stub 시그니처 줄 말고는 수정 없이 통과한다

### S16: Ctrl+C는 그대로다
Priority: Must
Given SPEC-MODELUX-002 S12의 기존 test, WZ6 wizard에서 3단계에 `huh.ErrUserAborted`를 돌려주는 runner, S4의 `RunNav`
When 기존 test를 돌리고 wizard를 그 runner로 돌린다
Then 기존 test는 분류 1 밖의 수정 없이 통과한다(update는 중단 줄과 0이 아닌 종료 코드, init은 `  init cancelled.`과 파일 0, quality는 오류 없음)
And wizard 결과는 `Cancelled: true`이고 S4 (2)는 `huh.ErrUserAborted`다

### S17: 기존 test는 분류 안에서만 바뀐다
Priority: Must
Given 8ec13935의 test 파일과 research `### 기존 test 분류`
When `go test ./internal/cli/... ./pkg/...`를 필터 없이 돌리고 `git diff 8ec13935 -- '*_test.go'`를 본다
Then 모든 test가 통과한다
And 기존 test 파일의 변경 줄은 셋뿐이다: (i) 분류 1의 다섯 줄(stub 4곳의 시그니처, `quality_compose_form_test.go:94`의 호출 인자), (ii) 그 다섯 파일마다 `github.com/insajin/autopus-adk/internal/cli/tui` import 한 줄과 gofmt가 그 import block에 낸 정렬 변경, (iii) 분류표에 적힌 입력 줄. 새 helper와 새 test는 새 파일에만 있다
And 단언의 기대 값은 그대로다. `quality_compose_form_test.go:94`는 `require.NoError(t, composeRunForm(cmd, form, tui.NavFrame{Accessible: true}))`로 호출 인자만 하나 늘어난 줄로 허용되고, 그 test의 `value == "b"`와 `2. b` 단언 줄은 바뀌지 않은 채 통과한다. (i)–(iii) 밖에서 `assert.`·`require.` 줄이 바뀐 곳은 0이다

### S18: 실제 터미널 smoke
Priority: Must
Given 실제 터미널(ko locale, 24줄 이상), 빈 디렉터리, `make build`로 만든 바이너리
When `auto init --platforms claude-code,codex`에서 1단계 Esc, 3단계 Esc, 역할별 직접 구성, 구성 단계의 Claude Code 에이전트별 편집에서 에이전트 열 개 이상 답한 뒤 Esc 두 번, 적용, 유지, 확인 화면 뒤로, 유지, 취소를 하고, 다시 init을 적용까지 한 뒤 그 project에서 `auto update`의 메뉴 Esc, 구성 편집 → Codex → Esc, Ctrl+C를 한다
Then 기록에 1단계 안내 문장, 3단계에서 돌아간 2단계의 미리 선택된 언어, Contract 그대로의 header·요약·footer, 줄인 요약의 `외 k개` 줄과 화면 맨 위의 header, 확인 블록의 세 묶음, 취소 뒤 `ls -A`의 빈 출력, 적용 뒤 생성 파일, update Ctrl+C 뒤 `echo $?`의 0이 아닌 값이 있다
And 기록은 `evidence/manual-smoke-tuiux.txt`에 경로와 사용자 이름을 바꿔 남는다

### S19: 문서
Priority: Should
Given CLI help, README, CHANGELOG
When `auto init --help`, `auto update --help`, `auto quality --help`를 실행하고 문서를 읽는다
Then 세 help와 README, v0.50.126 CHANGELOG에 Esc 뒤로, accessible `b`, 첫 단계 안내, header·요약·footer, init 확인 화면(적용·뒤로·취소, 세 묶음, 확인 화면에 없는 것)이 있고 `quality_compose_docs_test.go`가 고정한 문장이 그대로다

### S20: 터미널 높이
Priority: Must
Given `NavFrame{Header: "h", Summary: [머리 줄 + "  a1: v" … "  a30: v"], Lang}`(ko와 en), Select(제목 + 선택지 3개)와 Select(선택지 18개) 폼
When `RunNav`를 `tea.WindowSizeMsg{Width: 80, Height: 24}`로 3개짜리 폼에, `Height: 10`으로 18개짜리 폼에 돌려 view(ANSI 제거)를 읽고, 같은 frame을 `RunNavAccessible`로 돈다
Then 높이 24 view는 24줄 이하이고 첫 줄 `h`, 마지막 줄이 footer다. test가 잰 form 높이 h로 avail = 22 − h일 때 요약은 머리 줄, ko `  … 외 <k>개`(en `  … <k> more`, k = 30 − (avail − 2)), 마지막 avail − 2개 답이고 마지막 답 줄은 `  a30: v`다
And 높이 10 view는 10줄 이하이고 첫 줄 `h`, 마지막 줄이 footer이며 요약 머리 줄이 없다(avail < 3). form은 8줄 이하다
And accessible 출력에는 답 30줄이 모두 있고 `외`·`more` 줄이 없다

### S21: TERM=dumb의 accessible 경로
Priority: Must
Given `TERM=dumb`과 `TERM=xterm-256color`, frame을 기록하는 `composeRunForm` stub, D `auto quality`(forms probe true), Select(A, B, C, 값 b)
When 각 TERM에서 D quality를 메뉴 Enter로 돌리고, `RunNav(form, NavFrame{Header: "h", Lang: "ko", Accessible: true}, in, out)`를 `b\n`과 `3\n`으로 돌리며, wizard 기본 runner의 stream 선택을 accessible true·false로 묻는다
Then `TERM=dumb`에서 기록된 frame은 모두 `Accessible: true`, `xterm-256color`에서는 모두 false다
And `RunNav`는 `b\n`에 `tui.ErrNavBack`, `3\n`에 nil과 값 c를 돌려주고, out에는 `h`, `답을 입력하고 Enter를 누르세요 · 뒤로: b`, `1. A`가 있으며 `\x1b` byte가 없다(tea.Program을 돌리지 않았다)
And wizard 기본 stream은 accessible이면 (os.Stdin, os.Stdout), 아니면 (os.Stdin, os.Stderr)다

## Oracle Acceptance Notes

- 모든 Must 시나리오는 concrete expected output을 갖는다. header·footer·안내·요약·확인 블록 줄의 정확한 문자열과 순서, 결과 struct field 값, 오류 값(`tui.ErrNavBack`, `huh.ErrUserAborted`), 폼 제목 개수, seam 호출 수와 `Resume` 값, autopus.yaml key 값, view 줄 수와 `외 k개`의 k, NPD, 경로별 SHA-256·mtime이다. 파일 존재, heading, 종료 상태만으로 닫는 Must 시나리오는 없다.
- 이동이 값을 바꾸지 않는다는 oracle은 대조 run의 NPD다(S1, S7, S10–S12, S14). 대조는 같은 test binary에서 뒤로와 다시 답하기만 뺀 입력으로 돈다. 폼이 열리지 않는 자리의 oracle은 8ec13935 base golden이다(S6, S15). NPD는 시간 field만 정규화하므로 다른 차이는 모두 드러난다(F-004).
- 구성 상태의 보존과 폐기는 따로 판정한다. 보존(S7 (b))은 확인 화면에서 구성 메뉴로 돌아가 이어 한 run이 effort `low`를 남기는지, 폐기(S14 (c))는 구성 단계를 지나 wizard로 나간 run이 `low`와 `claude-custom`을 남기지 않는지다. 둘 다 시작값과 다른 값(`opus`, `low`)을 실제로 적용한 뒤 판정한다.
- 마지막 snapshot oracle(S7 (b), S14 (c))은 CAT-A에만 있는 `gpt-6.1-sol`이 생성 파일에 없는지와 CAT-B 대조의 NPD다. 같은 입력을 CAT-A로만 돈 음성 대조((b′), (c′))가 `gpt-6.1-sol`을 보여 oracle이 두 snapshot을 구별함을 확인한다. 첫 snapshot을 다시 쓰는 구현은 두 단언에서 모두 실패한다.
- 기대 값은 2026-10-16에 코드를 읽고 정했다(`wizard_steps.go`, `wizard_rea_step.go`, `quality_compose_form*.go`, `quality_compose_groups.go`, `init.go`, `init_rea.go`, `rea_run.go`, `pkg/config/defaults.go`, `schema.go`, `schema_spec.go`). 실행으로 확인한 것은 probe A1·A2(`evidence/probe-nav-a1.md`)뿐이다. Claude Code·Codex 탭에서 등급 뒤 effort 폼이 나오고 `low`가 자동 단계가 아니라는 S7·S10·S11·S14의 전제는 SPEC-MODELUX-003 REQ-MUX3-03에서 온다. balanced의 설계·판단 넷은 `fable`이므로(`role_model_policy_balanced.go:24-31`) 그 그룹의 `fable` 답만으로는 바뀌는 값이 없다. 코드 사실과 다르면 test가 아니라 이 SPEC의 기대를 고친다.
- S4·S20은 byte를 한 write에 모아 쓴다. byte마다 나눠 쓰면 화살표의 `\x1b`가 Esc로 읽힌다(probe P2, Open Risk R1). 이 동작은 oracle이 아니라 위험으로 둔다.
- accessible 출력은 입력을 echo하지 않는다. 줄 순서 판정은 header 줄 index 기준으로 한다. huh accessible 제목 줄은 끝에 공백 하나가 붙는다(probe P3 출력 `Q \n`).
