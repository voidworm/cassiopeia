FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder
WORKDIR /src

COPY go.mod ./
COPY . .
RUN go mod tidy
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/cassiopeia-backend .

FROM gcr.io/distroless/static-debian12
COPY --from=builder /out/cassiopeia-backend /cassiopeia-backend
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/cassiopeia-backend"]
