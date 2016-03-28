#include "wird/vm.h"
#include "wird/db.h"

int vm_create(vm_params_t* p, vm_t* vm) {
	int id = 0;
	int err = db_vm_insert(p, &id);
	if(err != ERRNOPE) {
		return err;
	}

	if(vm) {
		vm->id = id;
		vm->state = STATE_DOWN;
		vm->params = *p;
	}

	return ERRNOPE;
}

int vm_list(vm_t** vms, int* count) {
	return db_vm_list(vms, count);
}

int vm_delete(vm_t* vm) {
	return db_vm_delete(vm->id);
}

const char* vm_backend_str(vm_backend_t b) {
	switch(b) {
		case BACKEND_QEMU:
			return "qemu";
		case BACKEND_VZ:
			return "openvz";
		default:
			return "unknown";
	}
}

const char* vm_dev_str(vm_dev_type_t d) {
	switch(d) {
		case DEV_HDD:
			return "hard drive";
		case DEV_CDROM:
			return "cdrom drive";
		default:
			return "unknown drive";
	}
}

const char* vm_state_str(vm_state_t s) {
	switch(s) {
		case STATE_DOWN:
			return "down";
		case STATE_UP:
			return "up";
		default:
			return "unknown";
	}
}
