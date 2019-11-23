package protocol

import (
	"github.com/imulab/go-scim/core"
	"github.com/imulab/go-scim/persistence"
	"github.com/imulab/go-scim/protocol/stage"
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
		err error
	)
	{
		_ = core.Schemas.MustLoad("../resource/schema/user_schema.json")
		_ = core.Meta.MustLoad("../resource/metadata/default_metadata.json", new(core.DefaultMetadataProvider))

		resourceType := core.ResourceTypes.MustLoad("../resource/resource_type/user_resource_type.json")
		httpProvider := NewDefaultHttpProvider()
		persistenceProvider := persistence.NewMemoryProvider()
		filterStage := stage.NewFilterStage([]*core.ResourceType{
			resourceType,
		}, []stage.PropertyFilter{
			stage.NewIDFilter(),
			stage.NewMetaResourceTypeFilter(),
			stage.NewMetaCreatedFilter(),
			stage.NewMetaLastModifiedFilter(),
			stage.NewMetaLocationFilter(map[string]string{resourceType.Id: "https://test.org/%s"}),
			stage.NewMetaVersionFilter(),
			stage.NewMutabilityFilter(),
			stage.NewRequiredFilter(),
			stage.NewCanonicalValueFilter(),
			stage.NewUniquenessFilter([]persistence.Provider{persistenceProvider}),
			stage.NewBCryptFilter(10),
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
