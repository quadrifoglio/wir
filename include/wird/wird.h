#pragma once

#include "global.h"

#define WIRD_VERSION "0.0.1"

#define ERRNOPE    0x00
#define ERRDB      0x01
#define ERRNOHYP   0x02
#define ERREXEC    0x03
#define ERRDOWN    0x04
#define ERRNOTFOUND 0x05

const char* errstr(int err);
