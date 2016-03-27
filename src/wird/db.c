#include "wird/db.h"

sqlite3* global_db = 0;

int db_connect(char* file) {
	int err = sqlite3_open(file, &global_db);
	CHECK_DB_ERROR(err);

	const char* sql =
		"CREATE TABLE IF NOT EXISTS vm ("
			"vmid INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,"
			"vmstate INTEGER NOT NULL,"
			"vmbackend INTEGER NOT NULL,"
			"vmncpu INTEGER NOT NULL,"
			"vmmemory INTEGER NOT NULL"
		");"
		"CREATE TABLE IF NOT EXISTS dev ("
			"devid INTEGER PRIMARY KEY NOT NULL,"
			"devtype INTEGER NOT NULL,"
			"devvm INTEGER NOT NULL REFERENCES vm(vmid)"
		");";

	err = sqlite3_exec(global_db, sql, 0, 0, 0);
	CHECK_DB_ERROR(err);

	return ERRNOPE;
}

int db_insert_vm(vm_t* vm) {
	sqlite3_stmt* stmt;
	const char* sql = "INSERT INTO vm (vmstate, vmbackend, vmncpu, vmmemory) VALUES (?, ?, ?, ?)";

	int err = sqlite3_prepare_v2(global_db, sql, -1, &stmt, 0);
	CHECK_DB_ERROR(err);

	err = sqlite3_bind_int(stmt, 1, (int)vm->state);
	CHECK_DB_ERROR(err);

	err = sqlite3_bind_int(stmt, 2, (int)vm->params.backend);
	CHECK_DB_ERROR(err);

	err = sqlite3_bind_int(stmt, 3, vm->params.ncpu);
	CHECK_DB_ERROR(err);

	err = sqlite3_bind_int(stmt, 4, vm->params.memory);
	CHECK_DB_ERROR(err);

	err = sqlite3_step(stmt);
	if(err != SQLITE_DONE) {
		return ERRDB;
	}

	vm->id = (int)sqlite3_last_insert_rowid(global_db);

	sqlite3_finalize(stmt);

	sql = "INSERT INTO dev (devtype, devvm) VALUES (?, ?)";
	for(int i = 0; i < vm->params.device_count; ++i) {
		err = sqlite3_prepare_v2(global_db, sql, -1, &stmt, 0);
		CHECK_DB_ERROR(err);

		err = sqlite3_bind_int(stmt, 1, (int)vm->params.devices[i].type);
		CHECK_DB_ERROR(err);

		err = sqlite3_bind_int(stmt, 2, vm->id);
		CHECK_DB_ERROR(err);

		err = sqlite3_step(stmt);
		if(err != SQLITE_DONE) {
			return ERRDB;
		}

		sqlite3_finalize(stmt);
	}

	return ERRNOPE;
}

int db_close() {
	sqlite3_close(global_db);
	return ERRNOPE;
}
