# Start from a Debian image with the latest version of Go installed
# and a workspace (WORKDIR) configured at /go.
FROM golang:1.12.4-stretch

ARG GITHUB_TOKEN=""

ENV LIBVIPS_VERSION 8.6.2
ENV DEP_VERSION v0.5.1
# Installs libvips + required libraries
RUN printf "deb http://httpredir.debian.org/debian stretch-backports main non-free\ndeb-src http://httpredir.debian.org/debian stretch-backports main non-free" > /etc/apt/sources.list.d/backports.list

RUN \
  # Install dependencies
  apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get install -y \
  automake build-essential curl \
  gobject-introspection gtk-doc-tools libglib2.0-dev libjpeg-dev libpng-dev \
  libwebp-dev libtiff5-dev libgif-dev libexif-dev libxml2-dev libpoppler-glib-dev \
  swig libmagickwand-dev libpango1.0-dev libmatio-dev libopenslide-dev libcfitsio-dev \
  libgsf-1-dev fftw3-dev liborc-0.4-dev librsvg2-dev libbrotli-dev && \
  # Build libvips
  cd /tmp && \
  curl -OL https://github.com/libvips/libvips/releases/download/v${LIBVIPS_VERSION}/vips-${LIBVIPS_VERSION}.tar.gz && \
  tar zvxf vips-${LIBVIPS_VERSION}.tar.gz && \
  cd /tmp/vips-${LIBVIPS_VERSION} && \
  ./configure --enable-debug=no --without-python $1 && \
  make && \
  make install && \
  ldconfig && \
  # Clean up
  apt-get remove -y curl automake && \
  apt-get autoremove -y && \
  apt-get autoclean && \
  apt-get clean && \
  apt-get install ruby ruby-dev rubygems  -y && \
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# gcc for cgo
RUN apt-get update && apt-get install -y \
    gcc curl git libc6-dev make ca-certificates \
    --no-install-recommends \
  && rm -rf /var/lib/apt/lists/*

RUN gem install --no-ri --no-rdoc fpm

ENV WORKDIR /mort
ENV PATH $WORKDIR/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$WORKDIR/src" "$WORKDIR/bin" && chmod -R 777 "$WORKDIR"
WORKDIR $WORKDIR
ADD . $WORKDIR

RUN cd $WORKDIR; go mod vendor
# RUN build
RUN cd $WORKDIR; GITHUB_TOKEN=${GITHUB_TOKEN} curl -sL http://git.io/goreleaser | bash

