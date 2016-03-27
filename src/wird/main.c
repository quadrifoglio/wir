#include "wird/wird.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(int argc, char** argv) {
	const char* usage =
		"Usage: wird [OPTIONS]\n"
		"The HTTP virtualization node control server\n\n"
		"Options:\n\n"
		"-h, --help       Print this help and quit\n"
		"-v, --version    Print version information and quit\n"
		"-a, --address    Bind address (default 127.0.0.1)\n"
		"-p, --port       Bind port (default 1997)\n"
		"-d, --database   Database file (default wird.db)\n";

	char* dbs = "wird.db";
	char* addrs = "127.0.0.1";
	int port = 1997;

	for(int i = 0; i < argc; ++i) {
		char* arg = argv[i];

		if(strcmp(arg, "-h") == 0 || strcmp(arg, "--help") == 0) {
			printf("%s", usage);
			return 0;
		}
		else if(strcmp(arg, "-v") == 0 || strcmp(arg, "--version") == 0) {
			printf("wird version %s\n", WIRD_VERSION);
			return 0;
		}
		else if(strcmp(arg, "-a") == 0 || strcmp(arg, "--address") == 0) {
			if(i + 1 < argc) {
				addrs = argv[i + 1];
			}
			else {
				fprintf(stderr, "Please specify an address to bind to\n");
				printf("%s", usage);
				return 1;
			}
		}
		else if(strcmp(arg, "-p") == 0 || strcmp(arg, "--port") == 0) {
			if(i + 1 < argc) {
				port = atoi(argv[i + 1]);
			}
			else {
				fprintf(stderr, "Please specify a port to bind to\n");
				printf("%s", usage);
				return 1;
			}
		}
		else if(strcmp(arg, "-d") == 0 || strcmp(arg, "--database") == 0) {
			if(i + 1 < argc) {
				dbs = argv[i + 1];
			}
			else {
				fprintf(stderr, "Please specify a database file to use\n");
				printf("%s", usage);
				return 1;
			}
		}
	}

	printf("listening on %s:%d, db on %s\n", addrs, port, dbs);
	return 0;
}
