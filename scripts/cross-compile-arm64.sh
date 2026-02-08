#!/bin/bash
# Cross-compile Stratux components for ARM64

set -e

STRATUX_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="$STRATUX_DIR/build_arm64"
SYSROOT="/tmp/ncurses_arm64/usr"

export CC=aarch64-linux-gnu-gcc
export CXX=aarch64-linux-gnu-g++
export ARCH=aarch64
export GOOS=linux
export GOARCH=arm64
export CGO_ENABLED=1

mkdir -p "$BUILD_DIR"

echo "=== Building dump978 ==="
cd "$STRATUX_DIR/dump978"
make clean 2>/dev/null || true
make lib
cp "$STRATUX_DIR/libdump978.so" "$BUILD_DIR/"
file "$BUILD_DIR/libdump978.so"

echo "=== Building dump1090 ==="
cd "$STRATUX_DIR/dump1090"
make clean 2>/dev/null || true
# Temporarily modify Makefile to add tinfo to ncurses link
cp Makefile Makefile.bak
sed -i 's/LIBS_CURSES := -lncurses/LIBS_CURSES := -lncurses -ltinfo/' Makefile
RTLSDR=no BLADERF=no HACKRF=no LIMESDR=no SOAPYSDR=no \
LDFLAGS="-L$SYSROOT/lib/aarch64-linux-gnu -Wl,-rpath-link,$SYSROOT/lib/aarch64-linux-gnu" \
make dump1090
mv Makefile.bak Makefile
cp dump1090 "$BUILD_DIR/"
file "$BUILD_DIR/dump1090"

echo "=== Building rtl-ais ==="
cd "$STRATUX_DIR/rtl-ais"
make clean 2>/dev/null || true
# Cross-compile by patching Makefile to skip pkg-config
cp Makefile Makefile.bak
# Replace the ifeq block to use our paths
cat > Makefile.cross << 'EOFMAKE'
CFLAGS?=-O2 -g -Wall -W
CFLAGS+= -I./aisdecoder -I ./aisdecoder/lib -I./tcp_listener
LDFLAGS+=-lpthread -lm

# Cross-compile paths
CFLAGS += -I/tmp/ncurses_arm64/usr/include -I/tmp/ncurses_arm64/usr/include/libusb-1.0
LDFLAGS += -L/tmp/ncurses_arm64/usr/lib/aarch64-linux-gnu -Wl,-rpath-link,/tmp/ncurses_arm64/usr/lib/aarch64-linux-gnu -lrtlsdr -lusb-1.0

CC?=gcc
SOURCES= \
	main.c rtl_ais.c convenience.c \
	./aisdecoder/aisdecoder.c \
	./aisdecoder/sounddecoder.c \
	./aisdecoder/lib/receiver.c \
	./aisdecoder/lib/protodec.c \
	./aisdecoder/lib/hmalloc.c \
	./aisdecoder/lib/filter.c \
	./tcp_listener/tcp_listener.c

OBJECTS=$(SOURCES:.c=.o)
EXECUTABLE=rtl_ais

all: $(SOURCES) $(EXECUTABLE)

$(EXECUTABLE): $(OBJECTS)
	$(CC) $(OBJECTS) -o $@ $(LDFLAGS)

.c.o:
	$(CC) -c $< -o $@ $(CFLAGS)

clean:
	rm -f $(OBJECTS) $(EXECUTABLE) $(EXECUTABLE).exe
EOFMAKE
make -f Makefile.cross
rm Makefile.cross
mv Makefile.bak Makefile
cp rtl_ais "$BUILD_DIR/"
file "$BUILD_DIR/rtl_ais"

echo "=== Building stratuxrun ==="
cd "$STRATUX_DIR"

# Copy additional libraries needed for linking
cp "$SYSROOT/lib/aarch64-linux-gnu/libusb-1.0.so"* "$BUILD_DIR/" 2>/dev/null || true
cd "$BUILD_DIR" && ln -sf libusb-1.0.so.0.4.0 libusb-1.0.so 2>/dev/null || true
cd "$STRATUX_DIR"

CGO_ENABLED=1 \
CC=aarch64-linux-gnu-gcc \
GOOS=linux \
GOARCH=arm64 \
CGO_CFLAGS="-I$STRATUX_DIR -I$SYSROOT/include" \
CGO_LDFLAGS="-L$BUILD_DIR -L$SYSROOT/lib/aarch64-linux-gnu -Wl,-rpath-link,$SYSROOT/lib/aarch64-linux-gnu" \
go build -mod=mod \
    -ldflags "-X main.stratuxVersion=test-cross -X main.stratuxBuild=$(git log -n 1 --pretty=%H)" \
    -o "$BUILD_DIR/stratuxrun" ./main/
file "$BUILD_DIR/stratuxrun"

echo "=== Build complete ==="
ls -lh "$BUILD_DIR/"
