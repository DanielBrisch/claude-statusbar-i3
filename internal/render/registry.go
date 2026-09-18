package render

import (
	"fmt"
	"sort"
	"strings"
)

type Registry struct {
	renderer *Renderer
	builders map[string]func(*Renderer) Formatter
}

func NewRegistry(o Options) *Registry {
	return &Registry{
		renderer: NewRenderer(o),
		builders: map[string]func(*Renderer) Formatter{
			"i3blocks": func(r *Renderer) Formatter { return NewI3blocks(r) },
			"waybar":   func(r *Renderer) Formatter { return NewWaybar(r) },
			"polybar":  func(r *Renderer) Formatter { return NewPolybar(r) },
			"plain":    func(r *Renderer) Formatter { return NewPlain(r) },
			"json":     func(r *Renderer) Formatter { return NewJSON(r) },
		},
	}
}

func (reg *Registry) Names() []string {
	names := make([]string, 0, len(reg.builders))
	for name := range reg.builders {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (reg *Registry) Lookup(name string) (Formatter, error) {
	build, ok := reg.builders[name]
	if !ok {
		return nil, fmt.Errorf("unknown format %q (want %s)", name, strings.Join(reg.Names(), ", "))
	}
	return build(reg.renderer), nil
}

func (reg *Registry) Renderer() *Renderer { return reg.renderer }
