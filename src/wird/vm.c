#include "wird/vm.h"
#include "wird/db.h"

int vm_create(vm_params_t* p, vm_t* vm) {
	int id = 0;
	int err = db_insert_vm(p, &id);
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

int vm_delete(vm_t* vm) {
	return ERRNOPE;
}
