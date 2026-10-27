package adkchannel

const (
	// channelRepository is the repository identity after the move to the
	// autopus-ai organization.
	channelRepository = "autopus-ai/autopus-adk"
	// previousChannelRepository is the identity the repository had before that
	// move. The published channel document and the v0.50.109 rotation sidecar
	// were signed with it, and VerifyUpdate checks the published document
	// through the same pin as the candidate, so it has to stay accepted for as
	// long as a document that names it can be current.
	previousChannelRepository = "Insajin/autopus-adk"
)

// channelRepositoryAccepted reports whether name is a repository identity that
// a signed channel or rotation document may carry. It compares against the two
// exact strings above: no case folding, trimming, or prefix and suffix
// matching, so a near-miss such as "autopus-ai/autopus-adk-x" is refused. The
// Ed25519 signature remains the trust anchor; this check only binds a document
// to this repository.
func channelRepositoryAccepted(name string) bool {
	return name == channelRepository || name == previousChannelRepository
}
