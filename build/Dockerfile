# Stage 1: Build
FROM golang:1.26.5-alpine3.24 AS build

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-s -w" -o service ./cmd/heimdall

# Stage 2: Distroless static image
FROM gcr.io/distroless/static-debian13:nonroot

# Copy the binary from the builder stage
COPY --from=build /app/service /service
COPY --from=build /app/build/models.yaml /models.yaml
COPY --from=build /app/migrations /migrations

# Distroless images include a nonroot user by default
USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/service"]