FROM golang:1.22 AS builder

COPY . /src
WORKDIR /src

RUN apt-get update && apt-get install -y git

RUN GOPROXY=https://goproxy.cn make build

FROM debian:12-slim

# 手动创建 sources.list 文件（之前不存在）
RUN echo "deb http://mirrors.aliyun.com/debian/ bookworm main" > /etc/apt/sources.list && \
    echo "deb http://mirrors.aliyun.com/debian/ bookworm-updates main" >> /etc/apt/sources.list && \
    echo "deb http://mirrors.aliyun.com/debian-security bookworm-security main" >> /etc/apt/sources.list

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/bin /app

WORKDIR /app

EXPOSE 8000
EXPOSE 9000
VOLUME ["/app/configs", "/data/conf"]

CMD ["./micro-scheduler", "-conf", "configs"]
