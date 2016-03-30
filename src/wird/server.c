#include "wird/server.h"

#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include "lib/parson.h"
#include "wird/vm.h"
#include "wird/db.h"
#include "wird/qemu.h"

char* json_message(bool success, const char* s) {
	if(success) {
		return strdup("{\"success\": true}");
	}
	else {
		char* ss;
		asprintf(&ss, "{\"success\": false, \"message\": \"%s\"}", s);

		return ss;
	}

	return 0;
}

server_result_t server_image_create(const char* data) {
	server_result_t res;
	return res;
}

server_result_t server_image_list(void) {
	server_result_t res;
	return res;
}

server_result_t server_vm_create(const char* data) {
	server_result_t r = {500, 0};

	JSON_Value* rootv = json_parse_string(data);
	if(!rootv) {
		r.status = 400;
		return r;
	}

	if(json_value_get_type(rootv) != JSONObject) {
		json_value_free(rootv);

		r.status = 400;
		return r;
	}

	JSON_Object* obj = json_value_get_object(rootv);
	const char* bs = json_object_get_string(obj, "backend");
	int ncpu = (int)json_object_get_number(obj, "ncpu");
	int mem = (int)json_object_get_number(obj, "memory");

	if(!bs || ncpu == 0 || mem == 0) {
		json_value_free(rootv);

		r.status = 400;
		return r;
	}

	vm_params_t p = {0};
	p.backend = vm_backend_id(bs);
	p.ncpu = ncpu;
	p.memory = mem;

	vm_t vm;
	int err = vm_create(&p, &vm);
	if(err != ERRNOPE) {
		json_value_free(rootv);

		r.status = 500;
		r.message = json_message(false, errstr(err));
		return r;
	}

	JSON_Value* res;
	vm_json(&vm, &res);

	r.status = 201;
	r.message = json_serialize_to_string(res);

	json_value_free(res);
	json_value_free(rootv);

	return r;
}

server_result_t server_vm_list(void) {
	server_result_t r = {500, 0};

	vm_t* vms = 0;
	int count = 0;
	int err = vm_list(&vms, &count);
	if(err != ERRNOPE) {
		r.status = 500;
		r.message = json_message(false, errstr(err));
		return r;
	}

	JSON_Value* rootv = json_value_init_object();
	JSON_Object* root = json_value_get_object(rootv);

	json_object_set_boolean(root, "success", true);

	JSON_Value* arrv = json_value_init_array();
	JSON_Array* arr = json_value_get_array(arrv);

	for(int i = 0; i < count; ++i) {
		vm_t vm = vms[i];
		JSON_Value* v;

		vm_json(&vm, &v);
		json_array_append_value(arr, v);
	}

	json_object_set_value(root, "vms", arrv);

	r.status = 200;
	r.message = json_serialize_to_string(rootv);

	json_value_free(rootv);
	free(vms);

	return r;
}

server_result_t server_vm_start(const char* id) {
	server_result_t r = {500, 0};
	vm_t vm = {0};

	int err = db_vm_get_by_column_int(&vm, "vmid", atoi(id));
	if(err == ERRNOTFOUND) {
		r.status = 404;
		r.message = json_message(false, "vm not found");
		return r;
	}
	else if(err != ERRNOPE) {
		r.status = 500;
		r.message = json_message(false, errstr(err));
		return r;
	}

	err = qemu_start(&vm);
	if(err != ERRNOPE) {
		r.status = 500;
		r.message = json_message(false, errstr(err));
		return r;
	}

	err = db_vm_set_state(&vm, vm.state);
	if(err != ERRNOPE) {
		r.status = 500;
		r.message = json_message(false, errstr(err));
		return r;
	}

	r.status = 200;
	return r;
}

server_result_t server_vm_stop(const char* id) {
	server_result_t r = {500, 0};
	vm_t vm = {0};

	int err = db_vm_get_by_column_int(&vm, "vmid", atoi(id));
	if(err == ERRNOTFOUND) {
		r.status = 404;
		r.message = json_message(false, "vm not found");
		return r;
	}
	else if(err != ERRNOPE) {
		r.status = 500;
		r.message = json_message(false, errstr(err));
		return r;
	}

	err = qemu_stop(&vm, false);
	if(err != ERRNOPE) {
		r.status = 500;
		r.message = json_message(false, errstr(err));
		return r;
	}

	err = db_vm_set_state(&vm, vm.state);
	if(err != ERRNOPE) {
		r.status = 500;
		r.message = json_message(false, errstr(err));
		return r;
	}

	r.status = 200;
	return r;
}

server_result_t server_vm_get(const char* id) {
	server_result_t r = {500, 0};
	vm_t vm = {0};

	int err = db_vm_get_by_column_int(&vm, "vmid", atoi(id));
	if(err == ERRNOTFOUND) {
		r.status = 404;
		r.message = json_message(false, "vm not found");
		return r;
	}
	else if(err != ERRNOPE) {
		r.status = 500;
		r.message = json_message(false, errstr(err));
		return r;
	}

	r.status = 200;

	JSON_Value* v;
	vm_json(&vm, &v);

	r.message = json_serialize_to_string(v);
	free(v);

	return r;
}

server_result_t server_vm_action(const char* method, const char* id, const char* action, const char* data) {
	server_result_t r;
	r.status = 404;

	if(strcmp(method, "GET") == 0) {
		if(action && id == 0 && strcmp(action, "list") == 0) {
			r = server_vm_list();
		}
		if(action && id != 0 && strcmp(action, "start") == 0) {
			r = server_vm_start(id);
		}
		if(action && id != 0 && strcmp(action, "stop") == 0) {
			r = server_vm_stop(id);
		}
		if(!action && id != 0) {
			r = server_vm_get(id);
		}
	}
	if(strcmp(method, "POST") == 0) {
		if(id == 0 && strcmp(action, "create") == 0 && data != 0) {
			r = server_vm_create(data);
		}
	}

	if(r.status >= 200 && r.status <= 299 && r.message == 0) {
		r.message = strdup("{\"success\": true}");
	}
	else if(r.status == 404 && r.message == 0) {
		r.message = strdup("{\"success\": false, \"message\": \"Not found\"}");
	}
	else if(r.status == 500 && r.message == 0) {
		r.message = strdup("{\"success\": false, \"message\": \"Server error\"}");
	}

	return r;
}
