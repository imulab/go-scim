##########################################################################
# GO BUILDER
##########################################################################
FROM golang:1.13-buster AS builder

WORKDIR /build/scim

COPY . ./

WORKDIR /build/scim/server

RUN make build

##########################################################################
# FINAL IMAGE
##########################################################################
FROM debian:buster-slim

RUN apt-get -yq update \
  && apt-get -yq install ca-certificates curl openssl \
  && apt-get -yq autoremove \
  && apt-get -yq clean \
  && rm -rf /var/lib/apt/lists/* \
  && truncate -s 0 /var/log/*log

# copy binary
COPY --from=builder /build/scim/server/bin/linux_amd64/scim /usr/bin/scim

##########################################################################
# Copying resource files between containers
# We should be using a bindata tool, like github.com/markbates/pkger
##########################################################################

# copy schemas
COPY --from=builder /build/scim/server/public/schemas /usr/share/scim/schemas

# copy resource types
COPY --from=builder /build/scim/server/public/resource_types /usr/share/scim/resource_types

# copy service provider configuration
COPY --from=builder /build/scim/server/public/sp_config /usr/share/scim/sp_config

# copy mongo metadata
COPY --from=builder /build/scim/mongo/public /usr/share/scim/mongo_metadata

# run binary
CMD ["/usr/bin/scim"]