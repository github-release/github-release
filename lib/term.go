package lib

func Mark(ok bool) string {
	if ok {
		return "✔"
	} else {
		return "✗"
	}
}
