package cli

// composeReviewText holds every operator-facing string of the review tab
// (SPEC-REVMODEL-001 Form Text) beside composeFormText. Like it, a locale
// changes only what the operator reads: provider names, model selectors, and
// every value written to autopus.yaml stay the same in every locale. A keep
// option is composeFormText.Keep, " — ", and the current value.
type composeReviewText struct {
	Tab                              string // the tool tab option, the after-error select, and the summary
	TakesPart                        string // %s: the provider
	On, Off, TurnOff, TurnOn         string
	Model                            string // %s: the provider
	NotInCatalog                     string
	Thinking                         string // %[1]s: the provider, %[2]s: the provider/model selector
	CLINote                          string // %s: the provider
	CatalogNote                      string
	Judge, NoJudge, JudgeOff         string // JudgeOff: %s, the judge
	ErrCount                         string // %d: the participating providers
	ErrJudge                         string // %s: the judge
	ErrModel                         string // %[1]s: the provider, %[2]s: the selector
	ErrThinking                      string // %[1]s: the provider, %[2]s: the level, %[3]s: the model
	Retry, Leave                     string
	Warn                             string // %d: the baseline participating providers
	CommandsNote, ModelNote          string
	Check, CheckSkip                 string
	CheckRun                         string // %[1]d: providers, %[2]d: the most model calls
	CheckLine                        string // %[1]s: the provider and its model, %[2]s: the verdict
	CheckPass, CheckWarn             string
	CheckFail                        string // %s: the detail
	CheckFailed                      string
	CheckRefused                     string // %s: the provider name, terminal-safe
	WriteRefused                     string // %s: the first refused key
	ReloadFailed                     string // %s: the cause
	ColProvider, ColReview, ColModel string
	CLIBackend                       string
	SummaryOff                       string // appended to a provider that takes no part
	SummaryJudge                     string // %s: the judge
}

var composeReviewTextKO = composeReviewText{
	Tab: "멀티프로바이더 리뷰", TakesPart: "%s 리뷰 참여",
	On: "켜짐", Off: "꺼짐", TurnOff: "끄기", TurnOn: "켜기",
	Model: "%s model", NotInCatalog: " (catalog에 없음)", Thinking: "%[1]s thinking (%[2]s)",
	CLINote: "%s는 CLI backend라 model을 바꾸지 않습니다: 증명된 model 목록은 OMP backend에만 있고 " +
		"CLI model은 args와 quality 정책이 정합니다",
	CatalogNote: "OMP catalog를 읽지 못해 model은 그대로입니다",
	Judge:       "judge", NoJudge: "judge 없음", JudgeOff: "judge %s가 꺼져 judge를 다시 고릅니다",
	ErrCount:    "리뷰에는 provider가 두 개 이상 필요합니다 (지금 %d개)",
	ErrJudge:    "judge %s는 리뷰에 참여하지 않는 provider라 고를 수 없습니다",
	ErrModel:    "%[1]s model %[2]s는 설치된 OMP catalog가 증명하지 않습니다",
	ErrThinking: "%[1]s thinking %[2]s는 %[3]s가 보고하지 않는 단계입니다",
	Retry:       "다시 고르기", Leave: "나가기",
	Warn:         "지금 리뷰 provider가 %d개뿐입니다. 두 개 이상 켜야 적용할 수 있습니다",
	CommandsNote: "꺼진 provider도 orchestra 명령(brainstorm·plan·review·secure)과 auto spec review --multi에서는 계속 돕니다",
	ModelNote:    "바뀐 model은 auto spec review와 orchestra 명령 모두에 쓰입니다",
	Check:        "확인 호출", CheckSkip: "확인 호출 없이 계속",
	CheckRun:  "확인 호출 — provider %[1]d개에 요청 1회씩 (재시도 포함 최대 %[2]d회 model 호출)",
	CheckLine: "확인 호출 %[1]s: %[2]s",
	CheckPass: "통과", CheckWarn: "응답에 확인 문자열이 없습니다", CheckFail: "실패 — %s",
	CheckFailed:  "확인 호출이 실패한 provider가 있습니다. 적용하면 그 provider는 리뷰에서 실패합니다.",
	CheckRefused: "확인 호출을 할 수 없습니다: provider 이름 %s에 터미널 제어 문자가 있습니다",
	WriteRefused: "autopus.yaml의 리뷰 설정을 안전하게 고칠 수 없어 아무것도 쓰지 않았습니다: %s",
	ReloadFailed: "리뷰 설정은 적용됐지만 다시 읽지 못했습니다: %s",
	ColProvider:  "provider", ColReview: "리뷰", ColModel: "model", CLIBackend: "CLI backend",
	SummaryOff: "(꺼짐)", SummaryJudge: "judge %s",
}

var composeReviewTextEN = composeReviewText{
	Tab: "Multi-provider review", TakesPart: "%s takes part in review",
	On: "on", Off: "off", TurnOff: "turn off", TurnOn: "turn on",
	Model: "%s model", NotInCatalog: " (not in catalog)", Thinking: "%[1]s thinking (%[2]s)",
	CLINote: "%s runs on a CLI backend, so its model stays: attested model lists exist only for " +
		"OMP backends, and a CLI model comes from its args and quality policy",
	CatalogNote: "The OMP catalog is unreadable; models stay unchanged",
	Judge:       "judge", NoJudge: "no judge", JudgeOff: "judge %s was turned off; choose a new judge",
	ErrCount:    "The review needs at least two providers (now %d)",
	ErrJudge:    "judge %s does not take part in the review and cannot be chosen",
	ErrModel:    "The installed OMP catalog does not attest %[1]s model %[2]s",
	ErrThinking: "%[3]s reports no thinking level %[2]s for %[1]s",
	Retry:       "choose again", Leave: "leave",
	Warn:         "Only %d review providers now; turn on at least two to apply",
	CommandsNote: "Turned-off providers still run in orchestra commands (brainstorm, plan, review, secure) and in auto spec review --multi",
	ModelNote:    "A changed model applies to auto spec review and to orchestra commands",
	Check:        "Live check", CheckSkip: "continue without a live check",
	CheckRun:  "live check — one request for each of %[1]d providers (at most %[2]d model calls with retries)",
	CheckLine: "live check %[1]s: %[2]s",
	CheckPass: "passed", CheckWarn: "the reply lacks the check token", CheckFail: "failed — %s",
	CheckFailed:  "A live check failed; if you apply, that provider keeps failing in reviews.",
	CheckRefused: "Cannot run the live check: provider name %s holds terminal control characters",
	WriteRefused: "The review settings in autopus.yaml cannot be edited safely; nothing was written: %s",
	ReloadFailed: "The review settings were applied but could not be re-read: %s",
	ColProvider:  "provider", ColReview: "review", ColModel: "model", CLIBackend: "CLI backend",
	SummaryOff: " (off)", SummaryJudge: "judge %s",
}

// composeReviewTextFor returns Korean for ko and English for every other
// locale, as composeFormTextFor does.
func composeReviewTextFor(lang string) composeReviewText {
	if lang == "ko" {
		return composeReviewTextKO
	}
	return composeReviewTextEN
}

// reviewText is the review tab's text in the run's locale.
func (f *composeFormRun) reviewText() composeReviewText {
	return composeReviewTextFor(f.shared.lang)
}
