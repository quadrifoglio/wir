package inter

import (
	"fmt"
	"os/exec"
)

func Scp(src, dstAddr, dstFile string) error {
	dstf := fmt.Sprintf("%s:%s", dstAddr, dstFile)
	cmd := exec.Command("scp", "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", src, dstf)

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
