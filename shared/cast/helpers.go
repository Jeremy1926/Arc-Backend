package cast

func Str(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func Bool(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}
