# SPEC-REA-001 T2 기존 test 분류

기준: `git diff --name-status 4d73aa91..feat/rea-001 -- '*_test.go' '*.golden' 'evals/'`(acceptance S17)와 `git diff 2c9cf0c8..feat/rea-001 -- '*_test.go' '**/testdata/**' 'evals/**'`는 같은 파일 집합이다. 바뀐 기존 파일 101개를 spec.md `## Existing Test Changes`의 규칙에 하나씩 맞춘다. 새로 더한 test 파일 34개는 기존 test 변경이 아니라 여기 넣지 않는다.

- 규칙 5(MCP inventory test 무변경)는 지켜졌다: `GT-SETTINGS-MCP-SERVERS.json`, `antigravity_settings_test.go`, `antigravity_clean_ownership_test.go`, `claude_lifecycle_contract_test.go`, `codex_config_test.go`, `codex_latest_v2_test.go`는 이 목록에 없다.
- `N`은 새 기능 test가 쓰는 helper 추가이고 기존 단언을 바꾸지 않는다. `밖`은 같은 branch에 섞인 이 SPEC 밖의 commit이다.
- 분류 수(한 파일이 둘 이상의 규칙에 들 수 있다): 규칙 1 46건, 규칙 2 13건, 규칙 3 26건, 규칙 4 3건, 규칙 6 1건, 규칙 7 7건, 규칙 8 6건, 규칙 9 1건, 규칙 10 1건, 규칙 11 1건, 규칙 12 1건, 규칙 13 1건, N 1건, 밖 1건.

| 파일 | 규칙 | 이유 |
|---|---|---|
| `evals/harness/baseline.json` | 2 | reverse를 더한 GT-ROUTE task 6개의 expectation digest와 set digest만 바뀐다 |
| `evals/harness/tasks/surface/GT-ROUTE-ANTIGRAVITY-COMMANDS.json` | 2 | frozen route 목록·router 문장에 `reverse` 항목만 더한다 |
| `evals/harness/tasks/surface/GT-ROUTE-ANTIGRAVITY-DETAILS.json` | 2 | frozen route 목록·router 문장에 `reverse` 항목만 더한다 |
| `evals/harness/tasks/surface/GT-ROUTE-CLAUDE-DETAILS.json` | 2 | frozen route 목록·router 문장에 `reverse` 항목만 더한다 |
| `evals/harness/tasks/surface/GT-ROUTE-CODEX-SKILL-NAMES.json` | 2 | frozen route 목록·router 문장에 `reverse` 항목만 더한다 |
| `evals/harness/tasks/surface/GT-ROUTE-OMP-EXACT-MAP.json` | 2 | frozen route 목록·router 문장에 `reverse` 항목만 더한다 |
| `evals/harness/tasks/surface/GT-ROUTE-OPENCODE-DETAILS.json` | 2 | frozen route 목록·router 문장에 `reverse` 항목만 더한다 |
| `internal/cli/edit_guard_corpus_test.go` | 12 | 합법 corpus fixture가 `README.md`를 실제 파일로 심는다(없는 README의 Gemini replace는 보호 파일을 고친다) |
| `internal/cli/init_compose_omp_test.go` | 7 | OMP 선택지에 두 model이 더해져 같은 model을 고르는 입력 번호가 4→5로 밀린다 |
| `internal/cli/init_compose_test.go` | 1 | S11 custom 기대 map에 reverse-engineer(opus)를 더하고 Claude `*` 칸 수 16→17 |
| `internal/cli/init_gitignore_policy_test.go` | 9 | required 목록에 `.autopus/reverse/` 한 줄만 더한다 |
| `internal/cli/json_contract_commands_test.go` | 11 | 빌드한 binary의 doctor 행만 자식 HOME을 t.TempDir()로 둔다(envelope 단언 그대로) |
| `internal/cli/omp_dispatch_test.go` | 1, 2 | OMP agent 정의 16→17(규칙 1), command 20→21(`auto-reverse`, 규칙 2) |
| `internal/cli/pipeline_backend_omp_rpc_test.go` | 6 | TestMain이 reaNodeExec를 v24.11.0·종료 0으로 고정하는 한 줄 |
| `internal/cli/platform_omp_agent_migration_test.go` | 1 | migration 줄·파일 수 16→17, 표면 줄 수 18→19, 문장 `17개` |
| `internal/cli/platform_omp_autopus_agents_test.go` | 1 | OMP 기대 agent 수 17→18(17 + task)과 주석 |
| `internal/cli/platform_omp_profile_flags_test.go` | 7 | 두 catalog fixture에 `claude-opus-5-5`·`gpt-6.1-sol` 행만 더한다 |
| `internal/cli/platform_omp_profile_pins_test.go` | 7 | attest 가능한 model 목록이 두 model만큼 4→6 |
| `internal/cli/quality_apply_test.go` | 1 | Codex agent 파일 수 16→17 |
| `internal/cli/quality_compose_bind_efforts_test.go` | 1 | 변경 줄 `(16개 에이전트)`→`(17개 에이전트)` |
| `internal/cli/quality_compose_bind_refresh_test.go` | 1, 4 | Historical 갱신 줄에 `reverse-engineer→opus` 추가(규칙 4), 변경 줄 에이전트 수 16→17(규칙 1) |
| `internal/cli/quality_compose_bind_text_test.go` | 1 | 변경 줄 에이전트 수 16→17(ko·en) |
| `internal/cli/quality_compose_driver_gate_test.go` | 1 | fixture X의 사용자 수정 다름 수 12→13과 주석 |
| `internal/cli/quality_compose_effort_test.go` | 3 | 탐색·보조 effort 선택지가 opus 멤버의 모델을 함께 보이고 저장 efforts key가 reverse-engineer로 바뀐다 |
| `internal/cli/quality_compose_entry_test.go` | 1 | Claude `*` 칸 16→17, Codex `*` 칸 4→5 |
| `internal/cli/quality_compose_form_edit_test.go` | 1 | Claude `*` 칸 수와 변경 줄 에이전트 수 16→17 |
| `internal/cli/quality_compose_form_effort_test.go` | 1 | 변경 줄 에이전트 수 16→17 |
| `internal/cli/quality_compose_form_efforts_carry_test.go` | 1 | 변경 줄 에이전트 수 16→17 |
| `internal/cli/quality_compose_form_entry_test.go` | 1 | Claude `*` 칸 16→17, Codex `대체 4칸`→`5칸`(reverse-engineer도 대체 model을 씀) |
| `internal/cli/quality_compose_form_paths_test.go` | 7 | OMP task 선택지에 두 model이 더해지고 같은 model의 입력 번호가 3→4로 밀린다 |
| `internal/cli/quality_compose_groups_test.go` | 3 | 탐색·보조 멤버에 reverse-engineer, `현재:` 혼합 표시 |
| `internal/cli/quality_compose_line_golden_test.go` | 10 | R fixture의 REA를 실제 `auto update --no-rea`로 off로 미리 정한다(golden 그대로) |
| `internal/cli/quality_compose_omp_autopus_test.go` | 7 | `composeOMPCatalogML`에 두 model 행, 그 목록 단언에 두 model만 더한다 |
| `internal/cli/quality_compose_omp_test.go` | 7 | S8 입력에서 같은 model을 고르는 번호가 3→4로 밀린다 |
| `internal/cli/quality_compose_session_test.go` | 1, 3 | 탐색·보조 줄 멤버·`현재:` 표시(규칙 3), 변경 줄 에이전트 수 16→17(규칙 1) |
| `internal/cli/quality_compose_table_cells_test.go` | 1 | Codex `대체 16칸`→`17칸`, `*` Codex 칸 목록에 reverse-engineer |
| `internal/cli/quality_compose_table_fixture_test.go` | 1 | S6 그룹 편집 기대 map에 탐색·보조 멤버 reverse-engineer(haiku) |
| `internal/cli/quality_compose_table_test.go` | 3 | 역할 그룹 순서의 행 목록에 reverse-engineer |
| `internal/cli/quality_compose_tool_test.go` | 1 | 변경 줄 에이전트 수 16→17, Codex custom 기대 map에 reverse-engineer(opus) |
| `internal/cli/quality_compose_view_test.go` | 8 | REA 질문이 열리는 대화형 update 입력 앞에 Enter(`reaEnter`) |
| `internal/cli/quality_config_block_linebreak_test.go` | 8 | REA 질문이 열리는 대화형 update 입력 앞에 Enter(`reaEnter`) |
| `internal/cli/quality_config_block_test.go` | 1 | S6 custom 기대 map에 탐색·보조 멤버 reverse-engineer(haiku) |
| `internal/cli/quality_omp_custom_test.go` | 7 | OMP custom 선택지 번호가 두 model만큼 밀린다(3→4, 2→3)와 주석 |
| `internal/cli/quality_tool_persist_test.go` | 1 | 사용자 수정 다름 수 12→13 |
| `internal/cli/testdata/quality_compose/line-quality-D-global.golden` | 3 | reverse-engineer 행 1줄만 더해진다 |
| `internal/cli/testdata/quality_compose/line-quality-D-groups.golden` | 1, 3 | reverse-engineer 행 2줄, 탐색·보조 멤버·`현재:` 표시(규칙 3), 변경 줄 16→17(규칙 1) |
| `internal/cli/testdata/quality_compose/line-update-R-enter.golden` | 3 | reverse-engineer 행 1줄만 더해진다 |
| `internal/cli/testdata/quality_compose/preview-D-K-CAT-A.golden` | 3 | reverse-engineer 행 1줄만 더해진다 |
| `internal/cli/testdata/quality_compose/table-B2-CAT-A.golden` | 3 | reverse-engineer 행 1줄만 더해진다 |
| `internal/cli/testdata/quality_compose/table-C-CAT-A.golden` | 3 | reverse-engineer 행 1줄만 더해진다 |
| `internal/cli/testdata/quality_compose/table-D-CAT-A.golden` | 3 | reverse-engineer 행 1줄만 더해진다 |
| `internal/cli/testdata/quality_compose/table-D-CAT-B.golden` | 3, 13 | reverse-engineer 행(규칙 3)의 Codex 칸 `세션 모델`과 footer 대체 줄 1개(규칙 13) |
| `internal/cli/testdata/quality_compose/table-D-CAT-C.golden` | 3 | reverse-engineer 행 1줄만 더해진다 |
| `internal/cli/testdata/quality_compose/table-D-CAT-D.golden` | 3 | reverse-engineer 행 1줄만 더해진다 |
| `internal/cli/testdata/quality_compose/table-H-CAT-A.golden` | 3 | reverse-engineer 행 1줄만 더해진다 |
| `internal/cli/testdata/quality_compose/table-K-CAT-A.golden` | 3 | reverse-engineer 행 1줄만 더해진다 |
| `internal/cli/testdata/quality_compose/table-O1-CAT-A.golden` | 3 | reverse-engineer 행과 그 행의 `6.1-sol · xhigh`가 넓힌 Codex 열 폭(값 변화 없음) |
| `internal/cli/testdata/quality_compose/table-O2-CAT-A.golden` | 3 | reverse-engineer 행 1줄만 더해진다 |
| `internal/cli/testdata/quality_compose/table-P-CAT-A.golden` | 3 | reverse-engineer 행 1줄만 더해진다 |
| `internal/cli/testdata/quality_compose/table-R-CAT-A.en.golden` | 3 | reverse-engineer 행 1줄만 더해진다 |
| `internal/cli/testdata/quality_compose/table-R-CAT-A.golden` | 3 | reverse-engineer 행 1줄만 더해진다 |
| `internal/cli/testdata/quality_compose/table-R-CAT-B.golden` | 1, 3 | reverse-engineer 행(규칙 3), `대체 16칸`→`17칸`(규칙 1) |
| `internal/cli/testdata/quality_compose/table-R-CAT-C.golden` | 3 | reverse-engineer 행 1줄만 더해진다 |
| `internal/cli/testdata/quality_compose/table-R-CAT-D.golden` | 3 | reverse-engineer 행 1줄만 더해진다 |
| `internal/cli/testdata/quality_compose/table-R2-CAT-A.golden` | 3 | reverse-engineer 행 1줄만 더해진다 |
| `internal/cli/testdata/quality_compose/table-X-CAT-A.golden` | 3 | reverse-engineer 행 1줄만 더해진다 |
| `internal/cli/tui/export_test.go` | N | 새 REA 단계 test가 쓰는 export alias 2개 추가, 기존 단언 변경 없음 |
| `internal/cli/update_compose_form_test.go` | 8 | REA 질문이 열리는 대화형 update 3건에 Enter(`reaEnter`) |
| `internal/cli/update_compose_omp_test.go` | 8 | REA 질문이 열리는 대화형 update에 Enter(`reaEnter`) |
| `internal/cli/update_compose_override_test.go` | 8 | REA 질문이 열리는 대화형 update에 Enter(`reaEnter`) |
| `internal/cli/update_compose_test.go` | 1, 8 | 대화형 update 입력에 Enter(규칙 8), Claude `*` 칸 16→17·Codex 4→5(규칙 1) |
| `pkg/adapter/antigravity/router_budget_test.go` | 2 | frozen route 목록에 `reverse` |
| `pkg/adapter/claude/claude_native_balanced_agent_test.go` | 1 | balanced 표준 배치 기대 map에 reverse-engineer(opus·high), 용량 16→17 |
| `pkg/adapter/claude/router_budget_test.go` | 2 | frozen route 목록에 `reverse` |
| `pkg/adapter/codex/codex_agents_test.go` | 1 | Codex agent TOML 수 16→17 |
| `pkg/adapter/codex/codex_native_balanced_test.go` | 1 | 표준 배치 기대 map에 reverse-engineer(sol·xhigh), catalog를 모를 때 보고하는 profile 수 2→3 |
| `pkg/adapter/codex/codex_surface_test.go` | 2 | workflow surface 목록에 `auto-reverse` |
| `pkg/adapter/integration_test.go` | 1 | Codex agent 파일 수 16→17 |
| `pkg/adapter/latest_cli_omp_contract_test.go` | 1 | OMP agent 정의 수 16→17 |
| `pkg/adapter/omp/omp_acceptance_surface_test.go` | 1 | OMP agent 정의 수 16→17 |
| `pkg/adapter/omp/omp_agent_surface_report_test.go` | 1 | missing·recorded agent 수 15→16, 16→17 |
| `pkg/adapter/omp/omp_agents_test.go` | 1 | agent 수 16→17과 test 이름 Sixteen→Seventeen |
| `pkg/adapter/omp/omp_latest_contract_test.go` | 1 | OMP agent 정의 수 16→17 |
| `pkg/adapter/omp/omp_model_activation_rpc_autopus_test.go` | 1 | 실패 메시지의 정규 에이전트 수 16→17 |
| `pkg/adapter/omp/omp_native_pipeline_route_test.go` | 2 | OMP command 파일 수 20→21(`auto-reverse`) |
| `pkg/adapter/omp/omp_native_registry_test.go` | 1 | OMP agent 정의 수 16→17 |
| `pkg/adapter/omp/omp_render_mode_test.go` | 1 | pipeline 문장 `16 Autopus agents`→`17` |
| `pkg/adapter/omp/omp_validate_model_integration_test.go` | 1 | agent registry 수 16→17 |
| `pkg/adapter/omp/omp_workflow_parity_test.go` | 2 | workflowSpecs 수 20→21과 메시지 |
| `pkg/config/codex_native_balanced_test.go` | 1 | Codex 표준 배치 기대 map에 reverse-engineer(sol·xhigh) |
| `pkg/config/defaults_test.go` | 1 | balanced 기대 map에 reverse-engineer(opus), ultra opus 수 9→10 |
| `pkg/config/native_balanced_test.go` | 1 | reverse-engineer만 analysis rung 후보(opus-5-5@high, 6.1-sol@xhigh)로 단언 |
| `pkg/config/omp_agent_source_test.go` | 1 | Autopus OMP 이름 수 17→18(17 + task) |
| `pkg/config/quality_effective_test.go` | 1 | fallback·표준 balanced 기대 map에 reverse-engineer(opus) |
| `pkg/config/quality_preset_history_runtime_test.go` | 4 | Historical 행에는 reverse-engineer가 없어 analysis rung을 받는다는 기대 행 |
| `pkg/config/quality_preset_history_test.go` | 1, 4 | HB6·HU3 Historical 행 추가(규칙 4), 다름 수 +1(규칙 1) |
| `pkg/config/role_model_policy_balanced_test.go` | 1 | balanced route 기대 map에 reverse-engineer analysis 후보, test 이름 Sixteen→Seventeen |
| `pkg/config/role_model_policy_builtin_test.go` | 1 | ultra capability 기본 기대 map에 reverse-engineer |
| `pkg/config/role_model_policy_pins_test.go` | 1 | OMP agent 이름 수 17→18 |
| `pkg/config/role_model_policy_test.go` | 1 | policy 계약 matrix에 reverse-engineer(`autopus_reverse_engineer`, deep_reasoning) 행 |
| `pkg/worker/a2a/control_plane_test.go` | 밖 | 같은 branch의 별도 commit `33b5bb55`(A2A `redline_instructions.v1` capability), 이 SPEC 변경 아님 |

## agy 1.4 수정 (2026-10-17, branch `fix/v0126-agy`, 규칙 14)

기준: `git diff --name-status 8a0da3a1..fix/v0126-agy -- '*_test.go' '*.golden' 'evals/'`. 바뀐 기존 파일 14개는 모두 규칙 14이고, 새 test 파일 `pkg/content/agent_transformer_gemini_test.go`, `pkg/adapter/antigravity/antigravity_rea_mcp_boundary_test.go`(다시 쓴 REA test의 clean·symlink 절을 300줄 한도로 떼어 냄), `internal/cli/rea_plugin_mcp_test.go`는 기존 test 변경이 아니다.

| 파일 | 규칙 | 이유 |
|---|---|---|
| `evals/harness/baseline.json` | 14 | GT-AGENT-READONLY-REVIEW의 expectation digest와 set digest만 바뀐다 |
| `evals/harness/tasks/surface/GT-AGENT-READONLY-REVIEW.json` | 14 | Antigravity needle이 `glob`·`run_shell_command`·`replace` 대신 `find_by_name`·`run_command`·`edit_file`이다 |
| `internal/cli/init_rea_test.go` | 14 | 등록 줄의 Antigravity 파일이 `.agents/plugins/autopus/mcp_config.json`이다 |
| `internal/cli/rea_run_test.go` | 14 | 등록 줄의 Antigravity 파일이 같다 |
| `internal/cli/rea_target_posix_test.go` | 14 | 거부·건너뜀 줄의 Antigravity 파일이 같다 |
| `internal/cli/update_rea_test.go` | 14 | SHA-256 목록의 Antigravity 파일이 같고 `--plan --rea`는 그 파일의 `create` 줄을 기대한다 |
| `internal/cli/update_rea_off_test.go` | 14 | off sweep 미리보기가 managed 항목만 담던 플러그인 파일의 `remove` 줄이다 |
| `pkg/adapter/antigravity/antigravity_rea_mcp_test.go` | 14 | 대상이 플러그인 파일로 옮겨 다시 썼다(add·remove·사용자 서버·other_pin·user_owned·읽을 수 없는 JSON·legacy sweep·clean·symlink) |
| `pkg/adapter/antigravity/antigravity_skill_compiler_evolution_test.go` | 14 | executor 도구 기대가 agy 이름이다 |
| `pkg/content/agent_transformer_rea_test.go` | 14 | Antigravity reverse-engineer는 읽기 도구 셋과 `inheritMcp: true`, REA gate fixture도 같은 규칙이다 |
| `pkg/content/context_evolution_test.go` | 14 | 도구 투영 기대가 agy 이름이고 frontmatter 구조체에 `inheritMcp`가 더해진다 |
| `pkg/content/hooks_rea_test.go` | 14 | 사전 승인 단언이 플러그인 파일의 managed 항목을 읽는다 |
| `pkg/reamcp/file_test.go` | 14 | 오류 문구의 Antigravity 파일이 같고, 주석 JSON 거부와 `RemovesFile` test가 더해진다 |
| `pkg/reamcp/shape_test.go` | 14 | `TargetRelPath(Antigravity)`와 `LegacyTargetRelPath` 기대 |
