package orchestra

import (
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

// writeExecutableFixture writes a script that the calling test will exec,
// holding syscall.ForkLock for reading while the write descriptor is open.
//
// On Linux, execve fails with ETXTBSY while any process holds the file open
// for writing. Parallel sibling tests fork real providers (cat, echo, sh, and
// other fixtures) through os/exec; a child forked inside this helper's
// open-write-close window inherits the descriptor and keeps it until its own
// execve, which can land after this test's exec of the fixture. Go takes
// ForkLock for writing around clone and, because Linux clones with
// CLONE_VFORK, does not release it until that child has exec'd, so a read
// lock across the write makes the two windows mutually exclusive
// (golang/go#22315). The write itself never takes ForkLock, so there is no
// recursive read lock.
//
// Without this, the race surfaced only as "모든 프로바이더가 응답하지 않았습니다"
// from runFastest (CI run 40431690478, attempt 1).
func writeExecutableFixture(t *testing.T, path, script string, perm os.FileMode) {
	t.Helper()
	syscall.ForkLock.RLock()
	err := os.WriteFile(path, []byte(script), perm)
	syscall.ForkLock.RUnlock()
	require.NoError(t, err)
}
