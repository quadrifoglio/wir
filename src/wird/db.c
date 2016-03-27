#include "wird/db.h"

sqlite3* global_db;

int db_connect(char* file) {
	int r = sqlite3_open(file, &global_db);
	if(r != 0) {
		return ERRDB;
	}

	return ERRNOPE;
}

int db_close() {
	return ERRNOPE;
}
