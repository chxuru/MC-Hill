FROM golang:latest

WORKDIR /openvpn

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o openvpn .

CMD ["./openvpn"]
