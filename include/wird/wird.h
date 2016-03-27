#pragma once

#include "global.h"

#define WIRD_VERSION "0.0.1"

#define ERRNOPE  0x00
#define ERRDB    0x01
#define ERRNOHYP 0x02

typedef enum {
	BACKEND_QEMU = 1,
	BACKEND_VZ = 2
} backend_t;

const char* errstr(int err);
