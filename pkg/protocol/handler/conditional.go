package handler

import (
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/protocol/http"
	"strings"
)

func interpretConditionalHeader(request http.Request) func(r *prop.Resource) bool {
	if ifMatch := request.Header("If-Match"); len(ifMatch) > 0 {
		return func(r *prop.Resource) bool {
			version := r.Version()
			ifMatch = strings.TrimSpace(ifMatch)
			if ifMatch == "*" {
				return true
			}
			for _, each := range strings.Split(ifMatch, ",") {
				if strings.TrimSpace(each) == version {
					return true
				}
			}
			return false
		}
	}

	if ifNoneMatch := request.Header("If-None-Match"); len(ifNoneMatch) > 0 {
		return func(r *prop.Resource) bool {
			version := r.Version()
			ifNoneMatch = strings.TrimSpace(ifNoneMatch)
			if ifNoneMatch == "*" {
				return false
			}
			for _, each := range strings.Split(ifNoneMatch, ",") {
				if strings.TrimSpace(each) == version {
					return false
				}
			}
			return true
		}
	}

	return func(r *prop.Resource) bool {
		return true
	}
}
