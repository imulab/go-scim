package v2

import (
	"context"
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/crud"
	"github.com/imulab/go-scim/pkg/v2/crud/expr"
	"github.com/imulab/go-scim/pkg/v2/db"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
)

// Create a db.DB implementation that persists data in MongoDB. This implementation supports one-to-one correspondence
// of a SCIM resource type to a MongoDB collection.
//
// The database will attempt to create MongoDB indexes on attributes whose uniqueness is global or server, or that has
// been annotated with "@MongoIndex". For unique attributes, a unique MongoDB index will be created, otherwise, it is
// just an ordinary index. Any index creation error are treated as non-error and simply ignored.
//
// This implementation has limited capability of correctly performing field projection according to the specification.
// It dumbly treats the *crud.Projection parameter as it is without performing any sanitation. As a result, if any
// returned=always field is excluded, it will not be returned; similarly, if any returned=never field is included,
// it will be returned. It is expected by downstream calls to perform a pre-sanitation on the parameters or perform
// a post-guard operation to ensure no sensitive information is leaked.
//
// The "github.com/imulab/go-scim/pkg/v2/json" provides a post-guard operation in its serialization function to ensure
// returned=never parameters are never leaked. When used with this database, the only situation needs to be worried
// about is that returned=always parameter may not be returned at all when included intentionally in the "attributes"
// parameter list. This behaviour might be acceptable. If not, pre-sanitation of the projection list is required.
//
// If so desired, use Options().IgnoreProjection() to ignore projection altogether and return a complete version of
// the result every time.
//
// This implementation also has limited capability to correctly performing sorting according to the specification. The
// control is not as fine grained as in crud.SeekSortTarget. It can sort on singular type, but may fail if asked to sort
// on multiValued type, or a singular type within a multiValued type.
//
// This implementation do not directly use the SCIM attribute path to persist into MongoDB. Instead, it uses a concept
// of MongoDB persistence paths (or mongo paths). These mongo paths are introduced to provide an alternative name to
// SCIM path when SCIM path consists characters illegal to MongoDB. For instance, group.$ref attribute consists a dollar
// sign that cannot be used as part of field names in MongoDB. Similarly, schema extensions will almost always introduce
// illegal characters because schemas such as "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User" contain dot
// which is used as path separators in MongoDB. When this is the case, this package allows used to register
// metadata (see metadata.go) that can be associated with the target attribute in order to provide an alias to the SCIM
// path suitable to be persisted in MongoDB. When a metadata is associated to a target attribute, the metadata's MongoName
// or MongoPath will be used; otherwise, the attribute's Name and Path will be used.
//
// The atomicity of MongoDB is utilized to avoid explicit locking when modifying the resource. When performing Replace
// (which provides service to SCIM replace and SCIM patch) and Delete operations, the resources id and version is used
// as the criteria to match a document in store before carrying out the operation. If the provided id and version failed
// to match a document, a conflict error is returned instead of a notFound error. This is because caller already
// provided a resource as argument which was fetched from the database, hence, the resource by the id must have existed.
// The only reason that id and version failed to match would then because another process modified the resource concurrently.
// Therefore, conflict seems to be a reasonable error code.
func DB(resourceType *spec.ResourceType, coll *mongo.Collection, opt *DBOptions) db.DB {
	d := &mongoDB{
		resourceType: resourceType,
		superAttr:    resourceType.SuperAttribute(true),
		coll:         coll,
		t:            newTransformer(resourceType),
		opt:          opt,
	}
	d.ensureIndex()
	return d
}

type mongoDB struct {
	superAttr    *spec.Attribute
	resourceType *spec.ResourceType
	coll         *mongo.Collection
	t            *transformer
	opt          *DBOptions
}

func (d *mongoDB) Insert(ctx context.Context, resource *prop.Resource) error {
	_, err := d.coll.InsertOne(ctx, newBsonAdapter(resource), options.InsertOne())
	if err != nil {
		return fmt.Errorf("%w: %v", spec.ErrInternal, err)
	}
	return nil
}

func (d *mongoDB) Count(ctx context.Context, filter string) (int, error) {
	tf, err := d.mongoFilter(filter)
	if err != nil {
		return 0, err
	}

	n, err := d.coll.CountDocuments(ctx, tf, options.Count())
	if err != nil {
		return 0, fmt.Errorf("%w: %v", spec.ErrInternal, err)
	}

	return int(n), nil
}

func (d *mongoDB) Get(ctx context.Context, id string, projection *crud.Projection) (*prop.Resource, error) {
	opt := options.FindOne()
	if !d.opt.ignoreProjection && projection != nil {
		opt = opt.SetProjection(d.mongoProjection(projection))
	}

	tf, err := d.mongoFilter(fmt.Sprintf("id eq %s", strconv.Quote(id)))
	if err != nil {
		return nil, err
	}

	sr := d.coll.FindOne(ctx, tf, opt)
	if err := sr.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("%w: resource not found by id '%s'", spec.ErrNotFound, id)
		}
		return nil, fmt.Errorf("%w: %v", spec.ErrInternal, err)
	}

	w := newResourceUnmarshaler(d.resourceType)
	if err := sr.Decode(w); err != nil {
		return nil, err
	}

	return w.Resource(), nil
}

func (d *mongoDB) Replace(ctx context.Context, ref *prop.Resource, resource *prop.Resource) error {
	var (
		id      = ref.IdOrEmpty()
		version = ref.MetaVersionOrEmpty()
	)
	tf, err := d.mongoFilter(fmt.Sprintf("(id eq %s) and (meta.version eq %s)", strconv.Quote(id), strconv.Quote(version)))
	if err != nil {
		return err
	}

	sr := d.coll.FindOneAndReplace(ctx, tf, newBsonAdapter(resource), options.FindOneAndReplace())
	if err := sr.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return d.errNotFoundOrModified(id)
		}
		return err
	}

	return nil
}

func (d *mongoDB) Delete(ctx context.Context, resource *prop.Resource) error {
	var (
		id      = resource.IdOrEmpty()
		version = resource.MetaVersionOrEmpty()
	)
	tf, err := d.mongoFilter(fmt.Sprintf("(id eq %s) and (meta.version eq %s)", strconv.Quote(id), strconv.Quote(version)))
	if err != nil {
		return err
	}

	sr := d.coll.FindOneAndDelete(ctx, tf, options.FindOneAndDelete())
	if err := sr.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return d.errNotFoundOrModified(id)
		}
		return fmt.Errorf("%w: %v", spec.ErrInternal, err)
	}

	return nil
}

func (d *mongoDB) Query(ctx context.Context, filter string, sort *crud.Sort, pagination *crud.Pagination, projection *crud.Projection) ([]*prop.Resource, error) {
	opt := options.Find()

	tf, err := d.mongoFilter(filter)
	if err != nil {
		return nil, err
	}

	if sort != nil {
		opt.SetSort(d.mongoSort(sort))
	}
	if pagination != nil {
		skip, limit := d.mongoPagination(pagination)
		opt.SetSkip(skip)
		opt.SetLimit(limit)
	}
	if !d.opt.ignoreProjection && projection != nil {
		opt.SetProjection(d.mongoProjection(projection))
	}

	cursor, err := d.coll.Find(ctx, tf, opt)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", spec.ErrInternal, err)
	}

	defer func() {
		_ = cursor.Close(ctx)
	}()

	results := make([]*prop.Resource, 0)
	for cursor.Next(ctx) {
		w := newResourceUnmarshaler(d.resourceType)
		if err := cursor.Decode(w); err != nil {
			return nil, err
		}
		results = append(results, w.Resource())
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// Traverse the attributes structure along the tokens in the given path and
// return the path used in mongoDB persistence.
//
// The MongoDB persistence may be different with the SCIM attribute path so as to avoid introducing prohibited
// tokens such as "$" in mongoDB paths. The MongoDB persistence path, if necessary, should be registered in the
// metadata (see metadata.go). If there's no registered metadata associated with the target attribute, the path
// of the attribute will be used.
//
// If this method is unable to find a path, or encounters any error, an empty string is returned.
func (d *mongoDB) mongoPathFor(path string) string {
	curAttr := d.superAttr
	cursor, err := expr.CompilePath(path)
	if err != nil {
		return ""
	}

	// skip the first token in the path starts with the id of the resource type's default schema.
	// For instance, "urn:ietf:params:scim:schemas:core:2.0:User:userName" should just be treated as "userName"
	if cursor.Token() == d.resourceType.Schema().ID() {
		cursor = cursor.Next()
	}
	if cursor == nil {
		return ""
	}

	for cursor != nil {
		curAttr = curAttr.SubAttributeForName(cursor.Token())
		if curAttr == nil {
			return ""
		}
		cursor = cursor.Next()
	}

	mp := curAttr.Path()
	if md, ok := metadataHub[curAttr.ID()]; ok {
		mp = md.MongoPath
	}

	return mp
}

// Convert the crud.Sort structure to MongoDB driver compatible bson.D structure, so that it can be serialized by the
// driver. The supplied sort parameter must not be nil. If the sort.By is empty, or sort.By cannot resolve its
// corresponding MongoDB persistence path, sort is done on the internal "_id" field instead.
func (d *mongoDB) mongoSort(sort *crud.Sort) bson.D {
	var by string
	{
		if len(sort.By) > 0 {
			by = d.mongoPathFor(sort.By)
		}
		if len(by) == 0 {
			by = "_id"
		}
	}

	switch sort.Order {
	case crud.SortAsc, crud.SortDefault:
		return bson.D{{Key: by, Value: 1}}
	case crud.SortDesc:
		return bson.D{{Key: by, Value: -1}}
	default:
		panic("invalid sort order")
	}
}

// Convert crud.Pagination parameter to Mongo compatible option parameters. The supplied pagination parameter
// must not be nil.
func (d *mongoDB) mongoPagination(pagination *crud.Pagination) (skip int64, limit int64) {
	skip = int64(pagination.StartIndex - 1)
	limit = int64(pagination.Count)
	return
}

// Convert the crud.Projection parameter to Mongo driver compatible bson.D structure. The supplied projection
// parameter must not be nil and should conform to the constraint that only one of "attributes" and "excludedAttributes"
// shall be used. This method does not further check for that constraint. If a given path cannot resolve its MongoDB
// persistence path, it will be skipped.
func (d *mongoDB) mongoProjection(projection *crud.Projection) bson.D {
	if len(projection.Attributes) > 0 {
		include := bson.D{}
		for _, p := range projection.Attributes {
			if mp := d.mongoPathFor(p); len(mp) > 0 {
				include = append(include, bson.E{Key: mp, Value: 1})
			}
		}
		return include
	}

	if len(projection.ExcludedAttributes) > 0 {
		exclude := bson.D{}
		for _, p := range projection.Attributes {
			if mp := d.mongoPathFor(p); len(mp) > 0 {
				exclude = append(exclude, bson.E{Key: mp, Value: 0})
			}
		}
		return exclude
	}

	return bson.D{}
}

// Convert the SCIM filter to MongoDB driver compatible bson.D structure. This method uses transformer (see filter.go)
// to transform the compiled abstract syntax tree of the filter to bson.D containing MongoDB filter directives.
func (d *mongoDB) mongoFilter(filter string) (bson.D, error) {
	cf, err := expr.CompileFilter(filter)
	if err != nil {
		return nil, err
	}
	tf, err := d.t.transform(cf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (d *mongoDB) errNotFoundOrModified(id string) error {
	return fmt.Errorf("%w: resource by id '%s' was not found or was modified since by another request", spec.ErrConflict, id)
}

// DB options
func Options() *DBOptions {
	return &DBOptions{}
}

type DBOptions struct {
	ignoreProjection bool
}

// Ask the database to ignore any projection parameters. This might be reasonable when the downstream services
// wish to perform further actions on the complete version of the resource.
func (opt *DBOptions) IgnoreProjection() *DBOptions {
	opt.ignoreProjection = true
	return opt
}

var (
	_ db.DB = (*mongoDB)(nil)
)
