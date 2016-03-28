#include "wird/db.h"

#include <stdlib.h>
#include <string.h>

#define CHECK_DB_ERROR(err) if(err != SQLITE_OK) { err = ERRDB; goto cleanup; }

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
	if(err != SQLITE_OK) {
		db_close();
	}

cleanup:
	return err;
}

int db_vm_insert(vm_params_t* p, int* id) {
	int err = ERRNOPE;

	sqlite3_stmt* stmt = 0;
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
	else {
		err = ERRNOPE;
	}

	*id = (int)sqlite3_last_insert_rowid(global_db);

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
		else {
			err = ERRNOPE;
		}

		sqlite3_finalize(stmt);
	}

cleanup:
	sqlite3_finalize(stmt);
	return err;
}

int db_vm_cb(void* d, int argc, char** argv, char** colname) {
	struct s {
		vm_t** vms;
		int* count;
	};

	if(argc != 5) {
		return 1;
	}

	struct s* ss = (struct s*)d;
	++(*ss->count);
	*ss->vms = realloc(*ss->vms, *ss->count * sizeof(vm_t));

	vm_t* vm = *ss->vms + (*ss->count - 1);
	memset(vm, 0, sizeof(vm_t));

	for(int i = 0; i < argc; ++i) {
		if(strcmp(colname[i], "vmid") == 0) {
			vm->id = atoi(argv[i]);
		}
		else if(strcmp(colname[i], "vmstate") == 0) {
			vm->state = (vm_state_t)atoi(argv[i]);
		}
		else if(strcmp(colname[i], "vmbackend") == 0) {
			vm->params.backend = (vm_backend_t)atoi(argv[i]);
		}
		else if(strcmp(colname[i], "vmncpu") == 0) {
			vm->params.ncpu = atoi(argv[i]);
		}
		else if(strcmp(colname[i], "vmmemory") == 0) {
			vm->params.memory = atoi(argv[i]);
		}
	}

	return 0;
}

int db_vm_list(vm_t** vms, int* count) {
	int err = ERRNOPE;

	struct s {
		vm_t** vms;
		int* count;
	};

	struct s* d = malloc(sizeof(struct s));
	d->vms = vms;
	d->count = count;

	const char* sql = "SELECT * FROM vm";
	err = sqlite3_exec(global_db, sql, db_vm_cb, (void*)d, 0);

	free(d);
	CHECK_DB_ERROR(err);

cleanup:
	return err;
}

int db_vm_delete(int id) {
	int err = ERRNOPE;
	sqlite3_stmt* stmt;
	const char* sql = "DELETE FROM vm WHERE vmid = ?";

	err = sqlite3_prepare_v2(global_db, sql, -1, &stmt, 0);
	CHECK_DB_ERROR(err);

	err = sqlite3_bind_int(stmt, 1, id);
	CHECK_DB_ERROR(err);

	err = sqlite3_step(stmt);
	if(err != SQLITE_DONE) {
		err = ERRDB;
		goto cleanup;
	}
	else {
		err = ERRNOPE;
	}

cleanup:
	sqlite3_finalize(stmt);
	return err;
}

int db_close() {
	sqlite3_close(global_db);
	return ERRNOPE;
}
