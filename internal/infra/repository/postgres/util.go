package postgres

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullIfEmptyPtr(value *string) any {
	if value == nil || *value == "" {
		return nil
	}
	return *value
}
