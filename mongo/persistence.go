package mongo

import (
	"context"
	"fmt"
	"github.com/imulab/go-scim/core"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"sync"
	"time"
)

type (
	persistenceProvider struct {
		// List of resource type that this provider manages. The size of the list must match
		// that of collections.
		resourceTypes []*core.ResourceType
		// MongoDB collection corresponding to the id of the resource type that it manages.
		// The size of this map must match that of resourceTypes.
		collections map[string]*mongo.Collection
		// An integer between 0 and 100 that represents the percentage of the context deadline
		// that the operations in this provider should finish in. When setting to 100, we are
		// using the full deadline specified by the context; when setting to a number less than
		// 100, we are using a portion of the context deadline as our own deadline, effectively
		// saving some time for other operations that might have been carried out under the same
		// context. If the context does not specify a deadline, a default of 30 seconds is used.
		maxTimePercent int
	}

	// Generic wrapper of the results of a function call.
	singleResult struct {
		result interface{}
		err    error
	}

	// Function wrapper to execute to a single collection and return a result
	singleFunc func(resourceType *core.ResourceType) singleResult
)

func (p *persistenceProvider) IsFilterSupported() bool {
	return true
}

func (p *persistenceProvider) IsPaginationSupported() bool {
	return true
}

func (p *persistenceProvider) IsSortSupported() bool {
	return true
}

func (p *persistenceProvider) IsResourceTypeSupported(resourceType *core.ResourceType) bool {
	_, ok := p.collections[resourceType.Id]
	return ok
}

func (p *persistenceProvider) Total(ctx context.Context) (int64, error) {
	if len(p.collections) == 1 {
		return p.totalSingleCollection(ctx, p.resourceTypes[0].Id)
	}
	return p.totalAllCollection(ctx)
}

func (p *persistenceProvider) totalSingleCollection(ctx context.Context, resourceTypeId string) (int64, error) {
	maxTime := p.getMaxTime(ctx)
	n, err := p.collections[resourceTypeId].EstimatedDocumentCount(ctx, &options.EstimatedDocumentCountOptions{
		MaxTime: &maxTime,
	})
	if err != nil {
		return 0, p.errDatabase(err)
	}
	return n, nil
}

func (p *persistenceProvider) totalAllCollection(ctx context.Context) (int64, error) {
	var total int64 = 0

	results := p.withAllResourceTypes(ctx, func(resourceType *core.ResourceType) singleResult {
		n, err := p.totalSingleCollection(ctx, resourceType.Id)
		return singleResult{
			result: n,
			err:    err,
		}
	})

	for res := range results {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			if res.err != nil {
				return 0, res.err
			} else {
				total += res.result.(int64)
			}
		}
	}

	return total, nil
}

func (p *persistenceProvider) Count(ctx context.Context, scimFilter string) (int64, error) {
	if len(p.collections) == 1 {
		return p.countSingleCollection(ctx, scimFilter, p.resourceTypes[0].Id)
	}
	return p.countAllCollection(ctx, scimFilter)
}

func (p *persistenceProvider) countSingleCollection(ctx context.Context, scimFilter string, resourceTypeId string) (int64, error) {
	filter, err := TransformFilter(scimFilter, p.resourceTypes[0])
	if err != nil {
		return 0, err
	}

	n, err := p.collections[resourceTypeId].CountDocuments(ctx, filter, options.Count())
	if err != nil {
		return 0, p.errDatabase(err)
	}

	return n, nil
}

func (p *persistenceProvider) countAllCollection(ctx context.Context, scimFilter string) (int64, error) {
	var count int64 = 0

	results := p.withAllResourceTypes(ctx, func(resourceType *core.ResourceType) singleResult {
		n, err := p.countSingleCollection(ctx, scimFilter, resourceType.Id)
		return singleResult{
			result: n,
			err:    err,
		}
	})

	for res := range results {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			if res.err != nil {
				return 0, res.err
			} else {
				count += res.result.(int64)
			}
		}
	}

	return count, nil
}

func (p *persistenceProvider) InsertOne(ctx context.Context, resource *core.Resource) error {
	resourceType := resource.GetResourceType()
	if !p.IsResourceTypeSupported(resourceType) {
		return core.Errors.Internal(fmt.Sprintf("resource type '%s' is not supported by this persistence provider", resourceType.Id))
	}

	collection := p.collections[resourceType.Id]
	_, err := collection.InsertOne(ctx, newBsonAdapter(resource), options.InsertOne())
	if err != nil {
		return core.Errors.Internal(fmt.Sprintf("failed to create resource: %s", err.Error()))
	}

	return nil
}

func (p *persistenceProvider) GetById(ctx context.Context, id string) (*core.Resource, error) {
	if len(p.collections) == 1 {
		return p.getByIdFromSingleCollection(ctx, id, p.resourceTypes[0])
	}
	return p.getByIdFromAllCollection(ctx, id)
}

func (p *persistenceProvider) getByIdFromSingleCollection(ctx context.Context, id string, resourceType *core.ResourceType) (*core.Resource, error) {
	sr := p.collections[resourceType.Id].FindOne(ctx, p.byIdFilter(id), options.FindOne())
	if err := sr.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, core.Errors.NotFound("%s by id '%s' is not found", resourceType.Name, id)
		}
		return nil, p.errDatabase(err)
	}

	unmarshaler := newResourceUnmarshaler(resourceType)
	if err := sr.Decode(unmarshaler); err != nil {
		return nil, err
	}

	return unmarshaler.resource, nil
}

func (p *persistenceProvider) getByIdFromAllCollection(ctx context.Context, id string) (*core.Resource, error) {
	results := p.withAllResourceTypes(ctx, func(resourceType *core.ResourceType) singleResult {
		r, err := p.getByIdFromSingleCollection(ctx, id, resourceType)
		return singleResult{
			result: r,
			err:    err,
		}
	})

	var (
		resource *core.Resource
		err      error
	)

	for res := range results {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			if res.err != nil {
				err = res.err
				continue
			} else {
				resource = res.result.(*core.Resource)
				break
			}
		}
	}

	if err != nil {
		return nil, err
	}
	return resource, nil
}

// Execute the given single function on all collections/resourceTypes available and return a channel of all their results.
// The call will execute the single function on each collection/resourceType in parallel and block until they are all
// completed. The returned channel is buffered, and will be closed before returning, so the caller don't need to burden
// itself with closing the channel.
func (p *persistenceProvider) withAllResourceTypes(ctx context.Context, job singleFunc) <-chan singleResult {
	var (
		wg      sync.WaitGroup
		resChan = make(chan singleResult, len(p.collections))
	)

	wg.Add(len(p.collections))

	for _, resourceType := range p.resourceTypes {
		go func(ctx context.Context, resourceType *core.ResourceType, resChan chan singleResult) {
			defer wg.Done()
			sr := job(resourceType)
			select {
			case resChan <- sr:
				return
			default:
				return
			}
		}(ctx, resourceType, resChan)
	}

	wg.Wait()
	close(resChan)

	return resChan
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
