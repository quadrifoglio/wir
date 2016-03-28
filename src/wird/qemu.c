#include "wird/qemu.h"

#include <stdio.h>
#include <stdlib.h>
#include <stdarg.h>
#include <unistd.h>
#include <string.h>
#include <errno.h>
#include <signal.h>

const char* executable = "/usr/bin/qemu-system-x86_64";

int exec(int argc, char** argv, pid_t* tpid) {
	char** args = malloc((argc + 1) * sizeof(char**));
	memcpy(args, argv, argc * sizeof(char*));
	args[argc] = 0;

	errno = 0;
	pid_t pid = fork();
	if(pid == 0) { // Child process
		if(execv(executable, args) < 0) {
			perror("exec: ");
			exit(1);
		}
	}
	else if(pid < 0) {
		return ERREXEC;
	}

	*tpid = pid;
	free(args);
	free(argv);

	return ERRNOPE;
}

void arg_add(int* argc, char*** argv, int n, ...) {
	int index = *argc;
	*argv = realloc(*argv, (*argc += n) * sizeof(char*));

	va_list v;
	va_start(v, n);

	for(int i = 0; i < n; ++i) {
		*(*argv + index++) = va_arg(v, char*);
	}

	va_end(v);
}

int qemu_start(vm_t* vm) {
	int argc = 0;
	char** argv = 0;

	char* cpus;
	char* mems;
	asprintf(&cpus, "%d", vm->params.ncpu);
	asprintf(&mems, "%d", vm->params.memory);

	arg_add(&argc, &argv, 6, "qemu-system-x86_64", "-enable-kvm", "-smp", cpus, "-m", mems);

	vm_dev_t d = {DEV_CDROM, "/home/quadrifoglio/vm/debian.iso"};
	vm->params.devices = &d;
	vm->params.device_count = 1;

	for(int i = 0; i < vm->params.device_count; ++i) {
		vm_dev_t dev = vm->params.devices[i];

		if(dev.type == DEV_HDD && dev.file != 0) {
			arg_add(&argc, &argv, 2, "-hda", dev.file);
		}
		else if(dev.type == DEV_CDROM && dev.file != 0) {
			arg_add(&argc, &argv, 2, "-cdrom", dev.file);
		}
	}

	pid_t pid;
	int err = exec(argc, argv, &pid);
	if(err != ERRNOPE) {
		free(cpus);
		free(mems);
		return err;
	}

	vm->backend_data = malloc(sizeof(pid_t));
	*((pid_t*)vm->backend_data) = pid;

	free(cpus);
	free(mems);

	return ERRNOPE;
}

int qemu_stop(vm_t* vm, bool violent) {
	if(!vm->backend_data) {
		return ERRDOWN;
	}

	pid_t pid = *((pid_t*)vm->backend_data);
	kill(pid, violent ? SIGKILL : SIGTERM);

	return ERRNOPE;
}
