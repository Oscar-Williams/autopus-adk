package config

// ReaFeaturesConf records the project's REA MCP decision (SPEC-REA-001).
// Enabled is a pointer so a stored false survives a load/save round trip and
// an absent key stays absent: the key is written only when the managed MCP
// entry cannot carry the decision itself (Decision State Contract).
type ReaFeaturesConf struct {
	Enabled *bool `yaml:"enabled,omitempty"`
}

// Decided reports whether autopus.yaml stores an explicit REA decision.
func (c ReaFeaturesConf) Decided() bool { return c.Enabled != nil }

// ReaMCPAction tells a platform adapter what to do with the managed rea MCP
// entry during one run. The zero value keeps the file as it is, so any
// generation path that does not resolve the REA decision neither adds nor
// removes the entry.
type ReaMCPAction string

const (
	ReaMCPKeep   ReaMCPAction = ""
	ReaMCPAdd    ReaMCPAction = "add"
	ReaMCPRemove ReaMCPAction = "remove"
)
