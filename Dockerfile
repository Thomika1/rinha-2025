#FIRST STEP
FROM golang:1.23-alpine as stage1

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLE=0 GOOS=linux go build -o server
# SECOND STEP
FROM scratch

COPY --from=stage1 /app/server /

ENTRYPOINT [ "/server" ]

EXPOSE 9999

CMD ["./server"]



