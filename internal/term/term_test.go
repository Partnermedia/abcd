package term

import "testing"

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

// TestResolveColorModePrecedence pins the ruled ladder precedence
// (itd-112): --no-color > NO_COLOR (present and non-empty) > TERM
// dumb/unset > COLORTERM > 256color TERM > 16-colour floor.
func TestResolveColorModePrecedence(t *testing.T) {
	cases := []struct {
		name    string
		env     map[string]string
		noColor bool
		want    ColorMode
	}{
		{"flag beats everything", map[string]string{"COLORTERM": "truecolor", "TERM": "xterm-256color"}, true, Mono},
		{"NO_COLOR non-empty wins over truecolor", map[string]string{"NO_COLOR": "1", "COLORTERM": "truecolor", "TERM": "xterm-256color"}, false, Mono},
		{"NO_COLOR empty string does NOT count", map[string]string{"NO_COLOR": "", "TERM": "xterm-256color"}, false, Ansi256},
		{"TERM dumb forces mono despite leaked COLORTERM", map[string]string{"TERM": "dumb", "COLORTERM": "truecolor"}, false, Mono},
		{"TERM unset means mono", map[string]string{"COLORTERM": "truecolor"}, false, Mono},
		{"COLORTERM truecolor", map[string]string{"TERM": "xterm", "COLORTERM": "truecolor"}, false, TrueColor},
		{"COLORTERM 24bit", map[string]string{"TERM": "xterm", "COLORTERM": "24bit"}, false, TrueColor},
		{"256color TERM", map[string]string{"TERM": "screen-256color"}, false, Ansi256},
		{"plain TERM floors at 16", map[string]string{"TERM": "xterm"}, false, Ansi16},
	}
	for _, c := range cases {
		if got := ResolveColorMode(env(c.env), c.noColor); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestUTF8Locale(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{"LANG utf8", map[string]string{"LANG": "en_GB.UTF-8"}, true},
		{"LC_ALL wins and is not utf8", map[string]string{"LC_ALL": "C", "LANG": "en_GB.UTF-8"}, false},
		{"LC_CTYPE utf8 variant spelling", map[string]string{"LC_CTYPE": "de_DE.utf8"}, true},
		{"no locale variables", map[string]string{}, false},
	}
	for _, c := range cases {
		if got := UTF8Locale(env(c.env)); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
