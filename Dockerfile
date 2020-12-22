ARG FROM=golang:1.15-buster

FROM $FROM as builder

ARG DEVUAN_KEY=BB23C00C61FC752C

# Debian's libudev-dev package doesn't include a static library (libudev.a) needed for static linking.
# Fortunately we can install libeudev-dev package from Devuan repositories which provides udev-compatible
# development files and a static library as well.
RUN echo deb http://deb.devuan.org/merged ascii main > /etc/apt/sources.list.d/devuan.list
RUN gpg --keyserver keys.gnupg.net --recv-key $DEVUAN_KEY && gpg -a --export $DEVUAN_KEY | apt-key add -

RUN apt update && apt install --no-install-recommends -y libusb-1.0-0-dev libeudev-dev

WORKDIR /go/mp707
ADD . .
RUN make static

FROM scratch
COPY --from=builder /go/mp707/cmd/mp707/mp707 /
ENTRYPOINT ["/mp707"]
