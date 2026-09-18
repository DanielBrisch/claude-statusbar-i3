package cli

type Command interface {
	Run(args []string) (int, error)
}
