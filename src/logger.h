#pragma once

#include <stdbool.h>

typedef enum LogLevel {
    LOG_LEVEL_ERROR,
    LOG_LEVEL_WARN,
    LOG_LEVEL_INFO,
    LOG_LEVEL_DEBUG
} LogLevel;

bool logger_init();
void logger_quit();

void logger_output(LogLevel log_level, const char* message, ...);

#define log_error(message, ...) logger_output(LOG_LEVEL_ERROR, message, ##__VA_ARGS__);
#define log_warn(message, ...) logger_output(LOG_LEVEL_WARN, message, ##__VA_ARGS__);
#define log_info(message, ...) logger_output(LOG_LEVEL_INFO, message, ##__VA_ARGS__);
#define log_debug(message, ...) logger_output(LOG_LEVEL_DEBUG, message, ##__VA_ARGS__);
