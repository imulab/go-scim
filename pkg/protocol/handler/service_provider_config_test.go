package handler

import (
	"encoding/json"
	"github.com/imulab/go-scim/pkg/core/spec"
	"github.com/imulab/go-scim/pkg/protocol/http"
	"github.com/imulab/go-scim/pkg/protocol/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"testing"
)

func TestServiceProviderConfigHandler(t *testing.T) {
	s := new(ServiceProviderConfigHandlerTestSuite)
	s.resourceBase = "../../tests/service_provider_config_handler_test_suite"
	suite.Run(t, s)
}

type ServiceProviderConfigHandlerTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *ServiceProviderConfigHandlerTestSuite) TestHandle() {
	tests := []struct {
		name       string
		getHandler func(t *testing.T) *ServiceProviderConfig
		getRequest func(t *testing.T) http.Request
		expect     func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: 	"get",
			getHandler: func(t *testing.T) *ServiceProviderConfig {
				return &ServiceProviderConfig{
					Log:   log.None(),
					SPC:   s.mustServiceProviderConfig("/service_provider_config.json"),
				}
			},
			getRequest: func(t *testing.T) http.Request {
				return http.DefaultRequest(httptest.NewRequest("GET", "/ServiceProviderConfig", nil), nil)
			},
			expect: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, 200, rr.Result().StatusCode)
				assert.NotEmpty(t, rr.Body.String())
				println(rr.Body.String())
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			resp := http.DefaultResponse(rr)
			test.getHandler(t).Handle(test.getRequest(t), resp)
			test.expect(t, rr)
		})
	}
}

func (s *ServiceProviderConfigHandlerTestSuite) mustServiceProviderConfig(filePath string) *spec.ServiceProviderConfig {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	spc := new(spec.ServiceProviderConfig)
	err = json.Unmarshal(raw, spc)
	s.Require().Nil(err)

	return spc
}