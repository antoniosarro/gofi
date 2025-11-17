package emoji

import (
	"fmt"
	"os/exec"
)

// CopyToClipboard copies the emoji to clipboard using wl-copy or xclip
func CopyToClipboard(emoji string) error {
	// Try wl-copy first (Wayland)
	if err := copyWithWlCopy(emoji); err == nil {
		return nil
	}

	// Fall back to xclip (X11)
	if err := copyWithXclip(emoji); err == nil {
		return nil
	}

	return fmt.Errorf("no clipboard tool available (tried wl-copy, xclip)")
}

func copyWithWlCopy(text string) error {
	cmd := exec.Command("wl-copy", text)
	return cmd.Run()
}

func copyWithXclip(text string) error {
	cmd := exec.Command("xclip", "-selection", "clipboard")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if _, err := stdin.Write([]byte(text)); err != nil {
		return err
	}

	stdin.Close()
	return cmd.Wait()
}
