package claude

import "github.com/insajin/autopus-adk/templates"

// claudeReverseContract is the shared /auto reverse contract. Consent and the
// confirmation gate stay questions of the main session; the detail only tells
// it when to ask and what each dispatch may carry.
var claudeReverseContract = templates.ReverseRouteContract()
