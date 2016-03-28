#pragma once

#include "wird/wird.h"
#include "wird/vm.h"
#include "lib/sqlite3.h"

extern sqlite3* global_db;

int db_connect(char* file);
int db_insert_vm(vm_params_t* p, int* id);
int db_close();
