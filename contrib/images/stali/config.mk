# stali version
VERSION = 0.1

# paths
DESTDIR=$(PWD)/../rootfs-x86_64
PREFIX = /
MANPREFIX = $(PREFIX)/share/man

M4 = m4
CC = $(ROOT)/../toolchain/x86_64-linux-musl/bin/x86_64-linux-musl-gcc
LD = $(CC)

YACC = $(ROOT)/bin/hbase/yacc/yacc
AR = $(ROOT)/../toolchain/x86_64-linux-musl/bin/x86_64-linux-musl-ar
RANLIB = $(ROOT)/../toolchain/x86_64-linux-musl/bin/x86_64-linux-musl-ranlib

CPPFLAGS = -D_POSIX_SOURCE -D__stali__
CFLAGS   = -g -I$(ROOT)/../toolchain/x86_64-linux-musl/x86_64-linux-musl/include
#-std=c99 -Wall -pedantic
#LDFLAGS  = -s -static
LDFLAGS  = -static
