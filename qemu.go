package main

import (
	"os"
	"os/exec"
	"strconv"
)

func QemuStart(vm *Vm) error {
	args := make([]string, 4)
	args[0] = "-m"
	args[1] = strconv.Itoa(vm.Params.Memory)
	args[2] = "-smp"
	args[3] = strconv.Itoa(vm.Params.Cores)

	for _, d := range vm.Params.Drives {
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

	err := cmd.Start()
	if err != nil {
		return err
	}

	err = DatabaseSetVmAttr(vm, "pid", strconv.Itoa(cmd.Process.Pid))
	if err != nil {
		return err
	}

	vm.State = StateUp
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

	_, err = proc.Wait()
	if err != nil {
		return ErrKill
	}

	vm.State = StateDown
	return nil
}
