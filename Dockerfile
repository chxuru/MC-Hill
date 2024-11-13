FROM golang:latest

WORKDIR /openvpn

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o openvpn .

RUN apt-get update && \
    apt-get install -y curl unzip && \
    curl -L https://github.com/bitwarden/cli/releases/download/v1.20.0/bw-linux-1.20.0.zip -o bw.zip && \
    unzip bw.zip -d /usr/local/bin && \
    chmod +x /usr/local/bin/bw && \
    rm bw.zip && \
    apt-get clean

CMD ["./openvpn"]
