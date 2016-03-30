#pragma once

#include "global.h"

#define WIRD_VERSION "0.0.1"
#define WIRD_LOG     true

#define ERRNOPE     0x00
#define ERRDB       0x01
#define ERRNOHYP    0x02
#define ERREXEC     0x03
#define ERRKILL     0x04
#define ERRDOWN     0x05
#define ERRNOTFOUND 0x06
#define ERRINVALID  0x07

const char* errstr(int err);
void        wird_log(const char* fmt, ...);
