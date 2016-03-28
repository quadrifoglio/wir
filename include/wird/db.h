#pragma once

#include "wird/wird.h"
#include "wird/vm.h"
#include "lib/sqlite3.h"

extern sqlite3* global_db;

int db_connect(char* file);
int db_vm_insert(vm_params_t* p, int* id);
int db_vm_list(vm_t** vms, int* count);
int db_vm_delete(int id);
int db_close();
