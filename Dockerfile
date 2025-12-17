FROM ghcr.io/aldor007/mort-base AS builder

ENV LIBVIPS_VERSION=8.11.2
ENV GOLANG_VERSION=1.25.4
ARG TARGETARCH=amd64
ARG TAG='dev'
ARG COMMIT="master"
ARG DATE="now"

ENV GOLANG_DOWNLOAD_URL=https://golang.org/dl/go$GOLANG_VERSION.linux-$TARGETARCH.tar.gz

RUN rm -rf /usr/local/go/ && curl -fsSL --insecure "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz \
  && tar -C /usr/local -xzf golang.tar.gz \
  && rm golang.tar.gz

ENV WORKDIR=/workspace
ENV PATH=/usr/local/go/bin:$PATH

WORKDIR $WORKDIR

# Download dependencies first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ $WORKDIR/cmd
COPY .godir ${WORKDIR}/.godir
COPY configuration/ ${WORKDIR}/configuration
COPY etc/ ${WORKDIR}/etc
COPY pkg/ ${WORKDIR}/pkg

# Build binary with optimizations
RUN CGO_ENABLED=1 GOOS=linux GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w -X 'main.version=${TAG}' -X 'main.commit=${COMMIT}' -X 'main.date=${DATE}'" \
    -o /go/mort ./cmd/mort/mort.go

# Download mime.types at build time for reproducibility
RUN curl -fsSL -o /tmp/mime.types https://raw.githubusercontent.com/apache/httpd/refs/heads/trunk/docs/conf/mime.types


# Runtime stage - use ubuntu 20.04 to match builder libraries
FROM ubuntu:20.04

# Install runtime dependencies in a single layer
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install --no-install-recommends -y \
    ca-certificates \
    libglib2.0-0 libjpeg-turbo8 libpng16-16 libopenexr24 \
    libwebp6 libwebpmux3 libwebpdemux2 libtiff5 libgif7 libexif12 libxml2 libpoppler-glib8 \
    libmagickwand-6.q16-6 libpango1.0-0 libmatio-dev libopenslide0 \
    libgsf-1-114 fftw3 liborc-0.4-0 librsvg2-2 libcfitsio8 libimagequant0 libheif1 libbrotli-dev && \
    apt-get autoremove -y && \
    apt-get autoclean && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Copy libvips libraries and update linker cache
COPY --from=builder /usr/local/lib /usr/local/lib
RUN ldconfig /usr/local/lib

# Create non-root user for security
RUN useradd -r -u 1000 -g users mort && \
    mkdir -p /etc/mort && \
    chown -R mort:users /etc/mort

# Copy application files
COPY --from=builder /go/mort /usr/local/bin/mort
COPY --from=builder /workspace/configuration/config.yml /etc/mort/mort.yml
COPY --from=builder /workspace/configuration/parse.tengo /etc/mort/parse.tengo
COPY --from=builder /tmp/mime.types /etc/mime.types

ENV MORT_CONFIG_DIR=/etc/mort

# Verify installation
RUN /usr/local/bin/mort -version

# Switch to non-root user
USER mort

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/usr/local/bin/mort", "-version"]

# Run the server
ENTRYPOINT ["/usr/local/bin/mort"]

# Expose the server TCP port
EXPOSE 8080 8081
