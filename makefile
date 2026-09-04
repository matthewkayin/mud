# ------------------------------------------------------------------------------
# Project
# ------------------------------------------------------------------------------

ASSEMBLY        := mud
BUILD_DIR       := bin
OBJ_DIR         := obj
SRC_DIR         := src

CC               := clang
CC_FLAGS         := -std=c11
LD_FLAGS        :=

INCLUDE_FLAGS   := -Isrc -isystem vendor
DEFINES         := -D_CRT_SECURE_NO_WARNINGS

EXTENSION       :=

# ------------------------------------------------------------------------------
# Build configuration
# ------------------------------------------------------------------------------

CC_FLAGS += -g -O0
LD_FLAGS += -g

# ------------------------------------------------------------------------------
# Platform Detection
# ------------------------------------------------------------------------------

SRC_FILES := $(shell find $(SRC_DIR) -type f \( -name "*.c" \))
SRC_FILES += $(shell find vendor -type f \( -name "*.c" \))
DIRECTORIES := $(shell find src -type d)
DIRECTORIES += $(shell find vendor -type d)

# ------------------------------------------------------------------------------
# Source/Object Lists
# ------------------------------------------------------------------------------

OBJ_FILES := $(SRC_FILES:%=$(OBJ_DIR)/%.o)

# ------------------------------------------------------------------------------
# Target Recipes
# ------------------------------------------------------------------------------

.PHONY: all
all: scaffold compile link

# ------------------------------------------------------------------------------
# Scaffold
# ------------------------------------------------------------------------------

.PHONY: scaffold
scaffold:
	@echo Scaffolding...
	@mkdir -p $(BUILD_DIR)
	@mkdir -p $(addprefix $(OBJ_DIR)/,$(DIRECTORIES))
	@echo Done.

# ------------------------------------------------------------------------------
# Compile
# ------------------------------------------------------------------------------

# This phony just prints messages before compilation
.PHONY: compile
compile:
	@echo Compiler flags: $(CXX_FLAGS)
	@echo Defines:       $(DEFINES)
	@echo Compiling...

# Compile cpp to object
$(OBJ_DIR)/%.c.o: %.c
	@echo   $<...
	@$(CC) $< $(CC_FLAGS) -c -o $@ $(DEFINES) $(INCLUDE_FLAGS)

# ------------------------------------------------------------------------------
# Link
# ------------------------------------------------------------------------------

.PHONY: link
link: scaffold $(OBJ_FILES)
	@echo Linker flags $(LD_FLAGS)
	@echo Linking $(ASSEMBLY)...
	@$(CC) $(OBJ_FILES) -o $(BUILD_DIR)/$(ASSEMBLY)$(EXTENSION) $(LD_FLAGS)

# ------------------------------------------------------------------------------
# Clean
# ------------------------------------------------------------------------------

.PHONY: clean
clean:
	@echo Cleaning...
	@rm -rf $(OBJ_DIR)
	@rm -rf $(BUILD_DIR)
