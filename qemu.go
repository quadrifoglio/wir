package main

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

func QemuSetupImage(vm *Vm) error {
	img, err := DatabaseGetImage(vm.Params.ImageID)
	if err != nil {
		log.Println(err)
		return ErrImageNotFound
	}

	id, err := DatabaseFreeVmId()
	if err != nil {
		return err
	}

	dir := DrivesDir + "/" + strconv.Itoa(id) + "/"
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	path := dir + filepath.Base(img.Path) + ".img"

	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", "-o", "backing_file="+img.Path, path)
	if err != nil {
		return err
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return err
	}

	vm.Drives = make([]VmDrive, 1)
	vm.Drives[0] = VmDrive{"hdd", path}

	return nil
}

func QemuStart(vm *Vm) error {
	vm.State = StateDown

	args := make([]string, 4)
	args[0] = "-m"
	args[1] = strconv.Itoa(vm.Params.Memory)
	args[2] = "-smp"
	args[3] = strconv.Itoa(vm.Params.Cores)

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
		vm.State = StateUp
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
		vm.State = StateDown
		return nil
	}

	v := vm.Attrs["pid"]
	if len(v) == 0 {
		vm.State = StateDown
		return nil
	}

	pid, err := strconv.Atoi(v)
	if err != nil {
		vm.State = StateDown
		return ErrInvalidAttrType
	}

	out, err := exec.Command("kill", "-s", "0", strconv.Itoa(pid)).CombinedOutput()
	if err != nil {
		vm.State = StateDown
		return nil
	}

	if string(out) == "" {
		vm.State = StateUp
		return nil
	}

	vm.State = StateDown
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
		vm.State = StateDown
		return nil
	}

	err = proc.Kill()
	if err != nil {
		return ErrKill
	}

	vm.State = StateDown
	return nil
}
