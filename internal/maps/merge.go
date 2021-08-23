package maps

// Merge performs a deep merge of two maps.
// It is used to merge the additional parameters in the policies.
func Merge(current, overwrite map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(current))
	for k, v := range current {
		out[k] = v
	}
	for k, v := range overwrite {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = Merge(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
