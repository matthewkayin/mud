#include "network.h"

#include "defines.h"
#include "logger.h"
#include <sys/socket.h>
#include <poll.h>
#include <unistd.h>
#include <arpa/inet.h>
#include <fcntl.h>
#include <string.h>
#include <stdio.h>
#include <errno.h>
#include <stdarg.h>

#define NETWORK_PORT 7272
#define NETWORK_BUFFER_SIZE 1024
#define NETWORK_OUT_BUFFER_SIZE 2048
#define NETWORK_MAX_PLAYERS 32
#define NETWORK_MAX_FD_COUNT (NETWORK_MAX_PLAYERS + 1)
#define NETWORK_FD_NONE -1

#define NETWORK_BYTE_IAC 0xFF

typedef struct NetworkPlayer {
    char in_buffer[NETWORK_BUFFER_SIZE];
    int in_buffer_length;
} NetworkPlayer;

struct NetworkState {
    int server_fd;
    struct sockaddr_in server_address;

    // Note: The players array is one larger than necessary
    // so that the player_index matches the fd_index
    struct pollfd fds[NETWORK_MAX_FD_COUNT];
    NetworkPlayer players[NETWORK_MAX_FD_COUNT];
    int fd_count;
};
static struct NetworkState state;

// Internal
void network_set_socket_nonblocking(int fd);
void network_close_and_remove_fd_at_index(int index);
bool network_get_new_pending_client(int* out_new_client_fd);
bool network_receive_message(int fd_index, char* buffer, int buffer_size, ssize_t* out_bytes_received);
void network_send(int recipient_fd, const char* message, ...);
void network_broadcast(const char* message, ...);

bool network_init() {
    // Zero-init player FDs
    for (int index = 1; index < NETWORK_MAX_FD_COUNT; index++) {
        state.fds[index].fd = NETWORK_FD_NONE;
    }

    // Create server TCP socket
    state.server_fd = socket(AF_INET, SOCK_STREAM, 0);
    if (state.server_fd < 0) {
        log_error("Server socket creation failed.");
        return false;
    }

    // Allow immediate port re-use after crash / restart
    const int opt = 1;
    setsockopt(state.server_fd, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt));

    // Bind socket to port
    memset(&state.server_address, 0, sizeof(state.server_address));
    state.server_address.sin_family = AF_INET;
    state.server_address.sin_addr.s_addr = INADDR_ANY;
    state.server_address.sin_port = htons(NETWORK_PORT);

    const int bind_result = bind(state.server_fd, (struct sockaddr*)&state.server_address, sizeof(state.server_address));
    if (bind_result < 0) {
        log_error("Failed to bind server socket.");
        return false;
    }

    // Listen for incoming connections
    const int listen_result = listen(state.server_fd, 10);
    if (listen_result < 0) {
        log_error("Failed to listen on server socket.");
        return false;
    }

    network_set_socket_nonblocking(state.server_fd);

    // Add server socket to the FD list
    state.fds[0].fd = state.server_fd;
    state.fds[0].events = POLLIN; // Only watch for incoming connections
    state.fd_count = 1;

    log_info("%s server listening on port %d...", APP_NAME, NETWORK_PORT);
    return true;
}


void network_quit() {
    close(state.server_fd);
}

void network_poll() {
    const int poll_result = poll(state.fds, state.fd_count, 0);
    if (poll_result < 0) {
        log_error("Network polling error %s.", strerror(errno));
    }

    // Save the number of FDs here so that if a new client connects,
    // we will not iterate over that client's FD during this poll
    const int fd_count = state.fd_count;

    for (int index = 0; index < fd_count; index++) {
        // If no events, skip this FD
        if (state.fds[index].revents == 0) {
            continue;
        }

        // Handle errors
        if (state.fds[index].revents & (POLLERR | POLLHUP | POLLNVAL)) {
            log_error("Network socket %d disconnected or encountered an error.", state.fds[index].fd);
            network_close_and_remove_fd_at_index(index);
            if (index < fd_count - 1) {
                index--;
            }
            continue;
        }

        // Handle incoming connection on server socket
        if (state.fds[index].fd == state.server_fd && state.fds[index].revents & POLLIN) {
            int new_client_fd;
            while (network_get_new_pending_client(&new_client_fd)) {
                if (state.fd_count == NETWORK_MAX_FD_COUNT) {
                    network_send(new_client_fd, "The server is full. Sorry, friend.\r\n");
                    close(new_client_fd);
                    continue;
                }

                // Otherwise, accept the connection
                network_set_socket_nonblocking(new_client_fd);
                state.fds[state.fd_count].fd = new_client_fd;
                state.fds[state.fd_count].events = POLLIN;
                state.fds[state.fd_count].revents = 0;
                state.players[state.fd_count].in_buffer_length = 0;
                state.fd_count++;
                log_info("New player connected on FD %d", new_client_fd);

                // Send a welcome message
                network_send(new_client_fd, "Welcome to %s!\r\n", APP_NAME);
            }
        }

        // Handle incoming messages from client socket
        if (state.fds[index].fd != state.server_fd && state.fds[index].revents & POLLIN) {
            char buffer[NETWORK_BUFFER_SIZE];
            ssize_t bytes_received;
            while (network_receive_message(index, buffer, sizeof(buffer), &bytes_received)) {
                // If bytes received is 0, then this client has disconnected
                if (bytes_received == 0) {
                    log_info("Player on FD %d disconnected.", state.fds[index].fd);
                    network_close_and_remove_fd_at_index(index);

                    if (index < fd_count - 1) {
                        index--;
                    }
                    break;
                }

                int byte_index = 0;
                while (byte_index < bytes_received) {
                    char next_byte = buffer[byte_index];
                    byte_index++;

                    // If next byte is IAC (Interpret As Command), then the next 2 bytes
                    // are a command and we should not treat them as regular text
                    if ((unsigned char)next_byte == NETWORK_BYTE_IAC) {
                        // TODO: handle IAC?
                        byte_index += 2;
                        continue;
                    }

                    // If next byte is end line, then submit the user's message
                    if (next_byte == '\n' || next_byte == '\r') {
                        if (state.players[index].in_buffer_length == 0) {
                            continue;
                        }

                        // Nul terminate and log the player's message
                        state.players[index].in_buffer[state.players[index].in_buffer_length] = '\0';
                        log_info("[Player %d]: %s", state.fds[index].fd, state.players[index].in_buffer);

                        // Broadcast the message to the other players
                        network_broadcast("[Player %d]: %s\r\n", state.fds[index].fd, state.players[index].in_buffer);

                        // Reset the playe's in_buffer for the next command
                        state.players[index].in_buffer_length = 0;
                        continue;
                    }

                    // If the next byte is an ASCII character, add it to the player's in_buffer
                    if (next_byte >= 32 && next_byte <= 126 && state.players[index].in_buffer_length < NETWORK_BUFFER_SIZE - 1) {
                        state.players[index].in_buffer[state.players[index].in_buffer_length] = next_byte;
                        state.players[index].in_buffer_length++;
                    }
                }
            }
        }
    }
}

// INTERNAL

void network_set_socket_nonblocking(int fd) {
    int flags = fcntl(fd, F_GETFL, 0);
    fcntl(fd, F_SETFL, flags | O_NONBLOCK);
}

void network_close_and_remove_fd_at_index(int index) {
    close(state.fds[index].fd);
    state.fds[index] = state.fds[state.fd_count - 1];
    memcpy(&state.players[index], &state.players[state.fd_count - 1], sizeof(state.players[index]));
    state.fd_count--;
}

bool network_get_new_pending_client(int* out_new_client_fd) {
    int new_client_fd = accept(state.server_fd, NULL, NULL);

    // Return false if all pending connections have been handled
    if (new_client_fd < 0 && (errno == EWOULDBLOCK || errno == EAGAIN)) {
        return false;
    }

    // Return false and log error if we received a different errno
    if (new_client_fd < 0) {
        log_error("Network error accepting new client: %s.", strerror(errno));
        return false;
    }

    if (out_new_client_fd != NULL) {
        *out_new_client_fd = new_client_fd;
    }
    return true;
}

bool network_receive_message(int fd_index, char* buffer, int buffer_size, ssize_t* out_bytes_received) {
    int client_fd = state.fds[fd_index].fd;
    ssize_t bytes_received = recv(client_fd, buffer, buffer_size - 1, 0);

    // Return false if all pending messages have been handled
    if (bytes_received < 0 && (errno == EWOULDBLOCK || errno == EAGAIN)) {
        return false;
    }

    // Return false and log error if we received a different errno
    if (bytes_received < 0) {
        log_error("Network error receiving client message: %s.", strerror(errno));
        return false;
    }

    if (out_bytes_received != NULL) {
        *out_bytes_received = bytes_received;
    }
    return true;
}

void network_send(int recipient_fd, const char* message, ...) {
    char out_message[NETWORK_OUT_BUFFER_SIZE];
    memset(out_message, 0, sizeof(out_message));

    __builtin_va_list arg_ptr;
    va_start(arg_ptr, message);
    vsnprintf(out_message, NETWORK_OUT_BUFFER_SIZE, message, arg_ptr);
    va_end(arg_ptr);

    // TODO: I think we need to store output data in a buffer and continue calling
    // send until all output has been sent
    send(recipient_fd, out_message, strlen(out_message), 0);
}

void network_broadcast(const char* message, ...) {
    char out_message[NETWORK_OUT_BUFFER_SIZE];
    memset(out_message, 0, sizeof(out_message));

    __builtin_va_list arg_ptr;
    va_start(arg_ptr, message);
    vsnprintf(out_message, NETWORK_OUT_BUFFER_SIZE, message, arg_ptr);
    va_end(arg_ptr);

    size_t out_message_len = strlen(out_message);
    for (int index = 1; index < state.fd_count; index++) {
        send(state.fds[index].fd, out_message, out_message_len, 0);
    }
}
