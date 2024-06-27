# # Step 1: Modules caching
# FROM golang:1.23rc1-alpine3.20 as modules
# COPY go.mod go.sum /modules/
# WORKDIR /modules
# RUN go mod download

# # Step 2: Builder
# FROM golang:1.23rc1-alpine3.20 as builder
# COPY --from=modules /go/pkg /go/pkg
# COPY . /app
# WORKDIR /app
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
#     go build -tags migrate -o /bin/app ./cmd/app

# # Step 3: Final
# FROM scratch
# COPY --from=builder /app/config /config
# COPY --from=builder /app/migrations /migrations
# COPY --from=builder /bin/app /app
# COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# CMD ["/app"]



# Step 1: Modules caching
FROM golang:1.23rc1-alpine3.20 as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Builder
FROM golang:1.23rc1-alpine3.20 as builder
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -tags migrate -o /bin/app ./cmd/app

# Step 3: Final stage
FROM alpine:3.20

# Install CA certificates
RUN apk --no-cache add ca-certificates

# Copy built executable and necessary files from builder stage
COPY --from=builder /bin/app /bin/app
COPY --from=builder /app/config /app/config
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Expose port 8080 for your application
EXPOSE 8080

# Set the working directory
WORKDIR /bin

# Copy Swagger files and expose Swagger UI
COPY --from=builder /app/docs/swagger.json /app/docs/swagger.json
COPY --from=builder /app/docs/swagger.yaml /app/docs/swagger.yaml
COPY --from=builder /app/docs /swagger-ui

# Expose port 8081 for Swagger UI
EXPOSE 8081

# Set up CMD to run both application and Swagger UI
CMD /bin/sh -c "cd /swagger-ui && swag init -g /app/docs/swagger.yaml && swag init -g /app/docs/swagger.json -o . -d . -u && /bin/app"
