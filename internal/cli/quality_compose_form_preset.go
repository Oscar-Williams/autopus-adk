package cli

import (
	"fmt"
	"strings"
)

// choosePreset asks quality.default among the configured presets, the
// current default first selected, and writes another choice through the
// line writer on a detached copy: cfg follows only a successful write. A
// default that names no configured preset, such as an empty one, gets a
// keep option of its own, so Enter never writes the first preset.
func (f *composeFormRun) choosePreset() error {
	current := f.Config.Quality.Default
	var options []composeOption[string]
	if _, configured := f.Config.Quality.Presets[current]; !configured {
		shown := current
		if shown == "" {
			shown = f.form.PresetNone
		}
		options = append(options, composeOption[string]{f.form.Keep + " — " + shown, current})
	}
	for _, name := range orderedQualityPresets(f.Config) {
		label := name
		if description := strings.TrimSpace(f.Config.Quality.Presets[name].Description); description != "" {
			label += " — " + description
		}
		if name == current {
			label += " " + f.form.PresetCurrent
		}
		options = append(options, composeOption[string]{label, name})
	}
	choice, err := composeSelect(f.cmd, f.form.Preset, "", options, current)
	if err != nil || choice == current {
		return err
	}
	plan := composeBindingPlan{Writer: composeWriterLine, Default: choice}
	if err := persistComposeBinding(f.shared.dir, f.shared.cfg, plan); err != nil {
		f.warn(err)
		return nil
	}
	f.applied = true
	fmt.Fprintf(f.out, "quality.default = %s\n", terminalSafe(choice))
	return nil
}
