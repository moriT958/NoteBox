package cli

func validateArgs(args []string) bool {
	l := len(args)
	if l <= 1 {
		return true
	} else {
		return false
	}
}
