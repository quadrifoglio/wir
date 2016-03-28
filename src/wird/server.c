#include "wird/server.h"

#include <string.h>
#include <stdio.h>
#include "lib/parson.h"
#include "wird/vm.h"

server_result_t server_vm_create(const char* data) {
	server_result_t r = {500, strdup("{\"success\": false}")};

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
		r.message = strdup(errstr(err));
		return r;
	}

	JSON_Value* res;
	vm_json(&vm, &res);

	r.status = 201;
	r.message = json_serialize_to_string(res);

	return r;
}

server_result_t server_vm_list(void) {
	server_result_t r = {500, strdup("{\"success\": false}")};

	vm_t* vms = 0;
	int count = 0;
	int err = vm_list(&vms, &count);
	if(err != ERRNOPE) {
		r.status = 500;
		r.message = strdup(errstr(err));
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
	return r;
}

server_result_t server_vm_get(const char* id) {
	server_result_t r = {500, strdup("{\"success\": false}")};
	return r;
}

server_result_t server_vm_action(const char* method, const char* id, const char* action, const char* data) {
	server_result_t r;
	r.status = 404;
	r.message = strdup("{\"success\": false, \"message\": \"Action not found\"}");

	if(strcmp(method, "GET") == 0) {
		if(id == 0 && strcmp(action, "list") == 0) {
			r = server_vm_list();
		}
		if(id != 0 && action == 0) {
			r = server_vm_get(id);
		}
	}
	if(strcmp(method, "POST") == 0) {
		if(id == 0 && strcmp(action, "create") == 0 && data != 0) {
			r = server_vm_create(data);
		}
	}

	return r;
}
