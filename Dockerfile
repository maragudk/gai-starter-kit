FROM --platform=${BUILDPLATFORM} debian:stable AS cssbuilder
WORKDIR /app

RUN set -x && apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get install -y curl

# The URL uses x64 instead of amd64
ARG BUILDARCH
RUN ARCH=$( [ "${BUILDARCH}" = "amd64" ] && echo "x64" || echo "arm64" ) && \
  curl -sfLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-${ARCH}
RUN mv tailwindcss-linux-* tailwindcss
RUN chmod a+x tailwindcss

COPY tailwind.css ./

COPY html ./html/

RUN ./tailwindcss -i tailwind.css -o app.css --minify



FROM golang AS gobuilder
WORKDIR /app

RUN echo "deb http://deb.debian.org/debian unstable main" >>/etc/apt/sources.list
RUN set -x && apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get install -y libsqlite3-dev/unstable

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

ARG TARGETARCH
RUN GOOS=linux GOARCH=${TARGETARCH} CGO_ENABLED=1 go build -tags sqlite_fts5,libsqlite3 -buildvcs=false -ldflags="-s -w" -o ./app ./cmd/app



FROM debian:stable-slim AS runner
WORKDIR /app

RUN set -x && apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates && \
  rm -rf /var/lib/apt/lists/*

COPY public ./public/
COPY --from=cssbuilder /app/app.css ./public/styles/
COPY --from=gobuilder /app/app ./

EXPOSE 8080

CMD ["./app"]
