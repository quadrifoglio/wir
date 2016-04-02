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

void vm_params_drive_add(vm_params_t* p, vm_drive_type_t type, const char* file) {
	p->drives = realloc(p->drives, (++p->drive_count) * sizeof(vm_drive_t));
	p->drives[p->drive_count - 1] = (vm_drive_t){type, file};
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

int vm_load(int id, vm_t* vm) {
	int err = db_vm_get_by_column_int(vm, "vmid", id);
	if(err) {
		return err;
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

	if(vm->params.drive_count > 0) {
		JSON_Value* arrv = json_value_init_array();
		JSON_Array* arr = json_value_get_array(arrv);

		for(int i = 0; i < vm->params.drive_count; ++i) {
			JSON_Value* ov = json_value_init_object();
			JSON_Object* o = json_value_get_object(ov);

			json_object_set_string(o, "type", vm_drive_str(vm->params.drives[i].type));
			json_object_set_string(o, "file", vm->params.drives[i].file);

			json_array_append_value(arr, ov);
		}

		json_object_set_value(obj, "drives", arrv);
	}

	return ERRNOPE;
}

int vm_delete(vm_t* vm) {
	return db_vm_delete(vm->id);
}

void vm_params_free(vm_params_t* p) {
	for(int i = 0; i < p->drive_count; ++i) {
		free((void*)p->drives[i].file);
	}

	free(p->drives);
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

const char* vm_drive_str(vm_drive_type_t d) {
	switch(d) {
		case DRIVE_HDD:
			return "hdd";
		case DRIVE_CDROM:
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

vm_drive_type_t vm_drive_id(const char* name) {
	if(strcmp(name, "hdd") == 0) {
		return DRIVE_HDD;
	}
	else if(strcmp(name, "cdrom") == 0) {
		return DRIVE_CDROM;
	}

	return DRIVE_UNKNOWN;
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
