package cli

import "testing"

// The path rule is pure text, so both platform rules run on every host.
func TestClaudeVersionProbePathRuleOnWindows(t *testing.T) {
	allowed := []string{`C:\Users\me\.local\bin\claude.exe`, `C:/tools/claude.exe`, `d:\x\claude.exe`}
	refused := []string{
		`C:\x\claude`, // exec would try claude.com, claude.exe, claude.bat, claude.cmd
		`C:\x\claude.cmd`,
		`C:\x\claude.bat`,
		`C:\x\Claude.EXE`,
		`\\server\share\claude.exe`,
		`//server/share/claude.exe`,
		`\\?\C:\x\claude.exe`,
		`\\.\C:\x\claude.exe`,
		`\??\C:\x\claude.exe`,
		`C:claude.exe`,
		`\x\claude.exe`,
		`claude.exe`,
		`C:\x\claude.exe:stream`,
		`C:\x\claude.exe\`,
		``,
	}
	for _, value := range allowed {
		if !claudeVersionProbePathAllowed("windows", value) {
			t.Fatalf("%q refused", value)
		}
	}
	for _, value := range refused {
		if claudeVersionProbePathAllowed("windows", value) {
			t.Fatalf("%q allowed", value)
		}
	}
}

func TestClaudeVersionProbePathRuleElsewhere(t *testing.T) {
	for _, goos := range []string{"darwin", "linux"} {
		for _, value := range []string{"/opt/claude/bin/claude", "/x/claude.exe"} {
			if !claudeVersionProbePathAllowed(goos, value) {
				t.Fatalf("%s %q refused", goos, value)
			}
		}
		for _, value := range []string{"", "claude", "probe/claude", "/x/node", "/x/claude.sh", "/x/Claude"} {
			if claudeVersionProbePathAllowed(goos, value) {
				t.Fatalf("%s %q allowed", goos, value)
			}
		}
	}
}
