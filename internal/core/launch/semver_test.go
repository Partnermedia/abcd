package launch

import "testing"

func TestIsStrictSemver(t *testing.T) {
	valid := []string{"0.0.0", "1.2.3", "10.20.30", "1.0.0-alpha", "1.0.0-alpha.1", "1.0.0+build", "1.0.0-rc.1+exp.sha.5"}
	for _, v := range valid {
		if !IsStrictSemver(v) {
			t.Errorf("expected %q valid", v)
		}
	}
	invalid := []string{"v1.2.3", "1.2", "1.2.3.4", "01.2.3", "1.02.3", "1.2.3rc1", "1.2.3\n", "", "1.2.3 "}
	for _, v := range invalid {
		if IsStrictSemver(v) {
			t.Errorf("expected %q invalid", v)
		}
	}
}

func TestParseSemver(t *testing.T) {
	s, err := ParseSemver("1.2.3-rc.1+b.2")
	if err != nil {
		t.Fatal(err)
	}
	if s.Major != 1 || s.Minor != 2 || s.Patch != 3 || s.Prerelease != "rc.1" || s.Build != "b.2" {
		t.Errorf("bad parse: %+v", s)
	}
	if s.Line() != "1.2" {
		t.Errorf("Line=%q want 1.2", s.Line())
	}
	if s.Tag() != "v1.2.3" {
		t.Errorf("Tag=%q want v1.2.3", s.Tag())
	}
	if _, err := ParseSemver("1.2.3\n"); err == nil {
		t.Error("trailing newline must be rejected")
	}
}
