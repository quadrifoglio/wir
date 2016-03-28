#pragma once

#include "wird/wird.h"
#include "lib/parson.h"

typedef enum {
	BACKEND_UNKNOWN = 0,
	BACKEND_QEMU    = 1,
	BACKEND_VZ      = 2
} vm_backend_t;

typedef enum {
	DEV_UNKNOWN = 0,
	DEV_HDD     = 1,
	DEV_CDROM   = 2
} vm_dev_type_t;

typedef struct {
	vm_dev_type_t type;
	const char* file;
} vm_dev_t;

typedef struct {
	vm_backend_t backend;

	int ncpu;
	int memory;

	vm_dev_t* devices;
	int device_count;
} vm_params_t;

typedef enum {
	STATE_UNKNOWN = 0,
	STATE_DOWN    = 1,
	STATE_UP      = 2
} vm_state_t;

typedef struct {
	int id;
	vm_state_t state;
	vm_params_t params;

	void* backend_data;
} vm_t;

int           vm_create(vm_params_t* p, vm_t* vm);
int           vm_json(vm_t* vm, JSON_Value** v);
int           vm_list(vm_t** vms, int* count);
int           vm_delete(vm_t* vm);

const char*   vm_backend_str(vm_backend_t b);
const char*   vm_dev_str(vm_dev_type_t d);
const char*   vm_state_str(vm_state_t s);

vm_backend_t  vm_backend_id(const char* name);
vm_dev_type_t vm_dev_id(const char* name);
vm_state_t    vm_state_id(const char* name);
