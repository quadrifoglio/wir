package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

func QemuSetupImage(vm *Vm) error {
	var err error
	var img Image
	var state string = ImgStateLoading

	for state == ImgStateLoading {
		img, err = DatabaseGetImage(vm.Params.ImageID)
		if err != nil {
			log.Println(err)
			return ErrImageNotFound
		}

		state = img.State
		time.Sleep(1 * time.Second)
	}

	id, err := DatabaseFreeVmId()
	if err != nil {
		return err
	}

	dir := Config.DrivesDir + "/" + strconv.Itoa(id) + "/"
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	path := dir + filepath.Base(img.Path) + ".img"

	// If this is a migration, we should not create a new drive, it will already be there
	if !vm.Params.Migration {
		cmd := exec.Command("qemu-img", "create", "-b", img.Path, "-f", "qcow2", path)
		if err != nil {
			return err
		}

		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Println(string(out))
			return err
		}
	}

	vm.Drives = make([]VmDrive, 1)
	vm.Drives[0] = VmDrive{"hdd", path}

	return nil
}

func QemuStart(vm *Vm) error {
	vm.State = VmStateDown

	args := make([]string, 5)
	args[0] = "-enable-kvm"
	args[1] = "-m"
	args[2] = strconv.Itoa(vm.Params.Memory)
	args[3] = "-smp"
	args[4] = strconv.Itoa(vm.Params.Cores)

	for _, d := range vm.Drives {
		switch d.Type {
		case DriveHDD:
			args = append(args, "-hda")
			args = append(args, d.File)
			break
		case DriveCDROM:
			args = append(args, "-cdrom")
			args = append(args, d.File)
			break
		}
	}

	if len(vm.Params.NetBridgeOn) > 0 {
		err := NetCreateBridge("wir0")
		if err != nil {
			return err
		}

		NetDeleteTAP("tapvm" + strconv.Itoa(vm.ID))

		err = NetCreateTAP("tapvm" + strconv.Itoa(vm.ID))
		if err != nil {
			return err
		}

		err = NetBridgeAddIf("wir0", "tapvm"+strconv.Itoa(vm.ID))
		if err != nil {
			return err
		}

		err = NetBridgeAddIf("wir0", vm.Params.NetBridgeOn)
		if err != nil {
			return err
		}

		args = append(args, "-netdev")
		args = append(args, fmt.Sprintf("tap,id=net0,ifname=%s,script=no", "tapvm"+strconv.Itoa(vm.ID)))
		args = append(args, "-device")
		args = append(args, "driver=virtio-net,netdev=net0")
	}

	cmd := exec.Command("qemu-system-x86_64", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	go func() {
		in := bufio.NewScanner(stdout)
		for in.Scan() {
			log.Printf("qemu vm %d: %s\n", vm.ID, in.Text())
		}
	}()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	go func() {
		in := bufio.NewScanner(stderr)
		for in.Scan() {
			log.Printf("qemu vm %d: %s\n", vm.ID, in.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		return err
	}

	errc := make(chan bool)

	go func() {
		err := cmd.Wait()

		var errs string
		if err != nil {
			errs = err.Error()
		} else {
			errs = "exit status 0"
		}

		log.Printf("qemu vm %d: process exited: %s", vm.ID, errs)
		errc <- true
	}()

	time.Sleep(500 * time.Millisecond)

	select {
	case <-errc:
		return ErrStart
	default:
		vm.State = VmStateUp
		break
	}

	err = DatabaseSetVmAttr(vm, "pid", strconv.Itoa(cmd.Process.Pid))
	if err != nil {
		return err
	}

	return nil
}

func QemuStatus(vm *Vm) error {
	if vm.Attrs == nil {
		vm.State = VmStateDown
		return nil
	}

	v := vm.Attrs["pid"]
	if len(v) == 0 {
		vm.State = VmStateDown
		return nil
	}

	pid, err := strconv.Atoi(v)
	if err != nil {
		vm.State = VmStateDown
		return ErrInvalidAttrType
	}

	out, err := exec.Command("kill", "-s", "0", strconv.Itoa(pid)).CombinedOutput()
	if err != nil {
		vm.State = VmStateDown
		return nil
	}

	if string(out) == "" {
		vm.State = VmStateUp
		return nil
	}

	vm.State = VmStateDown
	return nil
}

func QemuStop(vm *Vm) error {
	if vm.Attrs == nil {
		return ErrNoAttrs
	}

	v := vm.Attrs["pid"]
	if len(v) == 0 {
		return ErrNoAttrs
	}

	pid, err := strconv.Atoi(v)
	if err != nil {
		return ErrInvalidAttrType
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		vm.State = VmStateDown
		return nil
	}

	err = proc.Kill()
	if err != nil {
		return ErrKill
	}

	NetDeleteTAP("tapvm" + strconv.Itoa(vm.ID))

	vm.State = VmStateDown
	return nil
}
