package inter

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/quadrifoglio/wir/client"
)

func RemoteMkdir(dst client.Remote, dstDir string) error {
	cmd := exec.Command("ssh", fmt.Sprintf("%s@%s", dst.SSHUser, dst), "mkdir -p "+dstDir)

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("RemoteMkdir: %s", string(out))
		return err
	}

	return nil
}

func SCP(src string, dst client.Remote, dstFile string) error {
	dstf := fmt.Sprintf("%s:%s", dst.Addr, dstFile)
	cmd := exec.Command("scp", "-r", src, dstf)

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("SCP: %s", string(out))
		return err
	}

	return nil
}
