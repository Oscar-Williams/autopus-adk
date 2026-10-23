package main

import (
	"strings"
	"testing"
)

func laneManifests(lane string, total, tests int) []manifest {
	out := make([]manifest, total)
	for i := range out {
		assigned := tests / total
		if i == 0 {
			assigned += tests % total
		}
		out[i] = manifest{
			Lane: lane, Shard: i, Total: total, Tests: tests,
			Assigned: assigned, Effective: assigned, Universe: "abc123",
		}
	}
	return out
}

func TestManifest_RoundTrips(t *testing.T) {
	want := manifest{Lane: "unit", Shard: 2, Total: 5, Tests: 3259, Assigned: 652, Effective: 641, Universe: "deadbeef"}
	got, err := parseManifest(strings.NewReader(want.String()))
	if err != nil {
		t.Fatalf("parseManifest: %v", err)
	}
	if got != want {
		t.Fatalf("round trip = %+v, want %+v", got, want)
	}
}

func TestParseManifest_RejectsEvidenceThatIsNotComplete(t *testing.T) {
	full := manifest{Lane: "unit", Shard: 1, Total: 3, Tests: 9, Assigned: 3, Effective: 3, Universe: "u"}.String()
	for name, input := range map[string]string{
		"missing key":   strings.Replace(full, "universe=u\n", "", 1),
		"repeated key":  full + "shard=1\n",
		"unknown key":   full + "extra=1\n",
		"not a number":  strings.Replace(full, "total=3", "total=three", 1),
		"negative":      strings.Replace(full, "assigned=3", "assigned=-3", 1),
		"not key=value": full + "garbage\n",
		"empty":         "",
	} {
		if got, err := parseManifest(strings.NewReader(input)); err == nil {
			t.Errorf("%s: parseManifest accepted the input and returned %+v", name, got)
		}
	}
}

func TestVerifyManifests_AcceptsAnExactCoverOfEveryRequiredLane(t *testing.T) {
	all := append(laneManifests("unit", 5, 3259), laneManifests("integration", 5, 3270)...)
	if err := verifyManifests(all, []string{"unit", "integration"}); err != nil {
		t.Fatalf("verifyManifests: %v", err)
	}
}

func TestVerifyManifests_RejectsAnythingShortOfAnExactCover(t *testing.T) {
	good := func() []manifest { return laneManifests("unit", 4, 100) }
	missingShard := good()[:3]
	duplicated := append(good()[:3], good()[0])
	otherUniverse := good()
	otherUniverse[2].Universe = "different"
	otherTotal := good()
	otherTotal[1].Total = 5
	lostTests := good()
	lostTests[3].Assigned--
	lostTests[3].Effective--
	nothingToRun := good()
	nothingToRun[1].Effective = 0
	outOfRange := good()
	outOfRange[3].Shard = 9

	for name, input := range map[string][]manifest{
		"missing shard":        missingShard,
		"duplicated shard":     duplicated,
		"different universe":   otherUniverse,
		"different total":      otherTotal,
		"tests lost":           lostTests,
		"shard runs nothing":   nothingToRun,
		"shard index too high": outOfRange,
	} {
		if err := verifyManifests(input, []string{"unit"}); err == nil {
			t.Errorf("%s: verifyManifests accepted the manifests", name)
		}
	}
	if err := verifyManifests(good(), []string{"unit", "integration"}); err == nil ||
		!strings.Contains(err.Error(), `"integration"`) {
		t.Errorf("a lane that stopped reporting was accepted: %v", err)
	}
	if err := verifyManifests(nil, nil); err != nil {
		t.Errorf("no lanes and none required should pass vacuously, got %v", err)
	}
}
