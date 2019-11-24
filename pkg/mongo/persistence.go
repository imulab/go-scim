package mongo

import (
	"context"
	"errors"
	"fmt"
	"github.com/imulab/go-scim/pkg/core"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"time"
)

type (
	persistenceProvider struct {
		// The resource type this provider supports
		resourceType *core.ResourceType
		// MongoDB collection that corresponds to the resource type.
		collection *mongo.Collection
		// An integer between 0 and 100 that represents the percentage of the context deadline
		// that the operations in this provider should finish in. When setting to 100, we are
		// using the full deadline specified by the context; when setting to a number less than
		// 100, we are using a portion of the context deadline as our own deadline, effectively
		// saving some time for other operations that might have been carried out under the same
		// context. If the context does not specify a deadline, a default of 30 seconds is used.
		maxTimePercent int
	}
)

func (p *persistenceProvider) SupportsFilter() bool {
	return true
}

func (p *persistenceProvider) SupportsPagination() bool {
	return true
}

func (p *persistenceProvider) SupportsSort() bool {
	return true
}

func (p *persistenceProvider) ResourceType() *core.ResourceType {
	return p.resourceType
}

func (p *persistenceProvider) Total(ctx context.Context) (int64, error) {
	maxTime := p.getMaxTime(ctx)
	n, err := p.collection.EstimatedDocumentCount(ctx, &options.EstimatedDocumentCountOptions{
		MaxTime: &maxTime,
	})
	if err != nil {
		return 0, p.errDatabase(err)
	}
	return n, nil
}

func (p *persistenceProvider) Count(ctx context.Context, scimFilter string) (int64, error) {
	filter, err := TransformFilter(scimFilter, p.resourceType)
	if err != nil {
		return 0, err
	}

	n, err := p.collection.CountDocuments(ctx, filter, options.Count())
	if err != nil {
		return 0, p.errDatabase(err)
	}

	return n, nil
}

func (p *persistenceProvider) InsertOne(ctx context.Context, resource *core.Resource) error {
	if err := p.ensureResourceTypeMatch(resource); err != nil {
		return err
	}

	_, err := p.collection.InsertOne(ctx, newBsonAdapter(resource), options.InsertOne())
	if err != nil {
		return core.Errors.Internal(fmt.Sprintf("failed to create resource: %s", err.Error()))
	}

	return nil
}

func (p *persistenceProvider) ReplaceOne(ctx context.Context, replacement *core.Resource) error {
	if err := p.ensureResourceTypeMatch(replacement); err != nil {
		return err
	}

	id, err := replacement.GetID()
	if err != nil {
		return err
	}

	ur, err := p.collection.ReplaceOne(ctx, p.byIdFilter(id), newBsonAdapter(replacement), options.Replace())
	if err != nil {
		return p.errDatabase(err)
	}

	if ur.MatchedCount == 0 {
		return p.errNotFoundById(id)
	} else if ur.ModifiedCount == 0 {
		return p.errDatabase(errors.New("no resource was modified"))
	}

	return nil
}

func (p *persistenceProvider) ensureResourceTypeMatch(resource *core.Resource) error {
	resourceType := resource.GetResourceType()
	if resource.GetResourceType().Id == p.resourceType.Id {
		return nil
	}
	return core.Errors.Internal("resource type '%s' is not supported by this persistence provider", resourceType.Id)
}

func (p *persistenceProvider) GetById(ctx context.Context, id string) (*core.Resource, error) {
	sr := p.collection.FindOne(ctx, p.byIdFilter(id), options.FindOne())
	if err := sr.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, p.errNotFoundById(id)
		}
		return nil, p.errDatabase(err)
	}

	unmarshaler := newResourceUnmarshaler(p.resourceType)
	if err := sr.Decode(unmarshaler); err != nil {
		return nil, err
	}

	return unmarshaler.resource, nil
}

// Return a mongo filter in the form {"id": "<id>"}
func (p *persistenceProvider) byIdFilter(id string) bsonx.Val {
	return bsonx.Document(bsonx.MDoc{
		"id": bsonx.String(id),
	})
}

// Return the duration of the maximum time allowed on an operation. This is determined by examining the deadline passed
// in from the context and the maxTimePercent. If no deadline was passed in, the default duration is 30 seconds. If a
// deadline was passed in, a percentage of maxTimePercent will be used to deduce the time allowed for operation in this
// persistence provider.
func (p *persistenceProvider) getMaxTime(ctx context.Context) (maxTime time.Duration) {
	maxTime = 30 * time.Second

	deadline, ok := ctx.Deadline()
	if !ok {
		// no deadline set on context, return default
		return
	}

	maxTime = deadline.Sub(time.Now())
	if maxTime < 0 {
		maxTime = 0
	}
	maxTime = time.Duration(maxTime.Nanoseconds() * int64(p.maxTimePercent) / 100)
	return
}

func (p *persistenceProvider) errDatabase(err error) error {
	return core.Errors.Internal("database error: " + err.Error())
}

func (p *persistenceProvider) errNotFoundById(id string) error {
	return core.Errors.NotFound("%s by id '%s' is not found", p.resourceType.Name, id)
}