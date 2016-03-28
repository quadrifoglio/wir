#pragma once

#include "wird/wird.h"
#include "wird/vm.h"

extern const char* executable;

int qemu_start(vm_t* vm);
int qemu_stop(vm_t* vm, bool violent);
