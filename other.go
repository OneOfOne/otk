package otk

// UniqueSlice returns all unique keys in `in` by modifying it in place
func UniqueSlice(in []string) (out []string) {
	set := make(Set, len(in))
	out = in[:0]

	for _, s := range in {
		if set.AddIfNotExists(s) {
			out = append(out, s)
		}
	}

	return
}
