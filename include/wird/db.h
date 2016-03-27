#pragma once

#include "wird/wird.h"
#include "lib/sqlite3.h"

extern sqlite3* global_db;

int db_connect(char* file);
int db_close();
