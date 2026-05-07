package v2

import (
	"context"
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
)

const (
	// @MongoIndex annotates a field so that a corresponding index is generated in MongoDB. If uniqueness=server on
	// the annotated field, a unique index is generated. Otherwise, an ordinary index is generated.
	AnnotationMongoIndex = "@MongoIndex"
)

func (d *mongoDB) ensureIndex() {
	d.superAttr.DFS(func(a *spec.Attribute) {
		if a.Uniqueness() == spec.UniquenessNone {
			return
		}
		if _, ok := a.Annotation(AnnotationMongoIndex); !ok {
			return
		}

		path := a.Path()
		if md, ok := metadataHub[a.ID()]; ok {
			path = md.MongoPath
		}

		idm := mongo.IndexModel{
			Keys:    bson.D{{Key: path, Value: 1}},
			Options: options.Index(),
		}
		if a.Uniqueness() == spec.UniquenessServer || a.Uniqueness() == spec.UniquenessGlobal {
			idm.Options.SetUnique(true)
		}
		if name := fmt.Sprintf("idx_%s", strings.Replace(path, ".", "_", -1)); len(name) < 127 {
			// https://docs.mongodb.com/manual/reference/command/createIndexes/
			// For MongoDB 4.0 and earlier, the index name has a limit of 127 bytes, here we still adhere to this
			// constraint without checking for server version. If the formed name is greater than 127 bytes, we will
			// just let MongoDB choose a random name.
			idm.Options.SetName(name)
		}

		_, err := d.coll.Indexes().CreateOne(context.Background(), idm, options.CreateIndexes())
		if err != nil {
			// https://docs.mongodb.com/manual/reference/command/createIndexes/
			// Starting from MongoDB 4.2, MongoDB will return error if the index was already created. Previous
			// version will return ok to indicate implicit success. Here, we regard any error as "not really an
			// error" and only display warning information to logger.
			return
		}
		return
	})
}
