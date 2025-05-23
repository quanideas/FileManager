FROM golang:latest as builder
WORKDIR /build
COPY ./ ./
RUN go mod download
RUN CGO_ENABLED=0 go build -o ./main


FROM scratch
WORKDIR /app
COPY --from=builder /build/main ./main
ENTRYPOINT ["./main"]