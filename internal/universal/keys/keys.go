package keys

func Contains(response map[string]any, keys []string) bool {
	for _, key := range keys {
		if response[key] == nil {
			return false
		}
	}

	return true
}
