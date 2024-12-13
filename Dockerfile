# go docker file with env varaibles to specify the image name and tag and path to main.go
# docker build -t metadata:<tag> --build-arg service=<servicename> .

FROM golang:1.23.4-alpine
ARG service
WORKDIR /app
COPY . .
RUN go build -o main  $service/cmd/main.go 

# 2nd stage
FROM alpine:latest
WORKDIR /app
COPY --from=0 /app/main .
CMD ["./main"]