FROM golang:1.22.1 as builder

WORKDIR /app

COPY go.* ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 go build -v -o main ./main.go

FROM gcr.io/distroless/static:nonroot 
COPY --from=builder --chown=nonroot:nonroot /app/main /main

ENV PORT=1000
EXPOSE 1000

CMD ["/main"]