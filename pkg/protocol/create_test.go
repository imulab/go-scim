package protocol

import (
	"github.com/imulab/go-scim/pkg/core"
	"github.com/imulab/go-scim/pkg/persistence"
	"github.com/imulab/go-scim/pkg/protocol/stage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCreateEndpoint(t *testing.T) {
	var (
		endpoint *CreateEndpoint
		err      error
	)
	{
		_ = core.Schemas.MustLoad("../resource/schema/user_schema.json")
		_ = core.Meta.MustLoad("../resource/metadata/default_metadata.json", new(core.DefaultMetadataProvider))

		resourceType := core.ResourceTypes.MustLoad("../resource/resource_type/user_resource_type.json")
		httpProvider := NewDefaultHttpProvider(nil)
		persistenceProvider := persistence.NewMemoryProvider(resourceType)
		filterStage := stage.NewFilterStage([]*core.ResourceType{
			resourceType,
		}, []stage.PropertyFilter{
			stage.NewReadOnlyFilter(1000),
			stage.NewSchemaFilter(1010),
			stage.NewIDFilter(1020),
			stage.NewMetaResourceTypeFilter(1030),
			stage.NewMetaCreatedFilter(1031),
			stage.NewMetaLastModifiedFilter(1032),
			stage.NewMetaLocationFilter(map[string]string{resourceType.Id: "https://test.org/%s"}, 1033),
			stage.NewMetaVersionFilter(1034),
			stage.NewBCryptFilter(10, 1040),
			stage.NewMutabilityFilter(2000),
			stage.NewRequiredFilter(2010),
			stage.NewCanonicalValueFilter(2020),
			stage.NewUniquenessFilter([]persistence.Provider{persistenceProvider}, 2030),
		})

		endpoint = &CreateEndpoint{
			HttpProvider:        httpProvider,
			ResourceType:        resourceType,
			FilterStage:         filterStage,
			PersistenceProvider: persistenceProvider,
		}
	}

	f, err := os.Open("../resource/test/test_valid_user_create_payload.json")
	require.Nil(t, err)

	req := httptest.NewRequest(http.MethodPost, "/User", f)
	req.Header.Set(HeaderContentType, ContentTypeApplicationJsonScim)

	rr := httptest.NewRecorder()

	endpoint.ServeHTTP(rr, req)

	assert.Equal(t, 200, rr.Code)
	assert.NotEmpty(t, rr.Body.String())
	assert.Equal(t, ContentTypeApplicationJsonScim, rr.Header().Get(HeaderContentType))
}
