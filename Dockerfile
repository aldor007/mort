FROM --platform=$TARGETPLATFORM ghcr.io/aldor007/mort-base:master-6feb103 AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETARCH
ARG TAG='dev'
ARG COMMIT="master"
ARG DATE="now"

ENV WORKDIR=/workspace
ENV PATH=/usr/local/go/bin:$PATH

WORKDIR $WORKDIR

# Download dependencies first (better caching with cache mount)
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY cmd/ $WORKDIR/cmd
COPY .godir ${WORKDIR}/.godir
COPY configuration/ ${WORKDIR}/configuration
COPY etc/ ${WORKDIR}/etc
COPY pkg/ ${WORKDIR}/pkg

# Build binary with optimizations and build cache
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=1 GOOS=linux GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w -X 'main.version=${TAG}' -X 'main.commit=${COMMIT}' -X 'main.date=${DATE}'" \
    -trimpath \
    -o /go/mort ./cmd/mort/mort.go

# Download mime.types at build time for reproducibility
RUN curl -fsSL -o /tmp/mime.types https://raw.githubusercontent.com/apache/httpd/refs/heads/trunk/docs/conf/mime.types


# Runtime stage - use minimal ubuntu 20.04
FROM --platform=$TARGETPLATFORM ubuntu:20.04

# Install runtime dependencies, create user, and cleanup in single layer
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install --no-install-recommends -y \
    ca-certificates \
    libglib2.0-0 libjpeg-turbo8 libpng16-16 libopenexr24 \
    libwebp6 libwebpmux3 libwebpdemux2 libtiff5 libgif7 libexif12 libxml2 libpoppler-glib8 \
    libmagickwand-6.q16-6 libpango1.0-0 libmatio9 libopenslide0 \
    libgsf-1-114 fftw3 liborc-0.4-0 librsvg2-2 libcfitsio8 libimagequant0 libheif1 libbrotli1 && \
    useradd -r -u 1000 -g users mort && \
    mkdir -p /etc/mort && \
    chown -R mort:users /etc/mort && \
    apt-get autoremove -y && \
    apt-get autoclean && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* /var/cache/apt/archives/*

# Copy libvips libraries and application files
COPY --from=builder /usr/local/lib /usr/local/lib
COPY --from=builder /go/mort /usr/local/bin/mort
COPY --from=builder /workspace/configuration/config.yml /etc/mort/mort.yml
COPY --from=builder /workspace/configuration/parse.tengo /etc/mort/parse.tengo
COPY --from=builder /tmp/mime.types /etc/mime.types

# Update linker cache and verify installation
RUN ldconfig /usr/local/lib && \
    /usr/local/bin/mort -version

ENV MORT_CONFIG_DIR=/etc/mort

# Switch to non-root user
USER mort

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/usr/local/bin/mort", "-version"]

# Run the server
ENTRYPOINT ["/usr/local/bin/mort"]

# Expose the server TCP port
EXPOSE 8080 8081
