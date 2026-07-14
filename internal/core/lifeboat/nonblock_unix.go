//go:build unix

package lifeboat

import "syscall"

// nonBlock hardens the probe's contained file reads, matching the audit
// tracked-file open. os.Root already contains every read to the repo (no
// component escapes the root, symlinked intermediates included); these flags
// harden the leaf open: O_NONBLOCK makes opening a FIFO or device return
// immediately instead of blocking until a writer appears, so a probed tree
// cannot hang the probe at open() before the regular-file check; O_NOFOLLOW
// refuses a leaf symlink outright. Both are no-ops on a regular file. Release
// targets are darwin and linux (both unix); another platform needs its own
// definition.
const nonBlock = syscall.O_NOFOLLOW | syscall.O_NONBLOCK
