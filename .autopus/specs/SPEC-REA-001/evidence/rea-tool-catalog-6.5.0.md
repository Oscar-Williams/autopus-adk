# REA tool catalog — rea-agents@6.5.0 static/runtime split

Recorded 2026-10-16 for the REA security audit finding M2 (SPEC-REA-001). The `reverse-engineer` subagent holds only the static analysis tools below; runtime-capture tools are called only by the main session after the operator answers 허용 (allow).

## Source

- `npm pack rea-agents@6.5.0 --ignore-scripts` into a scratch directory (download only), then `tar -xzf`. No REA code was installed or executed.
- Tarball sha256 `531717a06e92ca651aab52691190f916fac044272c01f875e22c5e293218e699`, npm shasum `7ea5e6f669bc18c45dbf5dba653edaa1ada07ab3`, 1650 files, `dist/generatedPackageMetadata.js` version `6.5.0`.
- The package does not ship `docs/mcp-contracts.md` (the README links it). The authoritative per-tool inventory it ships is `dist/contracts/toolEffects.js` `TOOL_EFFECTS` ("Explicit effect audit for every public tool"), which `dist/contracts/toolContracts.js` resolves for every registered contract (`toolContractMetadata` throws on a tool without an audit). It was parsed as text: 206 tools, each with `executesForeignCode`, `mutatesDevice`, `changesUiState`, `accessesNetwork`, `launchesProcess`, `writesFilesystem`, `mutatesTarget`, `mutatesSession`, `mayDiscardData`, `idempotent`.
- Tool descriptions were read from `dist/contracts/**` to confirm borderline rows (for example `analyze_web_bundle` captures from a live CDP page; `observe_native_ui` attaches to a running app; `inspect_web_network_capture` reads a retained local HAR file).

## Rules (first match wins)

1. Runtime by family: names containing `adb`, `frida`, `gdb`, `rizin_debug`, `execute_rizin_command`, `cutter`, `capture_`, or `observe_` run, attach to, instrument, or capture a live target, device, or GUI.
2. Runtime by effect: `executesForeignCode`, `mutatesDevice`, `changesUiState`, or `accessesNetwork`. A tool that reaches the network is treated as runtime because the subagent cannot be shown to reach only a local, non-target service (this conservatively includes the JEB bridge and `trace_web_module_imports`).
3. Main session only, not runtime: tools that write to a caller-chosen path (`output_directory` or an export path). The subagent has no write capability, so they are excluded from its list as well.
4. Every other tool is static. `launchesProcess` alone means REA starts its own analysis tooling (Ghidra, objdump, tshark, apktool, …) on local files, not the target. `mutatesTarget` on annotation tools edits only the provider analysis database; the original bytes stay unchanged.

## Static — the subagent allow-list (135)

Claude Code grants each as `mcp__rea__<tool>`, Antigravity as `mcp_rea_<tool>`. OMP and OpenCode get none (Autopus registers no REA server there).

| Tool | Effects (X D U N P W) |
|---|---|
| `address_name` | `......` |
| `address_to_file_offset` | `......` |
| `analyze_function` | `......` |
| `analyze_javascript_application` | `......` |
| `analyze_swift_types` | `......` |
| `annotate_native_function` | `....PW` |
| `batch_decompile` | `......` |
| `binary_overview` | `......` |
| `binary_session` | `....P.` |
| `build_call_path` | `......` |
| `build_reconstruction_obligation_ledger` | `......` |
| `close_binary` | `.....W` |
| `comment` | `......` |
| `compare_application_versions` | `......` |
| `compare_artifacts` | `......` |
| `compare_bundles` | `......` |
| `compare_functions` | `......` |
| `compare_javascript_export_shapes` | `......` |
| `compare_managed_members` | `......` |
| `compare_process_captures` | `......` |
| `compare_source_to_bundle` | `......` |
| `compare_wasm_interfaces` | `......` |
| `compare_web_captures` | `......` |
| `compare_web_screenshots` | `......` |
| `correlate_static_and_runtime` | `......` |
| `current_address` | `......` |
| `current_document` | `......` |
| `current_procedure` | `......` |
| `decode_android_resources` | `....PW` |
| `decode_interface_builder` | `......` |
| `decode_signal_recording` | `....PW` |
| `demangle_swift` | `....P.` |
| `evaluate_reconstruction_coverage` | `......` |
| `extract_artifact` | `....PW` |
| `find_changed_behavior` | `......` |
| `find_xrefs_to_name` | `......` |
| `get_call_graph` | `......` |
| `get_evidence_bundle` | `......` |
| `get_navigation_context` | `......` |
| `get_objc_classes` | `......` |
| `get_objc_protocols` | `......` |
| `identify_flutter_build` | `.....W` |
| `import_evidence_bundle` | `......` |
| `import_managed_reconstruction` | `......` |
| `inline_comment` | `......` |
| `inspect_address_context` | `......` |
| `inspect_analysis_view` | `......` |
| `inspect_android_class` | `....PW` |
| `inspect_android_method` | `....PW` |
| `inspect_android_package` | `....PW` |
| `inspect_apktool_client` | `....P.` |
| `inspect_artifact` | `....PW` |
| `inspect_asset_catalog` | `....P.` |
| `inspect_binary_layout` | `....PW` |
| `inspect_dart_aot` | `.....W` |
| `inspect_electron_fuses` | `......` |
| `inspect_evm_interface` | `....PW` |
| `inspect_firmware_regions` | `....PW` |
| `inspect_go_binary` | `......` |
| `inspect_keyed_archive` | `......` |
| `inspect_macho` | `....P.` |
| `inspect_managed_artifact` | `......` |
| `inspect_managed_members` | `......` |
| `inspect_managed_native_boundaries` | `......` |
| `inspect_native_api` | `......` |
| `inspect_native_data_type` | `......` |
| `inspect_native_dispatch_metadata` | `......` |
| `inspect_native_instruction` | `......` |
| `inspect_native_load_image` | `......` |
| `inspect_packet_capture` | `....PW` |
| `inspect_packet_fields` | `....PW` |
| `inspect_pcap_capture` | `....PW` |
| `inspect_pcap_client` | `....PW` |
| `inspect_pe_resources` | `......` |
| `inspect_plist` | `....P.` |
| `inspect_recorded_crash` | `....PW` |
| `inspect_rust_binary` | `......` |
| `inspect_signal_decoders` | `....PW` |
| `inspect_signature` | `....P.` |
| `inspect_sqlite_database` | `....PW` |
| `inspect_wasm_artifact` | `....PW` |
| `inspect_web_network_capture` | `....PW` |
| `inspect_with_objdump` | `....P.` |
| `list_architectures` | `....P.` |
| `list_bookmarks` | `......` |
| `list_documents` | `......` |
| `list_names` | `......` |
| `list_procedures` | `......` |
| `list_segments` | `......` |
| `list_strings` | `......` |
| `list_unknowns` | `......` |
| `next_address` | `......` |
| `open_binary` | `....P.` |
| `prev_address` | `......` |
| `procedure_address` | `......` |
| `procedure_assembly` | `......` |
| `procedure_callees` | `......` |
| `procedure_callers` | `......` |
| `procedure_info` | `......` |
| `procedure_pseudo_code` | `......` |
| `procedure_references` | `......` |
| `project_android_application_graph` | `......` |
| `project_apple_application_graph` | `......` |
| `project_harmony_application_graph` | `......` |
| `project_managed_application_graph` | `......` |
| `read_bytes` | `......` |
| `read_function_instructions` | `......` |
| `reconcile_javascript_runtime` | `......` |
| `record_unknown` | `......` |
| `resolve_containing_procedure` | `......` |
| `resolve_native_call_targets` | `......` |
| `search_android_classes` | `....PW` |
| `search_bytes` | `......` |
| `search_procedures` | `......` |
| `search_strings` | `......` |
| `set_address_name` | `......` |
| `set_addresses_names` | `......` |
| `set_bookmark` | `......` |
| `set_comment` | `......` |
| `set_inline_comment` | `......` |
| `trace_android_references` | `....PW` |
| `trace_application_feature` | `......` |
| `trace_call_path` | `......` |
| `trace_dylib_resolution` | `......` |
| `trace_feature` | `......` |
| `trace_javascript_semantics` | `......` |
| `trace_native_ui_action` | `......` |
| `trace_native_values` | `......` |
| `trace_web_source_location` | `....PW` |
| `unset_bookmark` | `......` |
| `update_unknown` | `......` |
| `verify_managed_native_boundaries` | `......` |
| `verify_reconstruction` | `......` |
| `verify_unknown_resolution` | `......` |
| `xrefs` | `......` |

## Runtime — main session only, after an explicit 허용 (67)

The agent names one of these with its arguments in a `## Runtime Capture Request`; the main session asks the operator once and calls the tool itself only on 허용. `--auto` never skips the question.

| Tool | Effects (X D U N P W) | Why runtime |
|---|---|---|
| `analyze_web_bundle` | `...N..` | accessesNetwork |
| `capture_adb_screen` | `...NPW` | Android device (adb) |
| `capture_browser_scenario` | `X.UNPW` | target capture |
| `capture_electron_scenario` | `X.UNPW` | target capture |
| `capture_native_ui_scenario` | `X.UNPW` | target capture |
| `capture_process_scenario` | `X..NPW` | target capture |
| `capture_web_screenshot` | `...N..` | target capture |
| `close_frida_session` | `...NP.` | Frida instrumentation |
| `close_gdb_session` | `X...P.` | GDB debugger |
| `close_rizin_debug_session` | `X...P.` | Rizin debugger |
| `collect_adb_bugreport` | `...NPW` | Android device (adb) |
| `cutter_command` | `X.UNPW` | Cutter GUI bridge |
| `decompile_jeb_item` | `...N..` | accessesNetwork |
| `discover_webmcp_tools` | `...N..` | accessesNetwork |
| `execute_rizin_command` | `X..NPW` | arbitrary Rizin command |
| `frida_session_status` | `......` | Frida instrumentation |
| `gdb_console` | `X..NPW` | GDB debugger |
| `gdb_session_status` | `......` | GDB debugger |
| `goto_address` | `..U...` | changesUiState |
| `inspect_adb_client` | `....P.` | Android device (adb) |
| `inspect_adb_device` | `...NP.` | Android device (adb) |
| `inspect_adb_display` | `...NP.` | Android device (adb) |
| `inspect_adb_package` | `...NP.` | Android device (adb) |
| `inspect_adb_window` | `...NP.` | Android device (adb) |
| `inspect_electron_page` | `...N..` | accessesNetwork |
| `inspect_jeb_client` | `...N..` | accessesNetwork |
| `inspect_web_event_listeners` | `...N..` | accessesNetwork |
| `inspect_web_page` | `...N..` | accessesNetwork |
| `install_adb_package` | `.D.NP.` | Android device (adb) |
| `instrument_with_frida` | `X..NP.` | Frida instrumentation |
| `list_adb_devices` | `...NP.` | Android device (adb) |
| `list_adb_directory` | `...NP.` | Android device (adb) |
| `list_adb_features` | `...NP.` | Android device (adb) |
| `list_adb_packages` | `...NP.` | Android device (adb) |
| `list_adb_processes` | `...NP.` | Android device (adb) |
| `list_adb_services` | `...NP.` | Android device (adb) |
| `list_browser_targets` | `...N..` | accessesNetwork |
| `list_cutter_sessions` | `...N..` | Cutter GUI bridge |
| `list_electron_targets` | `...N..` | accessesNetwork |
| `list_frida_devices` | `...NP.` | Frida instrumentation |
| `list_frida_processes` | `...NP.` | Frida instrumentation |
| `list_javascript_runtime_targets` | `...N..` | accessesNetwork |
| `list_jeb_units` | `...N..` | accessesNetwork |
| `load_frida_script` | `X..NP.` | Frida instrumentation |
| `observe_javascript_runtime` | `...N..` | live observation |
| `observe_native_ui` | `....PW` | live observation |
| `observe_web_execution` | `...N..` | live observation |
| `observe_web_session` | `...N..` | live observation |
| `open_jeb_project` | `...N..` | accessesNetwork |
| `pull_adb_file` | `...NPW` | Android device (adb) |
| `pull_adb_package` | `...NPW` | Android device (adb) |
| `push_adb_file` | `.D.NPW` | Android device (adb) |
| `read_adb_logcat` | `...NP.` | Android device (adb) |
| `read_adb_setting` | `...NP.` | Android device (adb) |
| `resolve_adb_packages` | `...NP.` | Android device (adb) |
| `resume_frida_session` | `X..NP.` | Frida instrumentation |
| `rizin_debug_command` | `X.UNPW` | Rizin debugger |
| `rizin_debug_session_status` | `......` | Rizin debugger |
| `start_adb_activity` | `.D.NP.` | Android device (adb) |
| `start_adb_app` | `.D.NP.` | Android device (adb) |
| `start_frida_session` | `...NP.` | Frida instrumentation |
| `start_gdb_session` | `....P.` | GDB debugger |
| `start_rizin_debug_session` | `X...P.` | Rizin debugger |
| `stop_adb_app` | `.D.NP.` | Android device (adb) |
| `trace_web_module_imports` | `...NPW` | accessesNetwork |
| `uninstall_adb_package` | `.D.NP.` | Android device (adb) |
| `unload_frida_script` | `...NP.` | Frida instrumentation |

## Writes a caller-chosen path — main session only (4)

| Tool | Effects (X D U N P W) |
|---|---|
| `export_evidence_bundle` | `.....W` |
| `export_web_scripts` | `.....W` |
| `extract_firmware` | `....PW` |
| `recover_javascript_sources` | `....PW` |

Effect legend: X executesForeignCode, D mutatesDevice, U changesUiState, N accessesNetwork, P launchesProcess, W writesFilesystem; `.` is false.

A new pin (`reamcp.Pin`) must repeat this audit: a tool REA adds is not in the allow-list until it is classified here.
