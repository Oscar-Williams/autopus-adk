# Release v0.50.127 (A36)

Status: coordinate re-bound, not published. A36 is `v0.50.127`; nothing in
this change creates a tag, an evidence object, a release, or a tap commit.

> **PREDECESSOR ATTEMPT BURNED 2026-10-23.** `v0.50.126` is failed release
> history. The R2-signed tag object `3fed15013ae753af224b15b97efd28dfa7db3bcc`
> (commit `182627252570bd1903bb03ba1010e0ff6704219f`, tree
> `a7fce0cc580e25970bfd5d3f62393048e076de7f`) and the evidence tag
> `omp-context-evidence-v0.50.126` (object
> `6e4f7507650d5181e7172d80f72f435c918967e3`, commit
> `4787c4a5d54e741da4882b2e4c7be6a4698b6918`) exist and must never be moved,
> deleted, recreated, or reused. GitHub release `418394554` remains an
> unpublished draft with zero assets. Ruleset `25454106`
> (`autopus-v0.50.126-release-authority`) is sealed and deployment tag policy
> `63663329` (`v0.50.126`) stays as history.
>
> Nothing was published, so no consumer saw a partial release and the Homebrew
> tap never moved. What was lost is the coordinate. This runbook is the next
> attempt.

## What burned v0.50.126

Nothing in the source or the evidence failed. `release-prep.sh --apply` signed
fresh K3 evidence on omp/18.12.3 and pushed the tag, and release run
`39425488512` passed every CI, security, and `omp-production-evidence` job. The
protected `release` job then waited for the `adk-companion-release` environment
approval for more than two and a half days, and by the time it ran, both things
it carries forward from the tag run had expired:

| Time (UTC) | Event |
|---|---|
| 2026-10-20 12:49:16 | K3 attestation issued; `not_before` = `issued_at`, `expires_at` 2026-10-21 12:49:16 |
| 2026-10-20 12:52:04 | `omp-production-evidence` finishes and uploads `omp-context-evidence-v0.50.126` with `retention-days: 1` |
| 2026-10-20 13:32:08 | the tag CI finishes; the `release` job starts waiting for approval |
| 2026-10-21 12:49:16 | the attestation expires |
| about 2026-10-21 12:52 | the one-day artifact ages out (the run no longer lists it) |
| 2026-10-23 05:04:33 | approval lands and the `release` job starts |
| 2026-10-23 05:06:15 | `Download same-run production evidence` fails: `Artifact not found for name: omp-context-evidence-v0.50.126` |

A rerun cannot recover it. The artifact is gone, and even with it the
protected job re-verifies the evidence with `--mode active`, which requires
`now < expires_at` (`validAt` in
`pkg/promptlayer/omp_context_promotion_attestation_v2.go`); the signer and the
verifier both cap the attestation at 24 hours. Fresh evidence needs a fresh
evidence tag, and the evidence tag of a coordinate is immutable, so the
coordinate is burned.

The deadline is not a defect to fix in this change. It is the freshness bound
the active evidence lane is built on, and it is why the approval timing is now
a stop condition below.

## Coordinate move notes (A36 = v0.50.126 -> A36 = v0.50.127)

This is a same-phase re-binding, so it was done by hand the way the A35
`v0.50.124 -> v0.50.125` move was. `advance-release-coordinate.sh v0.50.126
A36 v0.50.127 A36` handles most of it: it measures `v0.50.126` as unpublished
and moves the history rows in place. Its receipt step, however, always
appends, so it would leave `v0.50.126 0.50.126 A36` beside the new row, and
`TestReleaseCoordinateTable_ProducerRowsMatchDeclaredPhases` would fail. A run
in a scratch copy produced output byte-identical to the manual move for every
other file the script owns: its eighteen targets, plus the phase-list file it
rewrites to the same `'A36'`.

- The burned tag is recorded as rejected: `burned_A36_tag_126` in both lineage
  contracts, `v0.50.126` forbidden in the A36 release workflow contract,
  `omp-context-evidence-v0.50.126` in the evidence-lane forbidden list, the
  `v0.50.126` deployment policy refused by the strict sealed verifier test, and
  the `v0.50.126` receipt row asserted absent. Like every burned coordinate, it
  never enters the phase table. The `v0.50.124` records are unchanged.
- The release trigger pin in `pkg/companionmanifest/release_signing_wiring_test.go`
  moved with everything else. Tag CI runs `go test ./...`, so a stale pin there
  burns the tag by itself.
- Phase labels, the A36 ancestor (`dd6d6ea2`, the A35 source commit), and
  every A35 predecessor pin are unchanged. The `validate-source.sh` case arm
  was swapped in place, so the file stays at 300 counted lines.

The A36 source commit is not pinned; `release-prep.sh --apply` binds it to
`origin/main` at publication. `origin/main` has moved since `v0.50.126` was
tagged, most visibly to Go `1.27.3` (`expected_go_toolchain='go1.27.3'`), so
`v0.50.127` is not a rebuild of the burned tree.

## A35 predecessor pins, re-checked 2026-10-23

Nothing was published, so the pins measured on 2026-10-19 (see
[release-v0.50.126.md](release-v0.50.126.md)) still apply. Read only:

| Pin | Value |
|---|---|
| Release | `410986343`, `draft=false`, `immutable=true`, 15 assets, author `204883817`, still Latest |
| Tag object | `0616d14060c9bbd53d5196faeaa0bc8e108edb2b`, pointing at `dd6d6ea2435d759144ec093b9b464a831f47703d` |
| Source tree | `8d6aa0c4cf84f2c806205d3210505d100b863233` |
| `checksums.txt` and the four darwin/linux archives | API digests equal the `A35_*` pins in `verify-public-key-lineage-coordinates.sh` |
| Homebrew tap | `verify-homebrew-tap-pins.sh`: match (head `37fec98d1155e061e8f31f3efed05b8455fd5a90`, Cask `aa3898fbfbe434ef9777a2db7af39811ed1fbf2a`) |

The darwin manifest pins were not re-measured: they live inside archives whose
digests are unchanged.

## Before `--apply`

| Item | State on 2026-10-23 |
|---|---|
| Ruleset `autopus-v0.50.127-release-authority` | absent; create it armed |
| Deployment tag policy `v0.50.127` on `adk-companion-release` | absent; create it |
| `v0.50.127` and `omp-context-evidence-v0.50.127` | absent locally and on origin |
| Draft release named `v0.50.127` | absent |

Create both before `--preflight`. Create the ruleset in the exact shape
`verify-release-tag-ruleset.sh --armed` accepts: target `tag`, enforcement
`active`, `ref_name.include == ["refs/tags/v0.50.127"]`, empty exclude, rules
`creation`, `update`, `deletion`, and the single bypass actor
`{actor_id: 204883817, actor_type: "User", bypass_mode: "always"}`. Prep seals
it right after the tag exists, as it sealed `25454106` for `v0.50.126`. Never
widen an older ruleset to cover the new tag.

A36 needs fresh evidence. Never copy or reuse any `v0.50.126`, `v0.50.125`, or
`v0.50.124` report, attestation, evidence commit, evidence tag object, draft,
or workflow artifact.

Publish only through the release tooling, with the release model and the
production context window named explicitly:

```bash
export OMP_AUTH_BROKER_URL=http://127.0.0.1:47311
export ADK_GATEWAY_URL=http://127.0.0.1:47312
export ADK_RELEASE_MODEL=gpt-6.1-sol
export ADK_RELEASE_MODEL_CONTEXT_WINDOW=262144
scripts/release-tools/release-prep.sh --preflight
sudo -v && scripts/release-tools/release-prep.sh --apply
```

## Stop conditions

- **Approve the release job within 24 hours of `--apply`.** The attestation
  `--apply` signs expires at most 24 hours after it is issued, and the
  same-run evidence artifact is kept for one day. Approve the
  `adk-companion-release` deployment as soon as the tag CI finishes; for
  `v0.50.126` that was about 41 minutes after the tag run started. Run
  `--apply` only when someone can approve inside that window. If the window is
  missed, do not rerun the job: the coordinate is burned. Carry it to
  `v0.50.128` and record the burn here.
- If the tap head moves off `37fec98d` before publication, re-measure
  `PRIOR_TAP_COMMIT` and `PRIOR_CASK_BLOB` first
  (`scripts/companion-release/verify-homebrew-tap-pins.sh` reports the drift).
- The tag is immutable once pushed. If anything mismatches after that, do not
  move, delete, or reuse `v0.50.127`; carry the coordinate to `v0.50.128` and
  record the burn here.
