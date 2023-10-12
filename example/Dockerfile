# syntax=docker/dockerfile:1

FROM golang:1.21 as builder

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build

FROM golang:1.21 as runner

COPY --from=builder /app/example /bin

RUN ls /bin/
RUN cat /bin/example

# Run
CMD ["/bin/example"]