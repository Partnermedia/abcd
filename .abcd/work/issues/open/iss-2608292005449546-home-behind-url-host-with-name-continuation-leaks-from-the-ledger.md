---
schema_version: 1
id: "iss-2608292005449546"
slug: "home-behind-url-host-with-name-continuation-leaks-from-the-ledger"
severity: "minor"
category: "security"
source: "impl-review"
found_during: "ultra-v0.6.8-followup"
found_at: "internal/adapter/scanner/residual.go"
---

ultra-v0.6.8 follow-up (confirmation pass): a home behind a URL host with a '.' or '-' continuation — https://ci.example.com/Users/<user>.zip, /Users/<user>-old — is committed verbatim by capture and intent, which have no store-side backstop: home_path_self declines on nameContinues, home_path_other declines at its leading boundary (the host's last letter), and local_username is URL-suppressed, so nothing reports it. Pre-existing on main. The branch's tests pin only the '/'-trailing shape behind a host (internal/adapter/scanner/residual_test.go TestHomeBehindAURLHostIsSweptNotRefused). Memory and history are covered by the rewriting backstop for /Users and /home homes only.
