package tui

import "testing"

func TestResolveThemeAliases(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "default empty", in: "", want: "default"},
		{name: "gruvbox exact", in: "gruvbox", want: "gruvbox"},
		{name: "gruvbox normalized", in: "Gruv_Box", want: "gruvbox"},
		{name: "catppuccin alias", in: "catppucin", want: "catppuccin"},
		{name: "unknown falls back", in: "solarized", want: "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveTheme(tt.in).Name; got != tt.want {
				t.Fatalf("ResolveTheme(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestApplyThemeUpdatesCurrentTheme(t *testing.T) {
	original := CurrentThemeName()
	t.Cleanup(func() {
		ApplyTheme(original)
	})

	if got := ApplyTheme("gruvbox"); got != "gruvbox" {
		t.Fatalf("ApplyTheme returned %q, want gruvbox", got)
	}
	if got := CurrentThemeName(); got != "gruvbox" {
		t.Fatalf("CurrentThemeName = %q, want gruvbox", got)
	}
}
