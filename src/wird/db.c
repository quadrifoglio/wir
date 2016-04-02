#include "wird/db.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "wird/utils.h"

#define CHECK_DB_ERROR(err) \
	if(err != SQLITE_OK) { \
		wird_log("database: %s\n", sqlite3_errstr(err)); \
		err = ERRDB; goto cleanup; \
	}

sqlite3* global_db = 0;

int db_connect(char* file) {
	int err = sqlite3_open(file, &global_db);
	CHECK_DB_ERROR(err);

	const char* sql =
		"CREATE TABLE IF NOT EXISTS image ("
			"imgid     INTEGER PRIMARY KEY NOT NULL,"
			"imgtype   INTEGER NOT NULL,"
			"imgpath   CHAR(128) NOT NULL"
		");"
		"CREATE TABLE IF NOT EXISTS vm ("
			"vmid      INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,"
			"vmstate   INTEGER NOT NULL,"
			"vmbackend INTEGER NOT NULL,"
			"vmncpu    INTEGER NOT NULL,"
			"vmmemory  INTEGER NOT NULL"
		");"
		"CREATE TABLE IF NOT EXISTS vm_param ("
			"pvm       INTEGER NOT NULL REFERENCES vm(vmid),"
			"pkey      CHAR(10) NOT NULL,"
			"pval      CHAR(10) NOT NULL,"
			"PRIMARY KEY(pvm, pkey)"
		");"
		"CREATE TABLE IF NOT EXISTS drive ("
			"driveid   INTEGER PRIMARY KEY NOT NULL,"
			"drivevm   INTEGER NOT NULL REFERENCES vm(vmid),"
			"drivetype INTEGER NOT NULL,"
			"drivefile CHAR(255) NOT NULL"
		");";

	err = sqlite3_exec(global_db, sql, 0, 0, 0);
	if(err != SQLITE_OK) {
		return err;
	}

cleanup:
	return err;
}

int db_image_insert(vm_image_t* img, int* id) {
	int err = ERRNOPE;

	sqlite3_stmt* stmt = 0;
	const char* sql = "INSERT INTO image (imgtype, imgpath) VALUES (?, ?)";

	err = sqlite3_prepare_v2(global_db, sql, -1, &stmt, 0);
	CHECK_DB_ERROR(err);

	err = sqlite3_bind_int(stmt, 1, (int)img->type);
	CHECK_DB_ERROR(err);

	err = sqlite3_bind_text(stmt, 2, img->path, strlen(img->path), SQLITE_STATIC);
	CHECK_DB_ERROR(err);

	err = sqlite3_step(stmt);
	if(err != SQLITE_DONE) {
		err = ERRDB;
		goto cleanup;
	}

	*id = (int)sqlite3_last_insert_rowid(global_db);
	if(*id) {
		err = ERRNOPE;
	}

cleanup:
	sqlite3_finalize(stmt);
	return err;
}

int db_image_cb(void* d, int argc, char** argv, char** colname) {
	struct s {
		vm_image_t** imgs;
		int* count;
	};

	if(argc != 3) {
		return 1;
	}

	struct s* ss = (struct s*)d;
	++(*ss->count);
	*ss->imgs = realloc(*ss->imgs, *ss->count * sizeof(vm_image_t));

	vm_image_t* img = *ss->imgs + (*ss->count - 1);
	memset(img, 0, sizeof(vm_image_t));

	for(int i = 0; i < argc; ++i) {
		if(strcmp(colname[i], "imgid") == 0) {
			img->id = atoi(argv[i]);
		}
		else if(strcmp(colname[i], "imgtype") == 0) {
			img->type = (vm_backend_t)atoi(argv[i]);
		}
		else if(strcmp(colname[i], "imgpath") == 0) {
			img->path = (const char*)strdup(argv[i]);
		}
	}

	return 0;
}

int db_image_list(vm_image_t** imgs, int* count) {
	int err = ERRNOPE;

	struct s {
		vm_image_t** imgs;
		int* count;
	};

	struct s* d = malloc(sizeof(struct s));
	d->imgs = imgs;
	d->count = count;

	const char* sql = "SELECT * FROM image";
	err = sqlite3_exec(global_db, sql, db_image_cb, (void*)d, 0);

	free(d);
	CHECK_DB_ERROR(err);

cleanup:
	return err;
}

int db_image_get_by_column_int(vm_image_t* img, const char* colname, int value) {
	int err = ERRNOPE;

	char* sql = 0;
	asprintf(&sql, "SELECT * FROM image WHERE %s = ? LIMIT 1", colname);

	sqlite3_stmt* stmt;

	err = sqlite3_prepare_v2(global_db, sql, -1, &stmt, 0);
	CHECK_DB_ERROR(err);

	err = sqlite3_bind_int(stmt, 1, value);
	CHECK_DB_ERROR(err);

	img->id = 0;

	while((err = sqlite3_step(stmt)) == SQLITE_ROW) {
		for(int col = 0; col < sqlite3_column_count(stmt); ++col) {
			const char* name = sqlite3_column_name(stmt, col);

			if(strcmp(name, "imgid") == 0) {
				img->id = sqlite3_column_int(stmt, col);
			}
			else if(strcmp(name, "imgtype") == 0) {
				img->type = (vm_backend_t)sqlite3_column_int(stmt, col);
			}
			else if(strcmp(name, "imgpath") == 0) {
				img->path = (const char*)sqlite3_column_text(stmt, col);
			}
		}
	}

	if(err != SQLITE_DONE) {
		err = ERRDB;
		goto cleanup;
	}
	else {
		err = ERRNOPE;
	}

	if(img->id == 0) {
		err = ERRNOTFOUND;
	}

cleanup:
	sqlite3_finalize(stmt);
	free(sql);
	return err;
}

int db_drive_cb(void* d, int argc, char** argv, char** colname) {
	struct s {
		vm_drive_t** drives;
		int* count;
	};

	if(argc != 5) {
		return 1;
	}

	struct s* ss = (struct s*)d;
	++(*ss->count);
	*ss->drives = realloc(*ss->drives, *ss->count * sizeof(vm_drive_t));

	vm_drive_t* drive = *ss->drives + (*ss->count - 1);
	memset(drive, 0, sizeof(vm_drive_t));

	for(int i = 0; i < argc; ++i) {
		if(strcmp(colname[i], "drivetype") == 0) {
			drive->type = (vm_drive_type_t)atoi(argv[i]);
		}
		else if(strcmp(colname[i], "drivefile") == 0) {
			puts(argv[i]);
			drive->file = (const char*)argv[i];
		}
	}

	return 0;
}

int db_drive_list_by_vmid(int vmid, vm_drive_t** drives, int* count) {
	int err = ERRNOPE;

	struct s {
		vm_drive_t** drives;
		int* count;
	};

	struct s* d = malloc(sizeof(struct s));
	d->drives = drives;
	d->count = count;

	char* sql;
	asprintf(&sql, "SELECT * FROM drive WHERE drivevm = %d", vmid);

	err = sqlite3_exec(global_db, sql, db_drive_cb, (void*)d, 0);

	free(d);
	CHECK_DB_ERROR(err);

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

	sql = "INSERT INTO drive (drivetype, drivevm, drivefile) VALUES (?, ?, ?)";
	for(int i = 0; i < p->drive_count; ++i) {
		sqlite3_stmt* stmt1;

		err = sqlite3_prepare_v2(global_db, sql, -1, &stmt1, 0);
		CHECK_DB_ERROR(err);

		err = sqlite3_bind_int(stmt1, 1, (int)p->drives[i].type);
		CHECK_DB_ERROR(err);

		err = sqlite3_bind_int(stmt1, 2, *id);
		CHECK_DB_ERROR(err);

		err = sqlite3_bind_text(stmt1, 3, p->drives[i].file, strlen(p->drives[i].file), SQLITE_STATIC);
		CHECK_DB_ERROR(err);

		err = sqlite3_step(stmt1);
		if(err != SQLITE_DONE) {
			err = ERRDB;
			goto cleanup;
		}
		else {
			err = ERRNOPE;
		}

		sqlite3_finalize(stmt1);
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

int db_vm_get_by_column_int(vm_t* vm, const char* colname, int value) {
	int err = ERRNOPE;

	char* sql = 0;
	asprintf(&sql, "SELECT * FROM vm WHERE %s = ? LIMIT 1", colname);

	sqlite3_stmt* stmt = 0;
	sqlite3_stmt* stmt1 = 0;

	err = sqlite3_prepare_v2(global_db, sql, -1, &stmt, 0);
	CHECK_DB_ERROR(err);

	err = sqlite3_bind_int(stmt, 1, value);
	CHECK_DB_ERROR(err);

	vm->id = 0;

	while((err = sqlite3_step(stmt)) == SQLITE_ROW) {
		for(int col = 0; col < sqlite3_column_count(stmt); ++col) {
			const char* name = sqlite3_column_name(stmt, col);

			if(strcmp(name, "vmid") == 0) {
				vm->id = sqlite3_column_int(stmt, col);
			}
			else if(strcmp(name, "vmstate") == 0) {
				vm->state = (vm_state_t)sqlite3_column_int(stmt, col);
			}
			else if(strcmp(name, "vmbackend") == 0) {
				vm->params.backend = (vm_backend_t)sqlite3_column_int(stmt, col);
			}
			else if(strcmp(name, "vmncpu") == 0) {
				vm->params.ncpu = sqlite3_column_int(stmt, col);
			}
			else if(strcmp(name, "vmmemory") == 0) {
				vm->params.memory = sqlite3_column_int(stmt, col);
			}
		}
	}

	if(err != SQLITE_DONE) {
		err = ERRDB;
		goto cleanup;
	}

	if(vm->id == 0) {
		err = ERRNOTFOUND;
		goto cleanup;
	}

	err = sqlite3_prepare_v2(global_db, "SELECT * FROM drive WHERE drivevm = ? LIMIT 1", -1, &stmt1, 0);
	CHECK_DB_ERROR(err);

	err = sqlite3_bind_int(stmt1, 1, vm->id);
	CHECK_DB_ERROR(err);

	while((err = sqlite3_step(stmt1)) == SQLITE_ROW) {
		vm_drive_type_t type = 0;
		char* file = 0;

		for(int col = 0; col < sqlite3_column_count(stmt1); ++col) {
			const char* name = sqlite3_column_name(stmt1, col);

			if(strcmp(name, "drivetype") == 0) {
				type = (vm_drive_type_t)sqlite3_column_int(stmt1, col);
			}
			else if(strcmp(name, "drivefile") == 0) {
				file = strdup((char*)sqlite3_column_text(stmt1, col));
			}
		}

		if(type && file) {
			vm_params_drive_add(&vm->params, type, file);
		}
	}

	if(err != SQLITE_DONE) {
		err = ERRDB;
		goto cleanup;
	}

	if(vm->id == 0) {
		err = ERRNOTFOUND;
	}
	else {
		err = ERRNOPE;
	}

cleanup:
	if(stmt)  sqlite3_finalize(stmt);
	if(stmt1) sqlite3_finalize(stmt1);
	free(sql);
	return err;
}

int db_vm_set_state(vm_t* vm, vm_state_t state) {
	int err = ERRNOPE;

	const char* sql = "UPDATE vm SET vmstate = ? WHERE vmid = ?";
	sqlite3_stmt* stmt;

	err = sqlite3_prepare_v2(global_db, sql, -1, &stmt, 0);
	CHECK_DB_ERROR(err);

	err = sqlite3_bind_int(stmt, 1, (int)state);
	CHECK_DB_ERROR(err);

	err = sqlite3_bind_int(stmt, 2, vm->id);
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

int db_vm_param_set(vm_t* vm, const char* key, const char* value) {
	int err = ERRNOPE;

	const char* sql = "INSERT OR REPLACE INTO vm_param VALUES (?, ?, ?)";
	sqlite3_stmt* stmt;

	err = sqlite3_prepare_v2(global_db, sql, -1, &stmt, 0);
	CHECK_DB_ERROR(err);

	err = sqlite3_bind_int(stmt, 1, vm->id);
	CHECK_DB_ERROR(err);

	err = sqlite3_bind_text(stmt, 2, key, strlen(key), SQLITE_STATIC);
	CHECK_DB_ERROR(err);

	err = sqlite3_bind_text(stmt, 3, value, strlen(value), SQLITE_STATIC);
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
	return err;
}

int db_vm_param_get(vm_t* vm, const char* key, char** value) {
	int err = ERRNOPE;

	const char* sql = "SELECT pval FROM vm_param WHERE pvm = ? AND pkey = ? LIMIT 1";
	sqlite3_stmt* stmt;

	err = sqlite3_prepare_v2(global_db, sql, -1, &stmt, 0);
	CHECK_DB_ERROR(err);

	err = sqlite3_bind_int(stmt, 1, vm->id);
	CHECK_DB_ERROR(err);

	err = sqlite3_bind_text(stmt, 2, key, strlen(key), SQLITE_STATIC);
	CHECK_DB_ERROR(err);

	err = sqlite3_step(stmt);
	if(err == SQLITE_ROW) {
		const unsigned char* v = sqlite3_column_text(stmt, 0);
		if(v) {
			*value = strdup((char*)v);
			err = ERRNOPE;

			goto cleanup;
		}
	}

	err = ERRNOTFOUND;

cleanup:
	sqlite3_finalize(stmt);
	return err;
}

int db_close() {
	sqlite3_close(global_db);
	return ERRNOPE;
}
