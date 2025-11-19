FROM golang:1.24.2-alpine AS build
RUN apk add --no-cache build-base

WORKDIR /app

COPY go.mod go.sum .
RUN go mod download

COPY . .
RUN GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -trimpath -o /dist/passman ./cmd/server
RUN ldd /dist/passman | tr -s [:blank:] '\n' | grep ^/ | xargs -I % install -D % /dist/%
RUN ln -s ld-musl-x86_64.so.1 /dist/lib/libc.musl-x86_64.so.1

FROM scratch

COPY --from=build /dist /
COPY --from=build /app/migrations /migrations
COPY --from=build /app/assets /assets

EXPOSE 5000

ENTRYPOINT ["/passman"]
