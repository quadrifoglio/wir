#include "wird/vm.h"

#include <string.h>
#include <stdlib.h>
#include "wird/db.h"

int vm_image_list(vm_image_t** imgs, int* count) {
	return db_image_list(imgs, count);
}

int vm_image_json(vm_image_t* img, JSON_Value** v) {
	*v = json_value_init_object();
	JSON_Object* obj = json_value_get_object(*v);

	json_object_set_number(obj, "id", img->id);
	json_object_set_string(obj, "type", vm_backend_str(img->type));
	json_object_set_string(obj, "path", img->path);

	return ERRNOPE;
}

void vm_image_free(vm_image_t* img) {
	free((void*)img->path);
}

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

int vm_json(vm_t* vm, JSON_Value** v) {
	*v = json_value_init_object();
	JSON_Object* obj = json_value_get_object(*v);

	json_object_set_number(obj, "id", vm->id);
	json_object_set_string(obj, "state", vm_state_str(vm->state));
	json_object_set_string(obj, "backend", vm_backend_str(vm->params.backend));
	json_object_set_number(obj, "ncpu", vm->params.ncpu);
	json_object_set_number(obj, "memory", vm->params.memory);

	return ERRNOPE;
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
			return "hdd";
		case DEV_CDROM:
			return "cdrom";
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

vm_backend_t vm_backend_id(const char* name) {
	if(strcmp(name, "qemu") == 0) {
		return BACKEND_QEMU;
	}
	else if(strcmp(name, "openvz") == 0) {
		return BACKEND_VZ;
	}

	return BACKEND_UNKNOWN;
}

vm_dev_type_t vm_dev_id(const char* name) {
	if(strcmp(name, "hdd") == 0) {
		return DEV_HDD;
	}
	else if(strcmp(name, "cdrom") == 0) {
		return DEV_CDROM;
	}

	return DEV_UNKNOWN;
}

vm_state_t vm_state_id(const char* name) {
	if(strcmp(name, "up") == 0) {
		return STATE_UP;
	}
	else if(strcmp(name, "down") == 0) {
		return STATE_DOWN;
	}

	return STATE_UNKNOWN;
}
