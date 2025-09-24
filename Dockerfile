# --- build stage ---
FROM golang:1.25rc3-bullseye AS builder

WORKDIR /app

# 安装构建 gosseract 需要的头文件
RUN apt-get update && apt-get install -y \
    libtesseract-dev \
    libleptonica-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /ocrserver .

# --- runtime stage ---
FROM debian:bullseye-slim

ARG LOAD_LANG=jpn

RUN apt-get update && apt-get install -y \
    ca-certificates \
    tesseract-ocr \
    && rm -rf /var/lib/apt/lists/*

# 可选语言包
RUN if [ -n "${LOAD_LANG}" ]; then apt-get update && apt-get install -y tesseract-ocr-${LOAD_LANG} && rm -rf /var/lib/apt/lists/*; fi

COPY --from=builder /ocrserver /usr/local/bin/ocrserver

ENV PORT=8080
CMD ["ocrserver"]