##########################################################################
# GO BUILDER
##########################################################################
FROM golang:1.13-buster AS builder

WORKDIR /build/scim

COPY . ./

RUN make build

##########################################################################
# FINAL IMAGE
##########################################################################
FROM debian:buster-slim
ARG TARGETARCH
# copy binary
COPY --from=builder /build/scim/bin/linux_${TARGETARCH}/scim /usr/bin/scim

# copy public files
COPY --from=builder /build/scim/public /usr/share/scim/public

# run
CMD ["/usr/bin/scim"]
