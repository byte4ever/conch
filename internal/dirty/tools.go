package dirty

func containsDuplicate[T comparable](list []T) bool {
	dedup := make(map[T]struct{}, len(list))
	for _, t := range list {
		dedup[t] = struct{}{}
	}

	return len(dedup) < len(list)
}
