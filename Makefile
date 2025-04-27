### make os=xxx 为 go 编译器指定 GOOS
os_linux = linux
os_darwin = darwin
os_windows = windows
goos =
TARGET_OS = $(os)
ifeq ($(TARGET_OS), $(os_linux))
	goos = GOOS=linux
else ifeq ($(TARGET_OS), $(os_darwin))
	goos = GOOS=darwin
else ifeq ($(TARGET_OS), $(os_windows))
	goos = GOOS=windows
endif
######################################

compile:
	go mod tidy
	go build .