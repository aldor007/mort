FROM ghcr.io/aldor007/mort-base as builder

ENV GOLANG_VERSION 1.16.6
ARG TARGETARCH amd64
ARG TAG 'dev'
ARG COMMIT "master"
ARG DATE "now"

ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-$TARGETARCH.tar.gz

RUN curl -fsSL --insecure "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz \
  && tar -C /usr/local -xzf golang.tar.gz \
  && rm golang.tar.gz

ENV WORKDIR /workspace
ENV PATH /usr/local/go/bin:$PATH


WORKDIR $WORKDIR
COPY go.mod  ./
COPY go.sum ./
RUN go mod  download

COPY cmd/  $WORKDIR/cmd
COPY .godir ${WORKDIR}/.godir
COPY configuration/ ${WORKDIR}/configuration
COPY etc/ ${WORKDIR}/etc
COPY pkg/ ${WORKDIR}/pkg

RUN go build -ldflags="-X 'main.version=${TAG}' -X 'main.commit=${COMMIT}' -X 'main.date=${DATE}'" -o /go/mort ./cmd/mort/mort.go


FROM ubuntu:20.04

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install --no-install-recommends -y \
    libglib2.0-0 libjpeg-turbo8 libpng16-16 libopenexr24  ca-certificates  \
    libwebp6 libwebpmux3 libwebpdemux2 libtiff5 libgif7 libexif12 libxml2 libpoppler-glib8 \
    libmagickwand-6.q16-6 libpango1.0-0 libmatio-dev libopenslide0 \
    libgsf-1-114 fftw3 liborc-0.4-0 librsvg2-2 libcfitsio8 libimagequant0 libheif1  libbrotli-dev && \
    apt-get autoremove -y && \
    apt-get autoclean && \
    apt-get clean && \
    ldconfig /usr/local/lib && \
    rm -rf /tmp/* /var/tmp/*

RUN mkdir -p /etc/mort/
# clean up
RUN rm -rf /go/src; rm -rf /usr/include/

COPY --from=builder /usr/local/lib /usr/local/lib
RUN ldconfig
COPY --from=builder /go/mort /go/mort
COPY --from=builder /workspace/configuration/config.yml /etc/mort/mort.yml
COPY --from=builder /workspace/configuration/parse.tengo /etc/mort/parse.tengo
ENV MORT_CONFIG_DIR /etc/mort
# add mime types
ADD http://svn.apache.org/viewvc/httpd/httpd/branches/2.2.x/docs/conf/mime.types?view=co /etc/mime.types

RUN /go/mort -version
# Run the outyet command by default when the container starts.
ENTRYPOINT ["/go/mort"]
# Expose the server TCP port
EXPOSE 8080 8081
