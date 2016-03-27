#pragma once

#include "wird/wird.h"
#include "wird/vm.h"
#include "lib/sqlite3.h"

#define CHECK_DB_ERROR(err) if(err != SQLITE_OK) return ERRDB;

extern sqlite3* global_db;

int db_connect(char* file);
int db_insert_vm(vm_t* vm);
int db_close();
