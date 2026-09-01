package ahoy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPluginDataDirResolvesEnvThenRootStamp pins the resolver's ladder
// (iss-2609012111168716): a hook's CLAUDE_PLUGIN_DATA wins as given; from a
// terminal the plugin root's .data-dir stamp answers, but only with an
// absolute path naming an existing directory; and every miss carries a story
// that names what was consulted, so the degradation note can repeat it.
func TestPluginDataDirResolvesEnvThenRootStamp(t *testing.T) {
	root := t.TempDir()
	data := t.TempDir()
	stamp := filepath.Join(root, dataDirStampFile)
	writeStamp := func(recorded string) {
		t.Helper()
		if err := os.WriteFile(stamp, []byte("data_dir="+recorded+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Setenv("CLAUDE_PLUGIN_DATA", "/from/hook")
	writeStamp(data)
	if got := pluginDataDir(root); got.dir != "/from/hook" || !strings.Contains(got.story, "CLAUDE_PLUGIN_DATA") {
		t.Errorf("the environment must win over the stamp: %+v", got)
	}

	t.Setenv("CLAUDE_PLUGIN_DATA", "")
	if got := pluginDataDir(root); got.dir != data {
		t.Errorf("with no environment the stamp must answer: %+v", got)
	}

	if got := pluginDataDir(""); got.dir != "" || !strings.Contains(got.story, "no plugin root") {
		t.Errorf("no plugin root means no stamp to read, and the story must say so: %+v", got)
	}

	if err := os.Remove(stamp); err != nil {
		t.Fatal(err)
	}
	if got := pluginDataDir(root); got.dir != "" || !strings.Contains(got.story, dataDirStampFile) {
		t.Errorf("an absent stamp is a miss that names the stamp: %+v", got)
	}

	for name, recorded := range map[string]string{
		"relative":     "relative/data",
		"absent":       filepath.Join(t.TempDir(), "gone"),
		"regular file": filepath.Join(root, dataDirStampFile),
	} {
		writeStamp(recorded)
		got := pluginDataDir(root)
		if got.dir != "" {
			t.Errorf("%s: a recorded path that is not an existing absolute directory must not be followed: %+v", name, got)
		}
		if !strings.Contains(got.story, "CLAUDE_PLUGIN_DATA is unset") || !strings.Contains(got.story, dataDirStampFile) {
			t.Errorf("%s: the story must name both sources: %+v", name, got)
		}
	}

	if err := os.WriteFile(stamp, []byte("release_tag=v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := pluginDataDir(root); got.dir != "" {
		t.Errorf("a stamp with no data_dir line records nothing: %+v", got)
	}

	// The explanation for a directory that resolved but holds no cache names the
	// directory and the gap, never a bare "unavailable".
	writeStamp(data)
	if why := pluginDataDir(root).explainMissingCache(); !strings.Contains(why, "recorded checksum") || !strings.Contains(why, dataDirStampFile) {
		t.Errorf("explainMissingCache must say the resolved directory holds no verified artefact: %q", why)
	}
}
