#include "wird/server.h"

#include <string.h>
#include <stdio.h>
#include "lib/parson.h"
#include "wird/vm.h"

server_result_t server_vm_create(void) {
	server_result_t r = {500, strdup("Server error")};
	return r;
}

server_result_t server_vm_list(void) {
	server_result_t r = {500, strdup("Server error")};

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

		JSON_Value* objv = json_value_init_object();
		JSON_Object* obj = json_value_get_object(objv);

		json_object_set_number(obj, "id", vm.id);
		json_object_set_string(obj, "state", vm_state_str(vm.state));
		json_object_set_string(obj, "backend", vm_backend_str(vm.params.backend));
		json_object_set_number(obj, "ncpu", vm.params.ncpu);
		json_object_set_number(obj, "memory", vm.params.memory);

		json_array_append_value(arr, objv);
	}

	json_object_set_value(root, "vms", arrv);

	r.status = 200;
	r.message = json_serialize_to_string(rootv);

	json_value_free(rootv);
	return r;
}

server_result_t server_vm_get(const char* id) {
	server_result_t r = {500, strdup("Server error")};
	return r;
}

server_result_t server_vm_action(const char* id, const char* action) {
	server_result_t r;
	r.status = 404;
	r.message = strdup("{\"success\": false, \"message\": \"Action not found\"}");

	if(id == 0 && strcmp(action, "create") == 0) {
		r = server_vm_create();
	}
	else if(id == 0 && strcmp(action, "list") == 0) {
		r = server_vm_list();
	}
	else if(id != 0 && action == 0) {
		r = server_vm_get(id);
	}

	return r;
}
