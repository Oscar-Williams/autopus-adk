package cli

import (
	"errors"
	"fmt"

	"github.com/insajin/autopus-adk/internal/cli/tui"
	"github.com/insajin/autopus-adk/pkg/config"
)

// initReviewChoice is the operator's answer on the init review.
type initReviewChoice int

const (
	initReviewApply initReviewChoice = iota
	initReviewBack
	initReviewCancel
)

// initReviewText is the init review's text in one locale (Init Review
// Contract). Setting keys and values stay as config.Save writes them.
type initReviewText struct {
	Title, Answers, Changed, Rea, None string
	ReaLine                            string // %s · %s: the decision and the action
	On, Off, Add, Keep, Remove         string
	Header, Question                   string
	Apply, Back, Cancel                string
}

var initReviewTextKO = initReviewText{
	Title: "적용할 설정", Answers: " 답한 값", Changed: " 바뀌는 설정", Rea: " REA", None: "(없음)",
	ReaLine: "  REA MCP: %s · %s", On: "켬", Off: "끔", Add: "등록", Keep: "그대로", Remove: "제거",
	Header: "확인", Question: "이대로 진행할까요?", Apply: "적용", Back: "뒤로", Cancel: "취소",
}

var initReviewTextEN = initReviewText{
	Title: "Settings to apply", Answers: " Answers", Changed: " Changed settings", Rea: " REA", None: "(none)",
	ReaLine: "  REA MCP: %s · %s", On: "on", Off: "off", Add: "add", Keep: "keep", Remove: "remove",
	Header: "Review", Question: "Proceed with these settings?", Apply: "apply", Back: "back", Cancel: "cancel",
}

// initReviewTextFor returns Korean for ko and English for every other
// locale, as composeFormTextFor does.
func initReviewTextFor(lang string) initReviewText {
	if lang == "ko" {
		return initReviewTextKO
	}
	return initReviewTextEN
}

// review writes the init review and asks 적용, 뒤로, or 취소, whose default
// is 적용 (REQ-TUX-11). Its question is not the composition step's
// 적용할까요?, so the two confirmations read apart. The block is plain lines before the form, as the
// composition table is: the answers of the steps the wizard asked, every
// setting key the run changes from the base config, and the run's REA
// decision with the action the platform writes take. Esc is 뒤로 and Ctrl+C
// is 취소; any other form error ends init with nothing written.
func (r *initInteractive) review(result *tui.InitWizardResult) (initReviewChoice, error) {
	lang := composeLanguage()
	text := initReviewTextFor(lang)
	changes, err := initReviewChanges(r.base, r.cfg, text.None)
	if err != nil {
		return initReviewCancel, err
	}
	out := r.cmd.OutOrStdout()
	fmt.Fprintln(out, text.Title)
	if answers := tui.WizardAnswerLines(r.opts, result, lang); len(answers) > 0 {
		fmt.Fprintln(out, text.Answers)
		for _, line := range answers {
			fmt.Fprintln(out, terminalSafe(line))
		}
	}
	fmt.Fprintln(out, text.Changed)
	if len(changes) == 0 {
		changes = []string{"  " + text.None}
	}
	for _, line := range changes {
		fmt.Fprintln(out, line)
	}
	if r.rea.active {
		fmt.Fprintln(out, text.Rea)
		fmt.Fprintf(out, text.ReaLine+"\n", r.reaDecision(text), r.reaAction(text))
	}
	nav := &composeNav{command: r.cmd.Name(), lang: lang, title: text.Header, backOut: true}
	defer nav.attach(r.cmd)()
	choice, err := composeSelect(r.cmd, text.Question, "", []composeOption[initReviewChoice]{
		{text.Apply, initReviewApply}, {text.Back, initReviewBack}, {text.Cancel, initReviewCancel},
	}, initReviewApply)
	switch {
	case errors.Is(err, tui.ErrNavBack):
		return initReviewBack, nil
	case errors.Is(err, errComposeInterrupted):
		return initReviewCancel, nil
	case err != nil:
		return initReviewCancel, fmt.Errorf("init review: %w", err)
	}
	return choice, nil
}

// reaDecision is the run's REA decision, which decideInit settled.
func (r *initInteractive) reaDecision(text initReviewText) string {
	if r.rea.state == reaOff {
		return text.Off
	}
	return text.On
}

// reaAction is the action decide set for the platform writes.
func (r *initInteractive) reaAction(text initReviewText) string {
	switch r.cfg.Runtime.ReaMCP {
	case config.ReaMCPAdd:
		return text.Add
	case config.ReaMCPRemove:
		return text.Remove
	}
	return text.Keep
}
