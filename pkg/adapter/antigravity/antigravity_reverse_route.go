package antigravity

import (
	"strings"

	"github.com/insajin/autopus-adk/templates"
)

// reverseContractMarker is the line of the auto-reverse skill template that the
// shared /auto reverse contract replaces. The template engine has no include,
// so the adapter splices the one shared body in after rendering instead of the
// Gemini template keeping a second copy of the contract.
const reverseContractMarker = "<!-- autopus:auto-reverse-contract -->"

func expandReverseRouteContract(rendered string) string {
	if !strings.Contains(rendered, reverseContractMarker) {
		return rendered
	}
	return strings.Replace(rendered, reverseContractMarker, templates.ReverseRouteContract(), 1)
}
