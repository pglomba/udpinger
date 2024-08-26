FROM golang:1.21.3

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY pkg/ pkg/
COPY example-configs/ example-configs/

RUN CGO_ENABLED=0 GOOS=linux go build -o udpinger ./cmd
