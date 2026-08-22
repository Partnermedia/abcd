# site-src

Build inputs for abcdev.app. Composition comes from `.abcd/site.json`; the text
comes from the repository. This directory holds:

- `ui.json` — the interface-string allowlist: every word the generator may add.
  It is decoded against a closed struct with unknown fields refused, so a word
  added here that no field reads fails the build.
- `site.css` — the stylesheet: light is the base and dark redefines tokens only.
  The landing page follows the reader's system preference; a page that sets an
  explicit `data-theme` attribute is honoured too, and the record explorer is
  the slice that ships the control which sets one.
- `site.js` — the landing page's only script: install tabs, copy buttons, and
  nothing else. It adds no words of its own — every string it shows is read back
  from the markup, where the build put it from `ui.json`. No analytics, no
  trackers, no network requests.
- `install.sh.tmpl` — the committed source of `/install.sh`.
- `redirects` — the `_redirects` source.
- `headers` — the `_headers` source.

`abcd site build` copies `site.css`, `site.js`, `record.js`, `redirects` and
`headers` into the output tree (the last two under their served names), renders
`install.sh.tmpl` to `install.sh` with one build-stamp comment added, and reads
`ui.json`.
