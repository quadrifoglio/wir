#include "wird/db.h"

#define CHECK_DB_ERROR(err) if(err != SQLITE_OK) err = ERRDB; goto cleanup;

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

cleanup:
	db_close();
	return err;
}

int db_insert_vm(vm_params_t* p, int* id) {
	int err = ERRNOPE;

	sqlite3_stmt* stmt;
	const char* sql = "INSERT INTO vm (vmstate, vmbackend, vmncpu, vmmemory) VALUES (?, ?, ?, ?)";

	err = sqlite3_prepare_v2(global_db, sql, -1, &stmt, 0);
	CHECK_DB_ERROR(err);

	err = sqlite3_bind_int(stmt, 1, (int)STATE_DOWN);
	CHECK_DB_ERROR(err);

	err = sqlite3_bind_int(stmt, 2, (int)p->backend);
	CHECK_DB_ERROR(err);

	err = sqlite3_bind_int(stmt, 3, p->ncpu);
	CHECK_DB_ERROR(err);

	err = sqlite3_bind_int(stmt, 4, p->memory);
	CHECK_DB_ERROR(err);

	err = sqlite3_step(stmt);
	if(err != SQLITE_DONE) {
		err = ERRDB;
		goto cleanup;
	}

	*id = (int)sqlite3_last_insert_rowid(global_db);

	sqlite3_finalize(stmt);

	sql = "INSERT INTO dev (devtype, devvm) VALUES (?, ?)";
	for(int i = 0; i < p->device_count; ++i) {
		err = sqlite3_prepare_v2(global_db, sql, -1, &stmt, 0);
		CHECK_DB_ERROR(err);

		err = sqlite3_bind_int(stmt, 1, (int)p->devices[i].type);
		CHECK_DB_ERROR(err);

		err = sqlite3_bind_int(stmt, 2, *id);
		CHECK_DB_ERROR(err);

		err = sqlite3_step(stmt);
		if(err != SQLITE_DONE) {
			err = ERRDB;
			goto cleanup;
		}

		sqlite3_finalize(stmt);
	}

cleanup:
	sqlite3_finalize(stmt);
	return err;
}

int db_close() {
	sqlite3_close(global_db);
	return ERRNOPE;
}
