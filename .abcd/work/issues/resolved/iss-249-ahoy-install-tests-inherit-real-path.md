---
schema_version: 1
id: "iss-249"
slug: "ahoy-install-tests-inherit-real-path"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "academic-references-baseline"
found_at: "internal/core/ahoy/dev_install_test.go"
resolution: "setupHermetic now pins PATH alongside HOME and the bin target: it filters out of the inherited PATH every directory that resolves an abcd, keeping the utilities the code under test shells out to. The four dev-install tests pass unmodified on a machine with a real abcd on PATH — previously they failed there with install_mode carrying '(shadowed on PATH)', which is the detector honestly reporting the developer's own install leaking into the sandbox."
impact: internal
---

ahoy install-mode tests inherit the real PATH: with an installed ~/.local/bin/abcd present, TestDevInstallWritesShimAndSurfacesMode, TestNormalInstallRegressionPin, TestInstallPinnedToDevTransition and TestInstallDevToPinnedTransition fail with 'shadowed on PATH' because the developer's live install shadows the test's; the tests pass when ~/.local/bin is removed from PATH, so the suite is red on any machine that has abcd installed the documented way