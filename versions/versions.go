package versions

func Since(versions []string, since string) ([]string, error) {
	for i, v := range versions {
		if v == since {
			return versions[:i], nil
		}
	}

	return versions[:0], nil
}
