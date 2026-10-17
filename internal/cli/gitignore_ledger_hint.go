package cli

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/insajin/autopus-adk/pkg/adapter/claude"
)

// permissionLedgerRel is the Claude permission ledger. The managed
// .gitignore ignores it, but a repository that committed it before keeps
// tracking it until someone untracks it.
const permissionLedgerRel = claude.PermissionLedgerPath

// permissionLedgerTrackedHint is the update line for a ledger git still
// tracks. Untracking stays the user's call: the commit that drops the file
// deletes it from every clone that pulls it, where the next update then
// starts a fresh ledger.
const permissionLedgerTrackedHint = "참고: " + permissionLedgerRel + "은 이제 .gitignore 대상인데 git이 아직 추적합니다. " +
	"로컬 파일은 두고 추적만 멈추려면 git rm --cached " + permissionLedgerRel

// hintTrackedPermissionLedger prints permissionLedgerTrackedHint when the
// git index of dir holds the ledger.
func hintTrackedPermissionLedger(out io.Writer, dir string) {
	if gitTracksPath(dir, permissionLedgerRel) {
		fmt.Fprintf(out, "  %s\n", permissionLedgerTrackedHint)
	}
}

// gitTracksPath reports whether the git index of dir holds rel. Any git
// failure, such as no git binary or a directory outside a repository, reads
// as untracked: there is no index to clean. core.fsmonitor=false keeps git
// from starting a command the repository's own config names.
func gitTracksPath(dir, rel string) bool {
	cmd := exec.Command("git", "-c", "core.fsmonitor=false", "ls-files", "-z", "--", rel)
	cmd.Dir = dir
	out, err := cmd.Output()
	return err == nil && len(out) > 0
}
