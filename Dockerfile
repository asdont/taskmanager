FROM golang:alpine AS builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod vendor -a -installsuffix cgo -o taskmanager cmd/app/main.go

FROM alpine:latest
WORKDIR /taskmanager
COPY --from=builder ./build/taskmanager .
COPY --from=builder ./build/configs/ /taskmanager/configs/
COPY --from=builder ./build/docs/ /taskmanager/docs/
RUN mkdir "logs"
CMD ["./taskmanager"]