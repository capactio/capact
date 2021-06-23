package ptr

func String(in string) *string {
	return &in
}

func StringPtrToString(in *string) string {
	if in != nil {
		return *in
	}
	return ""
}

func Bool(in bool) *bool {
	return &in
}

func Int32(in int32) *int32 {
	return &in
}

func Int64(in int64) *int64 {
	return &in
}
