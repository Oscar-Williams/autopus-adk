package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"testing"
	"time"
)

// The repository names are spelled out literally on purpose: they are what the
// published channel documents carry, so they must not come from the code under
// test.
const (
	previousOwnerRepository = "Insajin/autopus-adk"
	currentOwnerRepository  = "autopus-ai/autopus-adk"
)

// unacceptedRepositories are a third repository, a near-miss, and a wrong-case
// spelling of each accepted owner.
var unacceptedRepositories = []struct{ name, repository string }{
	{"third repository", "autopus-ai/autopus-desktop"},
	{"near-miss", "autopus-ai/autopus-adk-x"},
	{"wrong case of the current owner", "Autopus-AI/autopus-adk"},
	{"wrong case of the previous owner", "insajin/autopus-adk"},
}

// writeChannelFiles signs a channel document that names repository at version
// with the key that newCommandFixture trusts, and returns the paths of the
// document and its detached signature.
func writeChannelFiles(t *testing.T, directory, prefix, repository, version string) (string, string) {
	t.Helper()
	private := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{0x33}, ed25519.SeedSize))
	now := time.Now().UTC().Truncate(time.Second)
	document, err := json.Marshal(map[string]string{
		"schema_version": "adk-channel.v1",
		"channel":        "stable",
		"repository":     repository,
		"version":        version,
		"tag":            "v" + version,
		"issued_at":      now.Add(-time.Hour).Format(time.RFC3339),
		"expires_at":     now.Add(time.Hour).Format(time.RFC3339),
		"channel_key_id": "test-channel-key",
	})
	if err != nil {
		t.Fatal(err)
	}
	return writeCommandFile(t, directory, prefix+".json", document),
		writeCommandFile(t, directory, prefix+".json.sig", ed25519.Sign(private, document))
}

func verifyUpdateArguments(candidateDocument, candidateSignature, currentDocument, currentSignature string) []string {
	return []string{
		"verify-update",
		"--document", candidateDocument,
		"--signature", candidateSignature,
		"--current-document", currentDocument,
		"--current-signature", currentSignature,
	}
}

func TestRunVerifyUpdateAcceptsTheRenamedRepository(t *testing.T) {
	fixture := newCommandFixture(t)
	tests := []struct {
		name      string
		current   string
		candidate string
	}{
		{"previous owner document replaced by a current owner document", previousOwnerRepository, currentOwnerRepository},
		{"current owner document replaced by a current owner document", currentOwnerRepository, currentOwnerRepository},
		{"previous owner document replaced by a previous owner document", previousOwnerRepository, previousOwnerRepository},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			currentDocument, currentSignature := writeChannelFiles(t, directory, "current", test.current, "1.2.3")
			candidateDocument, candidateSignature := writeChannelFiles(t, directory, "candidate", test.candidate, "1.2.4")

			version, err := run(verifyUpdateArguments(candidateDocument, candidateSignature, currentDocument, currentSignature), fixture.getenv)
			if err != nil || version != "1.2.4" {
				t.Fatalf("verify-update = (%q, %v)", version, err)
			}
			// The workflow verifies the staged candidate again before it commits.
			version, err = run([]string{"verify", "--document", candidateDocument, "--signature", candidateSignature}, fixture.getenv)
			if err != nil || version != "1.2.4" {
				t.Fatalf("verify = (%q, %v)", version, err)
			}
		})
	}
}

func TestRunVerifyUpdateRejectsAnUnacceptedRepositoryOnEitherSide(t *testing.T) {
	fixture := newCommandFixture(t)
	for _, unaccepted := range unacceptedRepositories {
		repository := unaccepted.repository
		t.Run(unaccepted.name, func(t *testing.T) {
			directory := t.TempDir()
			goodCurrentDocument, goodCurrentSignature := writeChannelFiles(t, directory, "good-current", previousOwnerRepository, "1.2.3")
			goodCandidateDocument, goodCandidateSignature := writeChannelFiles(t, directory, "good-candidate", currentOwnerRepository, "1.2.4")
			badCurrentDocument, badCurrentSignature := writeChannelFiles(t, directory, "bad-current", repository, "1.2.3")
			badCandidateDocument, badCandidateSignature := writeChannelFiles(t, directory, "bad-candidate", repository, "1.2.4")

			if _, err := run(verifyUpdateArguments(badCandidateDocument, badCandidateSignature, goodCurrentDocument, goodCurrentSignature), fixture.getenv); err == nil {
				t.Fatal("verify-update with an unaccepted candidate repository error = nil")
			}
			if _, err := run(verifyUpdateArguments(goodCandidateDocument, goodCandidateSignature, badCurrentDocument, badCurrentSignature), fixture.getenv); err == nil {
				t.Fatal("verify-update with an unaccepted current repository error = nil")
			}
			if _, err := run([]string{"verify", "--document", badCandidateDocument, "--signature", badCandidateSignature}, fixture.getenv); err == nil {
				t.Fatal("verify with an unaccepted repository error = nil")
			}
		})
	}
}

func TestRunVerifyRotationAcceptsTheRenamedRepository(t *testing.T) {
	fixture := newRotationCommandFixtureFor(t, currentOwnerRepository)
	for _, command := range []string{"verify-rotation", "verify-rotation-historical"} {
		t.Run(command, func(t *testing.T) {
			arguments := fixture.arguments()
			arguments[0] = command
			output, err := run(arguments, fixture.getenv)
			if err != nil {
				t.Fatalf("run() error = %v", err)
			}
			var stdout bytes.Buffer
			writeCommandOutput(&stdout, command, output)
			if !bytes.Equal(stdout.Bytes(), fixture.document) {
				t.Fatalf("command output = %q", stdout.Bytes())
			}
		})
	}
}

func TestRunVerifyRotationRejectsAnUnacceptedRepository(t *testing.T) {
	for _, unaccepted := range unacceptedRepositories {
		fixture := newRotationCommandFixtureFor(t, unaccepted.repository)
		for _, command := range []string{"verify-rotation", "verify-rotation-historical"} {
			t.Run(command+" "+unaccepted.name, func(t *testing.T) {
				arguments := fixture.arguments()
				arguments[0] = command
				if _, err := run(arguments, fixture.getenv); err == nil {
					t.Fatal("run() error = nil")
				}
			})
		}
	}
}
