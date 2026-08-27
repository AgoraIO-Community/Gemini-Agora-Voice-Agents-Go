# Build stage
FROM golang:1.23-bookworm AS build

WORKDIR /src

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
RUN CGO_ENABLED=0 go build -o /out/agent .

# Runtime stage
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/agent /agent

EXPOSE 8000

ENTRYPOINT ["/agent"]
