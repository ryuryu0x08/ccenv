//go:build windows

package claudehome

import (
	"fmt"
	"os"
	"os/exec"
)

// link creates dst as a link to src using the link type that needs no
// elevation/Developer Mode on Windows: a directory junction (mklink /J) for
// directories, a hard link (os.Link) for files. Plain symlinks are avoided
// because creating them requires SeCreateSymbolicLinkPrivilege or Developer
// Mode, neither of which ccenv should assume is available.
func link(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat %s: %w", src, err)
	}
	if !info.IsDir() {
		if err := os.Link(src, dst); err != nil {
			return fmt.Errorf("hardlink %s -> %s: %w", dst, src, err)
		}
		return nil
	}
	cmd := exec.Command("cmd", "/c", "mklink", "/J", dst, src)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mklink /J %s -> %s: %w (%s)", dst, src, err, out)
	}
	return nil
}
