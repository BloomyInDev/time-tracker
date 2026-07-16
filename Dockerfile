# Build on the native platform and cross-compile to the target, so no QEMU
# emulation is needed and any Go-supported linux/arch works.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go tool templ generate

ARG TARGETOS TARGETARCH TARGETVARIANT
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} \
	go build -trimpath -ldflags "-s -w" -o /out/time-tracker ./cmd/time-tracker

# Static binary (CGO disabled, assets embedded), so a scratch image is enough.
FROM scratch

WORKDIR /app

COPY --from=build /out/time-tracker /app/

EXPOSE 8080
ENV DB_PATH=/data/time-tracker.db
VOLUME ["/data"]

ENTRYPOINT ["/app/time-tracker"]
CMD ["serve"]
