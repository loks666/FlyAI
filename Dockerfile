FROM golang:1.23-alpine AS builder
WORKDIR /builder
COPY . .

#ENV GO111MODULE=on
#ENV GOPROXY=https://goproxy.cn,direct

RUN go version
RUN go mod download && go build -o flyai
RUN ls -lh && chmod -R +x ./*

FROM alpine:3.20 AS runner
LABEL org.opencontainers.image.source="https://github.com/loks666/FlyAI"

EXPOSE 19099
EXPOSE 9528

WORKDIR /app
COPY --from=builder /builder/flyai ./flyai
COPY --from=builder /builder/views ./views
CMD ["./flyai"]
