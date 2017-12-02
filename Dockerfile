# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.9.2-stretch as builder

ENV LIBVIPS_VERSION 8.5.9
ENV DEP_VERSION v0.3.2

# Installs libvips + required libraries
RUN \
  # Install dependencies
  apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get install -y \
  automake build-essential curl \
  gobject-introspection gtk-doc-tools libglib2.0-dev libjpeg-dev libpng-dev \
  libwebp-dev libtiff5-dev libgif-dev libexif-dev libxml2-dev libpoppler-glib-dev \
  swig libmagickwand-dev libpango1.0-dev libmatio-dev libopenslide-dev libcfitsio-dev \
  libgsf-1-dev fftw3-dev liborc-0.4-dev librsvg2-dev && \
  # Build libvips
  cd /tmp && \
  curl -OL https://github.com/jcupitt/libvips/releases/download/v${LIBVIPS_VERSION}/vips-${LIBVIPS_VERSION}.tar.gz && \
  tar zvxf vips-${LIBVIPS_VERSION}.tar.gz && \
  cd /tmp/vips-${LIBVIPS_VERSION} && \
  ./configure --enable-debug=no --without-python $1 && \
  make && \
  make install && \
  ldconfig && \
  # Clean up
  apt-get remove -y curl automake build-essential && \
  apt-get autoremove -y && \
  apt-get autoclean && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Server port to listen
ENV PORT 8080

# gcc for cgo
RUN apt-get update && apt-get install -y \
    gcc curl git libc6-dev make ca-certificates \
    --no-install-recommends \
  && rm -rf /var/lib/apt/lists/*

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
WORKDIR $GOPATH
RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/$DEP_VERSION/dep-linux-amd64 && chmod +x /usr/local/bin/dep
ADD . /go/src/github.com/aldor007/mort

RUN cd /go/src/github.com/aldor007/mort &&  dep ensure -vendor-only
# RUN build
RUN cd /go; go build -o /go/mort src/github.com/aldor007cmd/mort/mort.go; cp -r /go/src/github.com/aldor007/mort/configuration /go/
# clean up
RUN rm -rf /go/src; rm -rf /go/pkg; rm -rf /usr/share/; rm -rf /usr/include/


#FROM golang:1.9.2-stretch

FROM debian:stretch

RUN \
  # Install dependencies
  apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get install -y \
  libglib2.0 libjpeg libpng \
  libwebpv libtiff5 libgifv libexifv libxml2 libpoppler-glib \
  libmagickwand libpango1.0 libmatio libopenslidev libcfitsio \
  libgsf-1 fftw3 liborc-0.4 librsvg2 && \
  # Clean up
  apt-get remove -y curl automake build-essential && \
  apt-get autoremove -y && \
  apt-get autoclean && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# clean up
RUN rm -rf /go/src; rm -rf /usr/share/; rm -rf /usr/include/

COPY --from=builder /usr/local/lib /usr/local/lib
RUN ldconfig
COPY --from=builder /go/mort /go/mort

# Run the outyet command by default when the container starts.
ENTRYPOINT ["/go/mort"]

# Expose the server TCP port
EXPOSE 8080 8081
