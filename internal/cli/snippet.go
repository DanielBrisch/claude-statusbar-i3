package cli

import (
	"fmt"

	"github.com/DanielBrisch/claude-usage-status-i3/internal/setup"
)

var installStatusLine = setup.InstallStatusLine

func snippet(bar string) string {
	switch bar {
	case "waybar":
		return fmt.Sprintf(`Add to ~/.config/waybar/config:

  "custom/claude": {
    "exec": "%s render --format waybar",
    "return-type": "json",
    "interval": 10,
    "on-click": "%[1]s detail --notify"
  }
`, binaryName())
	case "polybar":
		return fmt.Sprintf(`Add to ~/.config/polybar/config.ini:

  [module/claude]
  type = custom/script
  exec = %s render --format polybar
  interval = 10
  format = <label>
`, binaryName())
	default:
		return fmt.Sprintf(`Add to ~/.config/i3blocks/config:

  [claude]
  command=%s render --format i3blocks
  interval=10

Then reload i3blocks (restart i3, or re-read the config).
`, binaryName())
	}
}
