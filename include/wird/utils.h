#pragma once

#include "wird/wird.h"

void         wird_log(const char* fmt, ...);
const char*  errstr(int err);
bool         isnum(const char* s);
