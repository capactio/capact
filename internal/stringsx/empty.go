package stringsx

func AreAllSlicesEmpty(slices ...[]string) bool {
	for _, slice := range slices {
		if len(slice) != 0 {
			return false
		}
	}
	return true
}

func AreAllSlicesNotEmpty(slices ...[]string) bool {
	for _, slice := range slices {
		if len(slice) == 0 {
			return false
		}
	}
	return true
}
