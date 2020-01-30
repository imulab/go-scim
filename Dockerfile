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
FROM alpine:3.11

# copy binary
COPY --from=builder /build/scim/bin/linux_amd64/scim /usr/bin/scim

# copy public files
COPY --from=builder /build/scim/public /usr/share/scim/public

# run
CMD ["/usr/bin/scim"]