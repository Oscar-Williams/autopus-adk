---
name: agent-teams
description: Native named-teammate orchestration for Claude Code Agent Teams
triggers:
  - agent teams
  - teams
  - 에이전트 팀
category: agentic
level1_metadata: "Claude native teams, named Agent teammates, SendMessage, shared task list"
platforms:
  - claude
---

# Agent Teams Skill

## Purpose

`/auto go SPEC-ID --team` uses Claude Code's native implicit team. It is not a
Workflow alias and does not invoke `route_team`. The top-level interactive
session is the lead; named `Agent` calls become teammates when the native team
feature is available.

## Preflight

Before dispatch:

1. Confirm Claude Code is 2.1.246 or newer.
2. Confirm `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` is already present in the
   session environment. Settings hooks do not inject it.
3. Confirm the session is interactive. Native teams are not a headless or print
   mode surface.
4. Load the runtime schemas for `Agent`, `SendMessage`, and the shared task-list
   tools when they are deferred.

If any precondition fails, stop with a concrete blocker and suggest the default
Route A subagent pipeline. Do not silently switch to `--workflow`.

## Native Lifecycle

The runtime creates the implicit team when the lead launches named teammates.
There is no manual create/delete step and no project-owned team configuration.
Claude cleans the native team configuration automatically after teammates shut
down.

```python
builder = Agent(
    description="Implement disjoint owned paths for the selected SPEC",
    prompt="""Act as Builder. Modify only owned_paths and never forbidden_paths.
    Follow TDD and return exactly: owned_paths, changed_files, verification,
    blockers, next_required_step.""",
    subagent_type="executor",
    name="builder-1",
)

tester = Agent(
    description="Write and verify tests in disjoint test-owned paths",
    prompt="""Act as Tester. Own only test owned_paths. Do not modify production
    paths. Return exactly: owned_paths, changed_files, verification, blockers,
    next_required_step.""",
    subagent_type="tester",
    name="tester",
)

guardian = Agent(
    description="Validate the integrated shared working tree without editing",
    prompt="""Act as Guardian. Treat every source path as read-only. Return
    exactly: owned_paths, changed_files, verification, blockers,
    next_required_step.""",
    subagent_type="validator",
    name="guardian",
)
```

Every call includes the required `description` and `prompt`. Model selection and
effort come from the installed agent definition. Teammate calls do not set
per-call effort, permission mode, team identity, or isolation.

## Shared Task List

The lead decomposes work before teammate dispatch and records it in Claude's
shared task list. Every writable task has a disjoint `owned_paths` set and an
explicit `forbidden_paths` complement. Interdependent source and test files stay
in one task so parallel teammates never need to edit the same file.

The lead assigns and updates tasks with the runtime-exposed task tools. It uses
`SendMessage` for coordination, blockers, shutdown requests, and receipt
clarification. Use the schema exposed by the current runtime rather than a
handwritten legacy payload.

## Result Handoff

Every teammate returns exactly these five fields:

1. `owned_paths`
2. `changed_files`
3. `verification`
4. `blockers`
5. `next_required_step`

The lead checks that changed files are a subset of owned paths, then hands the
integrated shared-tree result to Guardian and final review. A missing field or
an ownership violation is a blocker, not a successful receipt.

## Teardown

On success, abort, or circuit break:

1. Send each live teammate a shutdown request through `SendMessage`.
2. Wait for shutdown acknowledgements using the runtime's native delivery.
3. Let Claude perform automatic team cleanup.

Do not call removed lifecycle tools, edit runtime team configuration, or delete
runtime task directories manually.

## Automatic Entry from Route A

A Claude Code `/auto go` run without `--team`, `--workflow`, or `--solo` can
enter this lifecycle on its own; the selection is on by default. When Route A
reaches Phase 2, the decision the `auto-go` skill runs selects a team only for
an interactive top-level session that already has
`CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`, runs Claude Code 2.1.246 or newer,
and has at least two ready worker units whose owned paths do not overlap.
Otherwise the run stays on subagents and says why in one line.

- Same lifecycle: named teammates, the shared task list, the five receipt
  fields, and the shutdown and acknowledgement teardown above apply unchanged.
- Scope: one ready wave of Phase 2. Each ready unit becomes one Builder
  teammate, `builder-1` … `builder-n` in decision order. No `tester` or
  `guardian` teammate is launched, and every other phase stays on Route A.
- Schema failure: when the `SendMessage` or shared task-list schemas cannot be
  loaded, print the decision's fallback line and continue on Route A. This is
  a fallback, not a stop.
- Handoff: after every teammate acknowledged its shutdown, the lead implements
  the remaining Phase 2 units under Route A, never re-runs a finished unit, and
  only then enters Gate 2.
- Cost: every teammate is a separate Claude session, so a team run uses more
  tokens than subagents, and the feature is experimental. Turn it off with
  `workflow.auto_team: false` in `autopus.yaml`, or pass `--solo` for one run.
  auto 0.50.113 through 0.50.125 refuse an `autopus.yaml` that contains the
  key with any value; auto 0.50.112 and older ignore it, and their
  `auto update` deletes it.
- Version source: the version is read from the running binary named by
  `CLAUDE_CODE_EXECPATH`, never from a `claude` on `PATH`. An install whose
  executable is not named `claude` or `claude.exe` stays on subagents with the
  `claude_version_unknown` line.

## Failure Policy

| Failure | Required action |
|---|---|
| Experimental environment missing | Stop and report the required environment variable. |
| Non-interactive session | Fall back only after reporting that native teams require an interactive session. |
| Teammate dispatch fails | Reassign the disjoint task to another named teammate or return a blocker. |
| Ownership overlap | Stop both writers, repartition files, and dispatch again. |
| Guardian fails | Run a read-only validator subagent with required description and prompt. |

## Separation from Workflows

`--workflow` invokes the dynamic Route A script with `Workflow({scriptPath,
args})`. `--team` invokes the native implicit team described here. The flags
select different substrates and are never defaulted into one another.
