---
schema_version: 1
id: "iss-2608210738378295"
slug: "local-gates-lint-working-tree-not-commit"
severity: "minor"
category: "future-work-seed"
source: "review-followup"
found_during: "itd-130 session; #395 renumber"
---

Local record gates validate the WORKING TREE, not the committed/pushed tree, so a partial commit passes them while the committed tree fails CI. Demonstrated live this session: git mv staged a rename (iss-370->377) but the follow-up sed edit to the frontmatter id was left unstaged; git commit captured only the rename, so the committed file had filename iss-377 with frontmatter iss-370 — a record_schema blocker. record-lint had passed at commit time because it reads the working tree (which held the sed edit), and the pre-push make preflight passes for the same reason (the unstaged edit is still present); only CI, checking out the commit, sees the divergence. Fix direction: run the push-time gate against the tree that will actually ship (e.g. a clean worktree of HEAD, or git stash --keep-index before linting), so a working-tree/index divergence cannot pass locally and fail CI. Sibling of iss-147 (guard-load reads guard.json from the working tree).