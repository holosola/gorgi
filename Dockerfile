FROM golang:1.22.2-alpine as build
RUN mkdir /build
COPY ./ /build
WORKDIR /build
RUN go build . -o gorgiSrv


FROM alpine:latest as runtime
RUN mkdir /gorgi
WORKDIR /gorgi
COPY --from=builder /build/gorgiSrv /gorgi/
COPY ./configs /gorgi/configs
ENTRYPOINT [ "/gorgi/gorgiSrv" ]