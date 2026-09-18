package render

import (
	"strings"
	"testing"
)

func TestRegistryKnowsEveryFormat(t *testing.T) {
	reg := NewRegistry(opts())
	want := []string{"i3blocks", "json", "plain", "polybar", "waybar"}

	got := reg.Names()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("Names() = %v, want %v", got, want)
	}
	for _, name := range want {
		if _, err := reg.Lookup(name); err != nil {
			t.Errorf("Lookup(%q): %v", name, err)
		}
	}
}

func TestRegistryNamesTheAlternativesWhenAskedForSomethingItDoesNotHave(t *testing.T) {
	_, err := NewRegistry(opts()).Lookup("dzen2")
	if err == nil {
		t.Fatal("Lookup of an unknown format returned nil error")
	}
	if !strings.Contains(err.Error(), "dzen2") || !strings.Contains(err.Error(), "i3blocks") {
		t.Errorf("error %q should name both the bad format and the valid ones", err)
	}
}
