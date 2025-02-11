FROM golang:1.23 AS build
WORKDIR /app
COPY . .
RUN go build -o backend

FROM alpine:latest
RUN apk add --no-cache libc6-compat 
WORKDIR /root/
COPY --from=build /app/backend .
RUN chmod +x /root/backend
EXPOSE 8080
CMD ["./backend"]