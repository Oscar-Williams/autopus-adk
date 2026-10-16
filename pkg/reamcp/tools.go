package reamcp

// Tool classification of rea-agents@6.5.0 (SPEC-REA-001 security audit M2),
// recorded with its method in
// .autopus/specs/SPEC-REA-001/evidence/rea-tool-catalog-6.5.0.md. Every one of
// the 206 public tools of the pin is in exactly one list; a tool a new
// pin adds is in none until it is audited, so it is never granted by default.

// staticAnalysisTools reads local targets and retained evidence without
// running, attaching to, instrumenting, or capturing a live target. The
// reverse-engineer subagent holds these and no other REA tool.
var staticAnalysisTools = []string{
	"address_name", "address_to_file_offset", "analyze_function",
	"analyze_javascript_application", "analyze_swift_types", "annotate_native_function",
	"batch_decompile", "binary_overview", "binary_session", "build_call_path",
	"build_reconstruction_obligation_ledger", "close_binary", "comment",
	"compare_application_versions", "compare_artifacts", "compare_bundles", "compare_functions",
	"compare_javascript_export_shapes", "compare_managed_members", "compare_process_captures",
	"compare_source_to_bundle", "compare_wasm_interfaces", "compare_web_captures",
	"compare_web_screenshots", "correlate_static_and_runtime", "current_address",
	"current_document", "current_procedure", "decode_android_resources",
	"decode_interface_builder", "decode_signal_recording", "demangle_swift",
	"evaluate_reconstruction_coverage", "extract_artifact", "find_changed_behavior",
	"find_xrefs_to_name", "get_call_graph", "get_evidence_bundle", "get_navigation_context",
	"get_objc_classes", "get_objc_protocols", "identify_flutter_build", "import_evidence_bundle",
	"import_managed_reconstruction", "inline_comment", "inspect_address_context",
	"inspect_analysis_view", "inspect_android_class", "inspect_android_method",
	"inspect_android_package", "inspect_apktool_client", "inspect_artifact",
	"inspect_asset_catalog", "inspect_binary_layout", "inspect_dart_aot",
	"inspect_electron_fuses", "inspect_evm_interface", "inspect_firmware_regions",
	"inspect_go_binary", "inspect_keyed_archive", "inspect_macho", "inspect_managed_artifact",
	"inspect_managed_members", "inspect_managed_native_boundaries", "inspect_native_api",
	"inspect_native_data_type", "inspect_native_dispatch_metadata", "inspect_native_instruction",
	"inspect_native_load_image", "inspect_packet_capture", "inspect_packet_fields",
	"inspect_pcap_capture", "inspect_pcap_client", "inspect_pe_resources", "inspect_plist",
	"inspect_recorded_crash", "inspect_rust_binary", "inspect_signal_decoders",
	"inspect_signature", "inspect_sqlite_database", "inspect_wasm_artifact",
	"inspect_web_network_capture", "inspect_with_objdump", "list_architectures", "list_bookmarks",
	"list_documents", "list_names", "list_procedures", "list_segments", "list_strings",
	"list_unknowns", "next_address", "open_binary", "prev_address", "procedure_address",
	"procedure_assembly", "procedure_callees", "procedure_callers", "procedure_info",
	"procedure_pseudo_code", "procedure_references", "project_android_application_graph",
	"project_apple_application_graph", "project_harmony_application_graph",
	"project_managed_application_graph", "read_bytes", "read_function_instructions",
	"reconcile_javascript_runtime", "record_unknown", "resolve_containing_procedure",
	"resolve_native_call_targets", "search_android_classes", "search_bytes", "search_procedures",
	"search_strings", "set_address_name", "set_addresses_names", "set_bookmark", "set_comment",
	"set_inline_comment", "trace_android_references", "trace_application_feature",
	"trace_call_path", "trace_dylib_resolution", "trace_feature", "trace_javascript_semantics",
	"trace_native_ui_action", "trace_native_values", "trace_web_source_location",
	"unset_bookmark", "update_unknown", "verify_managed_native_boundaries",
	"verify_reconstruction", "verify_unknown_resolution", "xrefs",
}

// runtimeCaptureTools run, attach to, instrument, or capture a live target,
// device, GUI, or network peer. Only the main session calls them, after the
// operator answers 허용 to a Runtime Capture Request.
var runtimeCaptureTools = []string{
	"analyze_web_bundle", "capture_adb_screen", "capture_browser_scenario",
	"capture_electron_scenario", "capture_native_ui_scenario", "capture_process_scenario",
	"capture_web_screenshot", "close_frida_session", "close_gdb_session",
	"close_rizin_debug_session", "collect_adb_bugreport", "cutter_command", "decompile_jeb_item",
	"discover_webmcp_tools", "execute_rizin_command", "frida_session_status", "gdb_console",
	"gdb_session_status", "goto_address", "inspect_adb_client", "inspect_adb_device",
	"inspect_adb_display", "inspect_adb_package", "inspect_adb_window", "inspect_electron_page",
	"inspect_jeb_client", "inspect_web_event_listeners", "inspect_web_page",
	"install_adb_package", "instrument_with_frida", "list_adb_devices", "list_adb_directory",
	"list_adb_features", "list_adb_packages", "list_adb_processes", "list_adb_services",
	"list_browser_targets", "list_cutter_sessions", "list_electron_targets", "list_frida_devices",
	"list_frida_processes", "list_javascript_runtime_targets", "list_jeb_units",
	"load_frida_script", "observe_javascript_runtime", "observe_native_ui",
	"observe_web_execution", "observe_web_session", "open_jeb_project", "pull_adb_file",
	"pull_adb_package", "push_adb_file", "read_adb_logcat", "read_adb_setting",
	"resolve_adb_packages", "resume_frida_session", "rizin_debug_command",
	"rizin_debug_session_status", "start_adb_activity", "start_adb_app", "start_frida_session",
	"start_gdb_session", "start_rizin_debug_session", "stop_adb_app", "trace_web_module_imports",
	"uninstall_adb_package", "unload_frida_script",
}

// callerPathWriteTools write to a path the caller chooses. They are not
// runtime capture, but the subagent has no write capability, so only the main
// session calls them.
var callerPathWriteTools = []string{
	"export_evidence_bundle", "export_web_scripts", "extract_firmware",
	"recover_javascript_sources",
}

// StaticAnalysisTools returns the REA tools the reverse-engineer subagent may
// hold, sorted.
func StaticAnalysisTools() []string { return append([]string(nil), staticAnalysisTools...) }

// RuntimeCaptureTools returns the REA tools only the main session calls after
// an explicit 허용, sorted.
func RuntimeCaptureTools() []string { return append([]string(nil), runtimeCaptureTools...) }

// CallerPathWriteTools returns the REA tools that write a caller-chosen path,
// sorted.
func CallerPathWriteTools() []string { return append([]string(nil), callerPathWriteTools...) }
