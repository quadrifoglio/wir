#pragma once

#include "wird/wird.h"

typedef enum {
	DEV_HDD   = 1,
	DEV_CDROM = 2
} device_type_t;

typedef struct {
	device_type_t type;
} device_t;

typedef struct {
	backend_t backend;

	int ncpu;
	int memory;

	device_t* devices;
	int device_count;
} vm_params_t;

typedef enum {
	STATE_DOWN = 0,
	STATE_UP   = 1
} vm_state_t;

typedef struct {
	int id;
	vm_state_t state;
	vm_params_t params;
} vm_t;

int vm_create(vm_params_t* p, vm_t* vm);
int vm_delete(vm_t* vm);
