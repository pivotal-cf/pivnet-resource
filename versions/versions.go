package versions

func Since(versions []string, since string) ([]string, error) {
	for i, v := range versions {
		if v == since {
			return versions[:i], nil
		}
	}

	return versions[:1], nil
}

func Reverse(versions []string) ([]string, error) {
	var reversed []string
	for i := len(versions) - 1; i >= 0; i-- {
		reversed = append(reversed, versions[i])
	}

	return reversed, nil
}
