FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum .

RUN go mod download

COPY . .
RUN GOOS=linux go build -v -o cowg-app

FROM ubuntu:noble

RUN apt update && apt install -y wireguard resolvconf iproute2 iptables

COPY --from=builder /app/cowg-app /usr/bin/cowg
RUN chmod +x /usr/bin/cowg

RUN echo 'net.ipv4.ip_forward = 1' | tee -a /etc/sysctl.d/99-wireguard-cowg.conf
RUN echo 'net.ipv6.conf.all.forwarding = 1' | tee -a /etc/sysctl.d/99-wireguard-cowg.conf
RUN sysctl -p /etc/sysctl.d/99-wireguard-cowg.conf

COPY ./configs/wg0.conf /etc/wireguard/wg0.conf
COPY ./scripts/entrypoint.sh /usr/bin/entrypoint.sh
RUN chmod +x /usr/bin/entrypoint.sh

EXPOSE 22022
EXPOSE 22080
EXPOSE 51820

ENTRYPOINT ["/usr/bin/entrypoint.sh"]
