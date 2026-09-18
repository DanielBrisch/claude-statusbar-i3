package cli

type commands struct {
	byName map[string]Command
}

func newCommands(a *App) commands {
	c := commands{byName: map[string]Command{}}
	c.byName["collect"] = NewCollectCommand(a)
	c.byName["render"] = NewRenderCommand(a)
	c.byName["detail"] = NewDetailCommand(a)
	c.byName["init"] = NewInitCommand(a)
	c.byName["doctor"] = NewDoctorCommand(a)

	version := NewVersionCommand(a)
	for _, alias := range []string{"version", "--version", "-v"} {
		c.byName[alias] = version
	}
	help := NewHelpCommand(a)
	for _, alias := range []string{"help", "--help", "-h"} {
		c.byName[alias] = help
	}
	return c
}

func (c commands) lookup(name string) (Command, bool) {
	command, ok := c.byName[name]
	return command, ok
}

func (c commands) help() *HelpCommand {
	return c.byName["help"].(*HelpCommand)
}
