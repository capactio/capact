package ptr

// String returns pointer to a given input string value.
func String(in string) *string {
	return &in
}

// String returns string value for a given pointer string.
// Support nil values.
func StringPtrToString(in *string) string {
	if in != nil {
		return *in
	}
	return ""
}

// Bool returns pointer to a given input bool value.
func Bool(in bool) *bool {
	return &in
}

// Int32 returns pointer to a given input int32 value.
func Int32(in int32) *int32 {
	return &in
}

// Int64 returns pointer to a given input int64 value.
func Int64(in int64) *int64 {
	return &in
}
