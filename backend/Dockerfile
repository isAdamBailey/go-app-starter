FROM golang:1.27-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# distroless: no shell, no package manager, minimal attack surface. Runs as
# a non-root user by default.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/server /usr/local/bin/server
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/server"]
