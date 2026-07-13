//go:build unix

package audit

import "syscall"

// syscallNoFollow supplements the os.Root containment on the tracked-file open.
// Repo containment (no component escapes the root, symlinks included) is os.Root's
// job; these flags harden the leaf open itself: O_NONBLOCK makes opening a FIFO or
// device return immediately instead of blocking until a writer appears, so a
// tracked FIFO cannot hang the audit at open() before the regular-file check runs;
// O_NOFOLLOW additionally refuses a leaf symlink outright. O_NONBLOCK is a no-op on
// a regular file. Release targets are darwin and linux (both unix); a build on
// another platform would need its own definition.
const syscallNoFollow = syscall.O_NOFOLLOW | syscall.O_NONBLOCK
