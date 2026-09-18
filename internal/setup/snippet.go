package setup

import "fmt"

type Snippet struct {
	bar    string
	binary string
}

func NewSnippet(bar, binary string) Snippet {
	return Snippet{bar: bar, binary: binary}
}

func (s Snippet) String() string {
	switch s.bar {
	case "waybar":
		return fmt.Sprintf(`Add to ~/.config/waybar/config:

  "custom/claude": {
    "exec": "%s render --format waybar",
    "return-type": "json",
    "interval": 10,
    "on-click": "%[1]s detail --notify"
  }
`, s.binary)
	case "polybar":
		return fmt.Sprintf(`Add to ~/.config/polybar/config.ini:

  [module/claude]
  type = custom/script
  exec = %s render --format polybar
  interval = 10
  format = <label>
`, s.binary)
	default:
		return fmt.Sprintf(`Add to ~/.config/i3blocks/config:

  [claude]
  command=%s render --format i3blocks --markup pango --icon-size x-large
  markup=pango
  interval=10

Then restart i3 (i3-msg restart) so i3blocks re-reads its config.

markup=pango is what lets the icon be drawn larger than the rest of the block.
Drop both the property and the flag if you would rather keep it plain, and use
--icon-size to change how much larger it gets.
`, s.binary)
	}
}
