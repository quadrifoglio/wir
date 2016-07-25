package inter

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/quadrifoglio/wir/global"
)

func RemoteMkdir(dst global.Remote, dstDir string) error {
	cmd := exec.Command("ssh", fmt.Sprintf("%s@%s", dst.SSHUser, dst.Addr), "mkdir -p "+dstDir)

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("RemoteMkdir:", string(out))
		return err
	}

	return nil
}

func SCP(src string, dst global.Remote, dstFile string) error {
	dstf := fmt.Sprintf("%s:%s", dst.Addr, dstFile)
	cmd := exec.Command("scp", "-r", src, dstf)

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("SCP:", string(out))
		return err
	}

	return nil
}
