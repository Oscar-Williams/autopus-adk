package orchestra

import (
	"path/filepath"
	"runtime"
	"testing"
)

func typedJudgeProvider(t *testing.T, name string) ProviderConfig {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("typed judge fixture requires a POSIX shell")
	}
	dir := t.TempDir()
	binary := filepath.Join(dir, "typed-judge")
	script := "#!/bin/sh\ncat >/dev/null\nprintf '%s\\n' '{\"recommendation\":\"typed judge fixture\"}'\n"
	writeExecutableFixture(t, binary, script, 0o700)
	return ProviderConfig{Name: name, Binary: binary, ModelFamily: "judge-fixture"}
}
