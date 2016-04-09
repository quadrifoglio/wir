package main

import (
	"os"
	"os/exec"
	"strconv"
)

func QemuStart(vm *Vm) error {
	m := strconv.Itoa(vm.Params.Memory)
	c := strconv.Itoa(vm.Params.Cores)

	cmd := exec.Command("qemu-system-x86_64", "-m", m, "-smp", c)

	err := cmd.Start()
	if err != nil {
		return err
	}

	err = DatabaseSetVmAttr(vm, "pid", strconv.Itoa(cmd.Process.Pid))
	if err != nil {
		return err
	}

	return nil
}

func QemuStatus(vm *Vm) error {
	return ErrBackend
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
		return ErrProcessNotFound
	}

	err = proc.Kill()
	if err != nil {
		return ErrKill
	}

	return ErrBackend
}
