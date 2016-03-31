#include "wird/utils.h"

#include <stdio.h>
#include <string.h>
#include <ctype.h>
#include <time.h>
#include <stdarg.h>
#include "wird/db.h"

const char* errstr(int err) {
	switch(err) {
		case ERRDB:
			return sqlite3_errmsg(global_db);
		case ERRNOHYP:
			return "no hypervisor";
		case ERREXEC:
			return "exec system call failed";
		case ERRKILL:
			return "kill system call failed";
		case ERRDOWN:
			return "vm is down";
		case ERRNOTFOUND:
			return "not found";
		case ERRINVALID:
			return "invalid parameters";
		default:
			break;
	}

	return strerror(err);
}

bool isnum(const char* s) {
	char* ss = (char*)s;
	while(*ss) {
		if(!isdigit(*ss)) {
			return false;
		}
		else {
			++ss;
		}
	}

	return true;
}

void wird_log(const char* fmt, ...) {
	time_t timer;
	char buffer[26] = {0};
	struct tm* tm_info;

	time(&timer);
	tm_info = localtime(&timer);

	strftime(buffer, 26, "%Y:%m:%d %H:%M:%S", tm_info);
	printf("%s ", buffer);

	va_list v;
	va_start(v, fmt);
	vprintf(fmt, v);
	va_end(v);
}
