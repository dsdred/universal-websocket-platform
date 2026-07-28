package runtime

// Launch delegates one prepared Runtime launch to Bootstrap.
func Launch(request *BootstrapRequest) BootstrapOutcome {
	return Bootstrap(request)
}
