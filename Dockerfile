FROM docker-registry.geophy.com/docker/base-images/alpine:3.8
USER root
ARG COMMIT
ENV COMMIT ${COMMIT:-master}
RUN apk --update --no-cache add --virtual base autoconf automake build-base bash curl git snappy-dev libtool pkgconfig parallel gcc g++ go
RUN git clone https://github.com/openvenues/libpostal -b $COMMIT
COPY *.sh /libpostal/
WORKDIR /libpostal
RUN chmod +x *.sh
RUN ./build_libpostal.sh
RUN mkdir /libpostal/libpostalrest
COPY main.go /libpostal/
RUN GOPATH=/libpostal/libpostalrest go get github.com/gorilla/mux
RUN GOPATH=/libpostal/libpostalrest go get github.com/openvenues/gopostal/expand
RUN GOPATH=/libpostal/libpostalrest go get github.com/openvenues/gopostal/parser
RUN GOPATH=/libpostal/libpostalrest go get github.com/prometheus/client_golang/prometheus
RUN GOPATH=/libpostal/libpostalrest go get github.com/prometheus/client_golang/prometheus/promhttp
RUN GOPATH=/libpostal/libpostalrest go build .
RUN apk del base
USER nobody
EXPOSE 8080
CMD ["./libpostal"]
