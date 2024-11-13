# syntax=docker/dockerfile:1

FROM golang:1.22

RUN mkdir /root/handbrake
WORKDIR /root/handbrake
ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y autoconf automake build-essential cmake git libass-dev libbz2-dev libfontconfig-dev libfreetype-dev libfribidi-dev libharfbuzz-dev libjansson-dev liblzma-dev libmp3lame-dev libnuma-dev libogg-dev libopus-dev libsamplerate0-dev libspeex-dev libtheora-dev libtool libtool-bin libturbojpeg0-dev libvorbis-dev libx264-dev libxml2-dev libvpx-dev m4 make meson nasm ninja-build patch pkg-config tar zlib1g-dev
RUN git clone https://github.com/HandBrake/HandBrake.git
WORKDIR /root/handbrake/HandBrake
RUN git checkout 1.7.x
RUN ./configure --launch-jobs=$(nproc) --launch --disable-gtk && make --directory=build install

