#include "wird/wird.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <errno.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>

#include "lib/http.h"
#include "wird/db.h"

const char* errstr(int err) {
	switch(err) {
		case ERRNOHYP:
			return "No hypervisor";
		case ERRDB:
			return sqlite3_errmsg(global_db);
		default:
			break;
	}

	return strerror(err);
}

void on_request(http_request_t* req, http_response_t* res) {

}

int server_bind(char* addrs, int port) {
	int sockfd = socket(PF_INET, SOCK_STREAM, IPPROTO_TCP);
	if(sockfd == -1) {
		perror("socket");
		return errno;
	}

	struct sockaddr_in sa = {0};
	sa.sin_family = AF_INET;
	sa.sin_port = htons(port);

	if(inet_aton(addrs, &sa.sin_addr) == 0) {
		fputs("aton: Invalid address", stderr);
		return errno;
	}

	setsockopt(sockfd, SOL_SOCKET, SO_REUSEADDR, &(int){1}, sizeof(int));

	if(bind(sockfd, (struct sockaddr *)&sa, sizeof(sa)) != 0) {
		perror("bind");
		return errno;
	}

	if(listen(sockfd, 1) != 0) {
		perror("listen");
		return errno;
	}

	while(true) {
		int csfd = accept(sockfd, 0, 0);
		if(csfd == -1) {
			perror("accept");
			continue;
		}

		// TODO: Concurrency
		http_client_loop(csfd, on_request, 0);
		shutdown(csfd, SHUT_RDWR);
		close(csfd);
	}

	shutdown(sockfd, 2);
	close(sockfd);

	return 0;
}

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
				++i;
			}
			else {
				printf("%s\n", usage);
				fprintf(stderr, "Please specify an address to bind to\n");
				return 1;
			}
		}
		else if(strcmp(arg, "-p") == 0 || strcmp(arg, "--port") == 0) {
			if(i + 1 < argc) {
				port = atoi(argv[i + 1]);
				++i;
			}
			else {
				printf("%s\n", usage);
				fprintf(stderr, "Please specify a port to bind to\n");
				return 1;
			}
		}
		else if(strcmp(arg, "-d") == 0 || strcmp(arg, "--database") == 0) {
			if(i + 1 < argc) {
				dbs = argv[i + 1];
				++i;
			}
			else {
				printf("%s\n", usage);
				fprintf(stderr, "Please specify a database file to use\n");
				return 1;
			}
		}
		else if(i != 0) {
			printf("%s\n", usage);
			fprintf(stderr, "Invalid argument: %s\n", arg);
			return 1;
		}
	}

	int err = db_connect(dbs);
	if(err != ERRNOPE) {
		fprintf(stderr, "Can not connect to the database %s: %s\n", dbs, errstr(err));
		return 1;
	}

	err = server_bind(addrs, port);
	if(err != ERRNOPE) {
		fprintf(stderr, "Can not start server: %s\n", errstr(err));
		return 1;
	}
}
