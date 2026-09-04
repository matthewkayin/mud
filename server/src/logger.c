#include "logger.h"

#include <stdarg.h>
#include <stdio.h>
#include <string.h>
#include <time.h>
#include <errno.h>
#include <sys/stat.h>

static FILE* logfile = NULL;

bool logger_init() {
    const char* log_folder_path = "./logs";
    // 0777 grants read, write, and execute permissions to all users
    int mkdir_result = mkdir(log_folder_path, 0777);
    if (mkdir_result != 0 && errno != EEXIST) {
        printf("Failed to create directory: %s\n", strerror(errno));
        return false;
    }

    // Create logfile path string
    char logfile_path[256];
    time_t _time = time(NULL);
    struct tm _tm = *localtime(&_time);
    sprintf(logfile_path,
        "%s/%d-%02d-%02dT%02d%02d%02d.log",
        log_folder_path,
        _tm.tm_year + 1900, _tm.tm_mon + 1, _tm.tm_mday, _tm.tm_hour,
        _tm.tm_min, _tm.tm_sec);

    // Open logfile
    logfile = fopen(logfile_path, "w");
    if (!logfile) {
        return false;
    }

    return true;
}

void logger_quit() {
    fclose(logfile);
}

void logger_output(LogLevel log_level, const char* message, ...) {
    const char* LOG_PREFIX[4] = {"ERROR", "WARN", "INFO", "DEBUG"};
    const size_t MESSAGE_BUFFER_LENGTH = 32000;
    char out_message[MESSAGE_BUFFER_LENGTH];

    memset(out_message, 0, sizeof(out_message));

    __builtin_va_list arg_ptr;
    va_start(arg_ptr, message);
    vsnprintf(out_message, MESSAGE_BUFFER_LENGTH, message, arg_ptr);
    va_end(arg_ptr);

    printf("[%s]: %s\n", LOG_PREFIX[log_level], out_message);

    if (logfile) {
        fprintf(logfile, "[%s]: %s\n", LOG_PREFIX[log_level], out_message);
    }
}
