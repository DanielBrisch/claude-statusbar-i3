package notify

type backend struct {
	name string
	args []string
}

func newBackend(name string, args ...string) backend {
	return backend{name: name, args: args}
}
