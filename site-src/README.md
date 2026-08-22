# site-src

Build inputs for abcdev.app. Composition comes from `.abcd/site.json`; the text
comes from the repository. This directory holds:

- `ui.json` — the interface-string allowlist: every word the generator may add.
- `install.sh.tmpl` — the committed source of `/install.sh`.
- `redirects` — the `_redirects` source.
- `headers` — the `_headers` source.
