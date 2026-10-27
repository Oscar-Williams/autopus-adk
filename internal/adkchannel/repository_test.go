package adkchannel

import (
	"strings"
	"testing"
	"time"
)

// The repository names are spelled out literally on purpose. They are the
// identities the receiver has to accept, so they must not be derived from the
// production constants under test.
const (
	previousOwnerRepository = "Insajin/autopus-adk"
	currentOwnerRepository  = "autopus-ai/autopus-adk"
)

type repositoryCase struct {
	name       string
	repository string
}

var acceptedRepositories = []repositoryCase{
	{"previous owner", previousOwnerRepository},
	{"current owner", currentOwnerRepository},
}

// unacceptedRepositories holds a third repository, near-misses of both accepted
// names, and spellings that differ from them only in case, framing, or
// whitespace. The match is exact, so every one of them must be refused.
var unacceptedRepositories = []repositoryCase{
	{"empty", ""},
	{"unrelated repository of the current owner", "autopus-ai/autopus-desktop"},
	{"unrelated repository of the previous owner", "Insajin/autopus-desktop"},
	{"third owner", "someone-else/autopus-adk"},
	{"current owner name suffix", "autopus-ai/autopus-adk-x"},
	{"current owner name prefix", "autopus-ai/x-autopus-adk"},
	{"current owner owner suffix", "autopus-ai-x/autopus-adk"},
	{"current owner owner prefix", "xautopus-ai/autopus-adk"},
	{"previous owner name suffix", "Insajin/autopus-adk-x"},
	{"previous owner owner suffix", "Insajin2/autopus-adk"},
	{"previous owner owner prefix", "xInsajin/autopus-adk"},
	{"lowercase previous owner", "insajin/autopus-adk"},
	{"uppercase previous repository", "INSAJIN/AUTOPUS-ADK"},
	{"capitalized current owner", "Autopus-AI/autopus-adk"},
	{"capitalized current name", "autopus-ai/Autopus-ADK"},
	{"uppercase current repository", "AUTOPUS-AI/AUTOPUS-ADK"},
	{"leading space", " autopus-ai/autopus-adk"},
	{"trailing space", "autopus-ai/autopus-adk "},
	{"trailing newline", "autopus-ai/autopus-adk\n"},
	{"trailing NUL", "autopus-ai/autopus-adk\x00"},
	{"git suffix", "autopus-ai/autopus-adk.git"},
	{"trailing slash", "autopus-ai/autopus-adk/"},
	{"host qualified", "github.com/autopus-ai/autopus-adk"},
	{"URL", "https://github.com/autopus-ai/autopus-adk"},
	{"both names joined by comma", "Insajin/autopus-adk,autopus-ai/autopus-adk"},
	{"both names joined by space", "Insajin/autopus-adk autopus-ai/autopus-adk"},
	{"owner only", "autopus-ai"},
	{"name only", "autopus-adk"},
	{"reversed", "autopus-adk/autopus-ai"},
	{"homoglyph", "autopus-ai/autopus-\u0430dk"}, // U+0430 CYRILLIC SMALL LETTER A
}

// updateRefusals is the smaller set used where the check is only wiring around
// the same predicate: a third owner, a near-miss, and a wrong-case spelling.
var updateRefusals = []repositoryCase{
	{"third owner", "someone-else/autopus-adk"},
	{"near-miss", "autopus-ai/autopus-adk-x"},
	{"wrong case", "insajin/autopus-adk"},
}

func signedRepositoryDocument(t *testing.T, fixture verifierFixture, repository, version string, expiresAt time.Time) SignedBytes {
	t.Helper()
	document := fixture.document()
	document.Repository = repository
	document.Version = version
	document.Tag = "v" + version
	document.ExpiresAt = expiresAt.Format(time.RFC3339)
	return signDocument(t, fixture.private, document)
}

func TestVerifyAcceptsEitherChannelRepositoryName(t *testing.T) {
	fixture := newVerifierFixture()
	for _, test := range acceptedRepositories {
		t.Run(test.name, func(t *testing.T) {
			document := fixture.document()
			document.Repository = test.repository
			verified, err := Verify(signDocument(t, fixture.private, document), fixture.options)
			if err != nil {
				t.Fatalf("Verify() error = %v", err)
			}
			if verified.Document.Repository != test.repository {
				t.Fatalf("repository = %q, want %q", verified.Document.Repository, test.repository)
			}
		})
	}
}

func TestVerifyRejectsEveryOtherChannelRepositoryName(t *testing.T) {
	fixture := newVerifierFixture()
	for _, test := range unacceptedRepositories {
		t.Run(test.name, func(t *testing.T) {
			document := fixture.document()
			document.Repository = test.repository
			_, err := Verify(signDocument(t, fixture.private, document), fixture.options)
			if err == nil || !strings.Contains(err.Error(), "repository must be") {
				t.Fatalf("Verify() error = %v, want a repository refusal", err)
			}
		})
	}
}

func TestVerifyUpdateAcceptsEveryPairingOfTheAcceptedNames(t *testing.T) {
	fixture := newVerifierFixture()
	tests := []struct {
		name      string
		current   string
		candidate string
	}{
		{"rename from the previous to the current owner", previousOwnerRepository, currentOwnerRepository},
		{"current owner on both sides", currentOwnerRepository, currentOwnerRepository},
		{"previous owner on both sides", previousOwnerRepository, previousOwnerRepository},
		{"rollback from the current to the previous owner", currentOwnerRepository, previousOwnerRepository},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			current := signedRepositoryDocument(t, fixture, test.current, "1.2.3", fixture.now.Add(time.Hour))
			candidate := signedRepositoryDocument(t, fixture, test.candidate, "1.2.4", fixture.now.Add(time.Hour))
			verified, err := VerifyUpdate(candidate, &current, fixture.options)
			if err != nil {
				t.Fatalf("VerifyUpdate() error = %v", err)
			}
			if verified.Document.Repository != test.candidate || verified.Document.Version != "1.2.4" {
				t.Fatalf("verified = (%q, %q)", verified.Document.Repository, verified.Document.Version)
			}
		})
	}
}

func TestVerifyUpdateKeepsMonotonicPolicyAcrossTheRename(t *testing.T) {
	fixture := newVerifierFixture()
	current := signedRepositoryDocument(t, fixture, previousOwnerRepository, "1.2.3", fixture.now.Add(time.Hour))
	tests := []struct {
		name      string
		version   string
		expiresAt time.Time
		wantError string
	}{
		{name: "same version republished with a later expiry", version: "1.2.3", expiresAt: fixture.now.Add(2 * time.Hour)},
		{name: "same version with the same expiry", version: "1.2.3", expiresAt: fixture.now.Add(time.Hour), wantError: "equal version must move expiry later"},
		{name: "lower version with a later expiry", version: "1.2.2", expiresAt: fixture.now.Add(2 * time.Hour), wantError: "lower than current"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := signedRepositoryDocument(t, fixture, currentOwnerRepository, test.version, test.expiresAt)
			_, err := VerifyUpdate(candidate, &current, fixture.options)
			if test.wantError == "" {
				if err != nil {
					t.Fatalf("VerifyUpdate() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("VerifyUpdate() error = %v, want %q", err, test.wantError)
			}
		})
	}
}

func TestVerifyUpdateReplacesAnExpiredPreviousOwnerDocument(t *testing.T) {
	fixture := newVerifierFixture()
	expired := fixture.document()
	expired.Repository = previousOwnerRepository
	expired.IssuedAt = fixture.now.Add(-2 * time.Hour).Format(time.RFC3339)
	expired.ExpiresAt = fixture.now.Add(-time.Hour).Format(time.RFC3339)
	current := signDocument(t, fixture.private, expired)
	candidate := signedRepositoryDocument(t, fixture, currentOwnerRepository, "1.2.4", fixture.now.Add(time.Hour))

	if _, err := VerifyUpdate(candidate, &current, fixture.options); err != nil {
		t.Fatalf("VerifyUpdate() error = %v", err)
	}
}

func TestVerifyUpdateRejectsAnUnacceptedRepositoryOnEitherSide(t *testing.T) {
	fixture := newVerifierFixture()
	// Each case varies one side only; the other side names an accepted
	// repository, so the refusal can only come from the side under test.
	validCurrent := signedRepositoryDocument(t, fixture, previousOwnerRepository, "1.2.3", fixture.now.Add(time.Hour))
	validCandidate := signedRepositoryDocument(t, fixture, previousOwnerRepository, "1.2.4", fixture.now.Add(time.Hour))
	for _, test := range updateRefusals {
		t.Run("candidate "+test.name, func(t *testing.T) {
			candidate := signedRepositoryDocument(t, fixture, test.repository, "1.2.4", fixture.now.Add(time.Hour))
			_, err := VerifyUpdate(candidate, &validCurrent, fixture.options)
			if err == nil || !strings.HasPrefix(err.Error(), "candidate: ") || !strings.Contains(err.Error(), "repository must be") {
				t.Fatalf("VerifyUpdate() error = %v, want a candidate repository refusal", err)
			}
		})
		t.Run("current "+test.name, func(t *testing.T) {
			current := signedRepositoryDocument(t, fixture, test.repository, "1.2.3", fixture.now.Add(time.Hour))
			_, err := VerifyUpdate(validCandidate, &current, fixture.options)
			if err == nil || !strings.HasPrefix(err.Error(), "current: ") || !strings.Contains(err.Error(), "repository must be") {
				t.Fatalf("VerifyUpdate() error = %v, want a current repository refusal", err)
			}
		})
		t.Run("expired current "+test.name, func(t *testing.T) {
			document := fixture.document()
			document.Repository = test.repository
			document.IssuedAt = fixture.now.Add(-2 * time.Hour).Format(time.RFC3339)
			document.ExpiresAt = fixture.now.Add(-time.Hour).Format(time.RFC3339)
			expired := signDocument(t, fixture.private, document)
			_, err := VerifyUpdate(validCandidate, &expired, fixture.options)
			if err == nil || !strings.HasPrefix(err.Error(), "current: ") || !strings.Contains(err.Error(), "repository must be") {
				t.Fatalf("VerifyUpdate() error = %v, want an expired current repository refusal", err)
			}
		})
	}
}

func TestVerifyRotationAcceptsEitherChannelRepositoryName(t *testing.T) {
	fixture := newRotationFixture(t)
	for _, test := range acceptedRepositories {
		t.Run(test.name, func(t *testing.T) {
			document := fixture.document
			document.Repository = test.repository
			signed := signRotationDocument(t, fixture.private, document)

			verified, err := VerifyRotation(signed, fixture.pins, fixture.options)
			if err != nil {
				t.Fatalf("VerifyRotation() error = %v", err)
			}
			if verified.Document.Repository != test.repository {
				t.Fatalf("repository = %q, want %q", verified.Document.Repository, test.repository)
			}

			historical := fixture.options
			historical.Now = nil
			verified, err = VerifyRotationHistorical(signed, fixture.pins, historical)
			if err != nil {
				t.Fatalf("VerifyRotationHistorical() error = %v", err)
			}
			if verified.Document.Repository != test.repository {
				t.Fatalf("historical repository = %q, want %q", verified.Document.Repository, test.repository)
			}
		})
	}
}

func TestVerifyRotationRejectsEveryOtherChannelRepositoryName(t *testing.T) {
	fixture := newRotationFixture(t)
	const want = "repository does not match"
	for _, test := range unacceptedRepositories {
		t.Run(test.name, func(t *testing.T) {
			document := fixture.document
			document.Repository = test.repository
			signed := signRotationDocument(t, fixture.private, document)

			if _, err := VerifyRotation(signed, fixture.pins, fixture.options); err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("VerifyRotation() error = %v, want %q", err, want)
			}
			historical := fixture.options
			historical.Now = nil
			if _, err := VerifyRotationHistorical(signed, fixture.pins, historical); err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("VerifyRotationHistorical() error = %v, want %q", err, want)
			}
		})
	}
}
