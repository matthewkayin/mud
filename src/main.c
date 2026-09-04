#include "defines.h"
#include "logger.h"
#include "network.h"
#include <signal.h>
#include <stdatomic.h>

// This is atomic since the signals might change this while
// another signal is trying to change it or while the main
// loop is reading it
static atomic_bool server_is_running = true;

bool mud_init();
void mud_socket_set_nonblocking(int fd);
void mud_stop_running();
void mud_quit();

int main() {
    // This makes it so that we quit gracefully on Ctrl+C
    signal(SIGINT, mud_stop_running);
    signal(SIGHUP, mud_stop_running);
    signal(SIGTERM, mud_stop_running);

    if (!mud_init()) {
        return 1;
    }

    while (server_is_running) {
        network_poll();
    }

    mud_quit();
    return 0;
}


bool mud_init() {
    if (!logger_init()) {
        return false;
    }

    if (!network_init()) {
        return false;
    }

    return true;
}

void mud_stop_running() {
    server_is_running = false;
}

void mud_quit() {
    log_info("Quitting %s...", APP_NAME);
    network_quit();

    log_info("%s quit gracefully.", APP_NAME);
    logger_quit();
}
