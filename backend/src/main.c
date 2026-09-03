#include <stdio.h>
#include <unistd.h>
#include <wsServer/ws.h>

void onopen(ws_cli_conn_t client) {
    char* cli = ws_getaddress(client);
    printf("connection opened, addr: %s\n", cli);
}

void onclose(ws_cli_conn_t client) {
    char* cli = ws_getaddress(client);
    printf("Connection closed, addr: %s\n", cli);
}

void onmessage(ws_cli_conn_t client, const unsigned char* message, uint64_t size, int type) {
    char* cli = ws_getaddress(client);
    printf("I receive a message: %s (%zu), from %s", message, size, cli);

    sleep(2);
    ws_sendframe_txt(client, "hey");
    sleep(2);
    ws_sendframe_txt(client, "friend");
}

int main() {
    ws_socket(&(struct ws_server) {
        .host = "localhost",
        .port = 8080,
        .thread_loop = 0,
        .timeout_ms = 1000,
        .evs = {
            .onopen = &onopen,
            .onclose = &onclose,
            .onmessage = &onmessage
        }
    });

    return 0;
}
