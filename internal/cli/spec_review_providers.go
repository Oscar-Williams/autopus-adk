package cli

import (
	"fmt"
	"os"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/detect"
	"github.com/insajin/autopus-adk/pkg/orchestra"
)

// filterInstalledProviders keeps the providers whose binary is installed and
// warns on stderr about every other one, in the runtime wording of auto spec
// review and the doctor smoke.
func filterInstalledProviders(all []orchestra.ProviderConfig) []orchestra.ProviderConfig {
	return filterInstalledProvidersWith(all, warnMissingProviderBinary)
}

// filterInstalledProvidersWith keeps the providers whose binary is installed
// and hands every other one, with the binary it looked for, to missing.
func filterInstalledProvidersWith(all []orchestra.ProviderConfig, missing func(orchestra.ProviderConfig, string)) []orchestra.ProviderConfig {
	var available []orchestra.ProviderConfig
	for _, provider := range all {
		binary := provider.Binary
		if provider.Backend == config.ProviderBackendOMP {
			binary = config.ProviderBackendOMP
		}
		if detect.IsInstalled(binary) {
			available = append(available, provider)
			continue
		}
		missing(provider, binary)
	}
	return available
}

// warnMissingProviderBinary is the runtime warning. The binary comes from
// autopus.yaml and reaches the terminal, so it passes through terminalSafe
// (SPEC-REVMODEL-001 REQ-RVM-20).
func warnMissingProviderBinary(_ orchestra.ProviderConfig, binary string) {
	fmt.Fprintf(os.Stderr, "경고: %s 바이너리를 찾을 수 없습니다 (건너뜀)\n", terminalSafe(binary))
}

func configureSpecReviewProviders(providers []orchestra.ProviderConfig) []orchestra.ProviderConfig {
	configured := append([]orchestra.ProviderConfig(nil), providers...)
	for i := range configured {
		configured[i].ResultReadyPatterns = mergeStringValues(configured[i].ResultReadyPatterns, []string{"VERDICT:"})
		if configured[i].ResultReadyGrace <= 0 {
			configured[i].ResultReadyGrace = specReviewResultReadyGrace
		}
	}
	return configured
}

func resolveSpecReviewProviderNames(cfg *config.HarnessConfig, multi bool) []string {
	if cfg == nil {
		return nil
	}
	names := mergeProviderNames(cfg.Spec.ReviewGate.Providers)
	if !multi {
		return names
	}
	if command, ok := cfg.Orchestra.Commands["review"]; ok {
		names = mergeProviderNames(names, command.Providers)
	}
	return mergeProviderNames(names, sortedProviderKeys(cfg.Orchestra.Providers), defaultProviders())
}
