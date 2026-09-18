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
  command=%s render --format i3blocks --markup pango
  markup=pango
  interval=10

Then restart i3 (i3-msg restart) so i3blocks re-reads its config.

markup=pango is what lets the icon be drawn larger than the rest of the block.
Drop both the property and the flag if you would rather keep it plain, and use
--icon-size to change how much larger it gets.
`, binaryName())
	}
}
