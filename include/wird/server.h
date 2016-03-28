#pragma once

#include "wird/wird.h"

typedef struct {
	int status;
	char* message;
} server_result_t;

server_result_t server_vm_action(const char* id, const char* action);
