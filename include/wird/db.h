#pragma once

#include "wird/wird.h"
#include "wird/vm.h"
#include "lib/sqlite3.h"

extern sqlite3* global_db;

int db_connect(char* file);
int db_close();

int db_image_insert(vm_image_t* img, int* id);
int db_image_list(vm_image_t** images, int* count);
int db_image_get_by_column_int(vm_image_t* img, const char* colname, int value);

int db_dev_list_by_vmid(int vmid, vm_dev_t** devs, int* count);

int db_vm_insert(vm_params_t* p, int* id);
int db_vm_list(vm_t** vms, int* count);
int db_vm_get_by_column_int(vm_t* vm, const char* colname, int value);
int db_vm_set_state(vm_t* vm, vm_state_t state);
int db_vm_delete(int id);
int db_vm_param_set(vm_t* vm, const char* key, const char* value);
int db_vm_param_get(vm_t* vm, const char* key, char** value);
