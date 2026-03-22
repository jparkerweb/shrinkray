package engine

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"syscall"
)

// openFolderCmd returns an exec.Cmd that opens explorer with /select.
// We set SysProcAttr.CmdLine directly to bypass Go's argument escaping,
// which mangles the /select,"path" syntax that explorer requires.
func openFolderCmd(absPath string) *exec.Cmd {
	cmd := exec.Command("explorer.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: fmt.Sprintf(`explorer.exe /select,"%s"`, absPath),
	}
	return cmd
}

// OpenFolderCommand returns the command that would be used to open the folder,
// without executing it. Useful for testing.
func OpenFolderCommand(path string) (name string, args []string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	return "explorer.exe", []string{fmt.Sprintf(`/select,"%s"`, absPath)}
}
