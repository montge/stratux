# Image for building Stratux
#
FROM debian:bookworm

# Install build dependencies and clean apt cache
RUN apt-get update \
    && apt-get -y install --no-install-recommends \
        ca-certificates \
        file \
        nano \
        make \
        git \
        gcc \
        ncurses-dev \
        wget \
        libusb-1.0-0-dev \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Go 1.24.1 (Debian Bookworm only has Go 1.20 which doesn't support toolchain directive)
# Note: golang.org/x/net v0.47.0 requires Go 1.24+
WORKDIR /tmp
RUN wget https://go.dev/dl/go1.24.1.linux-arm64.tar.gz \
    && tar -C /usr/local -xzf go1.24.1.linux-arm64.tar.gz \
    && rm go1.24.1.linux-arm64.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"

# Install RTL-SDR libraries
RUN wget https://github.com/stratux/rtlsdr/releases/download/v1.0/librtlsdr0_2.0.2-2_arm64.deb \
    && dpkg -i librtlsdr0_2.0.2-2_arm64.deb \
    && rm librtlsdr0_2.0.2-2_arm64.deb

RUN wget https://github.com/stratux/rtlsdr/releases/download/v1.0/librtlsdr-dev_2.0.2-2_arm64.deb \
    && dpkg -i librtlsdr-dev_2.0.2-2_arm64.deb \
    && rm librtlsdr-dev_2.0.2-2_arm64.deb

# specific to debian, ubuntu images come with user 'ubuntu' that is uid 1000
ENV USERNAME="stratux"
ENV USER_HOME=/home/$USERNAME

RUN useradd -m -d $USER_HOME -s /bin/bash $USERNAME \
    && echo "$USERNAME ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers
