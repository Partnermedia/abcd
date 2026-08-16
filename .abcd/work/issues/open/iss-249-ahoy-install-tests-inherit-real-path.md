---
schema_version: 1
id: "iss-249"
slug: "ahoy-install-tests-inherit-real-path"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "academic-references-baseline"
found_at: "internal/core/ahoy/dev_install_test.go"
---

ahoy install-mode tests inherit the real PATH: with an installed ~/.local/bin/abcd present, TestDevInstallWritesShimAndSurfacesMode, TestNormalInstallRegressionPin, TestInstallPinnedToDevTransition and TestInstallDevToPinnedTransition fail with 'shadowed on PATH' because the developer's live install shadows the test's; the tests pass when ~/.local/bin is removed from PATH, so the suite is red on any machine that has abcd installed the documented way