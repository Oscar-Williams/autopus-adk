package claude

// The permission entry sets the ledger records and the retraction clean
// applies to .claude/settings.json.

// presentStrings returns the values of desired that existing holds, in
// desired's order.
func presentStrings(desired, existing []string) []string {
	held := make(map[string]bool, len(existing))
	for _, value := range existing {
		held[value] = true
	}
	present := make([]string, 0, len(desired))
	for _, value := range desired {
		if held[value] {
			present = append(present, value)
		}
	}
	return present
}

func missingStrings(desired, existing []string) []string {
	seen := make(map[string]bool, len(existing))
	for _, value := range existing {
		seen[value] = true
	}
	missing := make([]string, 0, len(desired))
	for _, value := range desired {
		if !seen[value] {
			missing = append(missing, value)
			seen[value] = true
		}
	}
	return missing
}

func appendUniqueStrings(base []string, values ...string) []string {
	seen := make(map[string]bool, len(base)+len(values))
	result := make([]string, 0, len(base)+len(values))
	for _, value := range base {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

// removeManagedPermissions retracts the ledger's allow entries from settings.
// It never retracts a deny: a ledger the checks trusted may still be forged
// (Windows has no mode mark; a umask 077 checkout writes 0600), and a deny it
// named must not leave the user's settings, while a leftover deny fails safe.
func removeManagedPermissions(settings map[string]any, ledger *claudePermissionLedger) bool {
	if ledger == nil {
		return false
	}
	permissions, ok := settings["permissions"].(map[string]any)
	if !ok {
		return false
	}
	changed := filterPermissionValues(permissions, "allow", ledger.Allow)
	if len(permissions) == 0 {
		delete(settings, "permissions")
	}
	return changed
}

func filterPermissionValues(permissions map[string]any, key string, managed []string) bool {
	if len(managed) == 0 {
		return false
	}
	values, ok := permissions[key].([]any)
	if !ok {
		return false
	}
	owned := make(map[string]bool, len(managed))
	for _, value := range managed {
		owned[value] = true
	}
	kept := make([]any, 0, len(values))
	changed := false
	for _, value := range values {
		text, isString := value.(string)
		if isString && owned[text] {
			changed = true
			continue
		}
		kept = append(kept, value)
	}
	if len(kept) == 0 {
		delete(permissions, key)
	} else {
		permissions[key] = kept
	}
	return changed
}
