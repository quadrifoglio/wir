#pragma once

#include "wird/wird.h"

typedef enum {
	DEV_HDD,
	DEV_CDROM
} dev_type_t;

typedef struct {
	dev_type_t type;
} dev_t;

typedef struct {
	backend_t backend;

	int ncpu;
	int memory;

	dev_t* devices;
	int device_count;
} vm_params_t;

typedef struct {
	vm_params_t params;
} vm_t;

int vm_create(vm_params_t* p, vm_t* vm);
int vm_delete(vm_t* vm);
