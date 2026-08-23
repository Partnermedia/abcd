---
schema_version: 1
id: "iss-2608230752158970"
slug: "site-no-print-stylesheet"
severity: "minor"
category: "ux"
source: "user-observation"
found_during: "user-observation"
found_at: "site-src/site.css"
---

Record pages print badly: no @media print stylesheet, so interactive chrome prints, collapsed details lose their content on paper, Foundations page 1 is blank (panel refuses to split), cards slice across page boundaries, and footers orphan onto empty pages (report 18, three PDFs).