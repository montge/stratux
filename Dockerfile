# Image for building Stratux
#
FROM debian:bookworm

# file and nano are nice to have
RUN apt-get update \
  && apt-get -y install file \
  && apt-get -y install nano \
  && apt-get -y install make \
  && apt-get -y install git \
  && apt-get -y install gcc \
  && apt-get -y install ncurses-dev \
  && apt-get -y install wget \
  && apt-get -y install libusb-1.0-0-dev

# Install Go 1.23.12 (Debian Bookworm only has Go 1.20 which doesn't support toolchain directive)
RUN cd /tmp \
    && wget https://go.dev/dl/go1.23.12.linux-arm64.tar.gz \
    && tar -C /usr/local -xzf go1.23.12.linux-arm64.tar.gz \
    && rm go1.23.12.linux-arm64.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"

RUN cd /tmp \
    && wget https://github.com/stratux/rtlsdr/releases/download/v1.0/librtlsdr0_2.0.2-2_arm64.deb \
    && dpkg -i librtlsdr0_2.0.2-2_arm64.deb

RUN cd /tmp \
    && wget https://github.com/stratux/rtlsdr/releases/download/v1.0/librtlsdr-dev_2.0.2-2_arm64.deb \
    && dpkg -i librtlsdr-dev_2.0.2-2_arm64.deb

# specific to debian, ubuntu images come with user 'ubuntu' that is uid 1000
ENV USERNAME="stratux"
ENV USER_HOME=/home/$USERNAME

RUN useradd -m -d $USER_HOME -s /bin/bash $USERNAME \
    && echo "$USERNAME ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers
