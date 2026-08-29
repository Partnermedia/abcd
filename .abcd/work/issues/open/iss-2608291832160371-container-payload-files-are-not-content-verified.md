---
schema_version: 1
id: "iss-2608291832160371"
slug: "container-payload-files-are-not-content-verified"
severity: "major"
category: "security"
source: "impl-review"
found_during: "v0.6.9-security-pass"
found_at: "internal/adapter/scanner/scanner.go"
---
GHSA-9wv7 residual: payload files in a compressed or container format cannot be content-verified by the byte scan. That is every archive (.zip .gz .tgz .tar .bz2 .xz .7z) and equally every media format on the skip list whose payload is a compressed stream — a PNG (deflate IDAT and zTXt: a valid 1x1 PNG can carry a token in zTXt with not even its prefix visible in the raw bytes), a JPEG entropy stream, a PDF FlateDecode stream, a GIF LZW stream, WebP, and the mp4/mov/webm/mp3 boxes and frames. They are byte-scanned (plaintext regions such as PNG tEXt or an uncompressed tar's entries are covered) and reported as content_unverified, not refused. Decide whether the launch gate should refuse them in the payload or decode known formats under the scan cap and scan their entries. Any refusal must key on magic bytes, not the extension: the current label keys on the name, so a zip renamed .png is labelled by its name.
