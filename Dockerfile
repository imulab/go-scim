##########################################################################
# GO BUILDER
##########################################################################
FROM golang:1.13-buster AS builder

WORKDIR /build/scim
COPY go.mod ./
COPY pkg/v2/go.mod ./pkg/v2/go.mod
COPY mongo/v2/go.mod ./mongo/v2/go.mod
COPY Makefile ./
RUN make deps

COPY . ./
RUN make binary

##########################################################################
# FINAL IMAGE
##########################################################################
FROM debian:buster-slim

# copy binary
COPY --from=builder /build/scim/bin/linux_amd64/scim /usr/bin/scim

# copy public files
COPY --from=builder /build/scim/public /usr/share/scim/public

# run
CMD ["/usr/bin/scim"]
