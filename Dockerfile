
ARG service


FROM golang:1.23.4-alpine
ENV HTTP_PORT=9092
ENV GRPC_PORT=8081
ARG service
WORKDIR /app
COPY . .
RUN go build -o main  $service/cmd/main.go 

# 2nd stage
FROM alpine:latest
ARG service
WORKDIR /app
COPY --from=0 /app/main .
EXPOSE $HTTP_PORT
EXPOSE $GRPC_PORT
CMD ["./main"]



# go docker file with env varaibles to specify the image name and tag and path to main.go
# docker build -t metadata:<tag> --build-arg service=<servicename> .