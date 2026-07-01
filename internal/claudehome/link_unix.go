//go:build !windows

package claudehome

import (
	"fmt"
	"os"
)

// link creates dst as a symlink to src. On unix, plain symlinks need no
// elevated privilege for either files or directories.
func link(src, dst string) error {
	if err := os.Symlink(src, dst); err != nil {
		return fmt.Errorf("symlink %s -> %s: %w", dst, src, err)
	}
	return nil
}
