package templates

// ReverseRouteContract returns the /auto reverse route contract that every
// platform's auto-reverse detail renders (SPEC-REA-001). The consent loop, the
// confirmation gate, and the clean-room handoff live in this one fragment so
// the five details cannot drift apart.
func ReverseRouteContract() string {
	return mustReadSharedFragment("shared/auto-reverse-contract.md.tmpl")
}
