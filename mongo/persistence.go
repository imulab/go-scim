package mongo

import (
	"context"
	"fmt"
	"github.com/imulab/go-scim/core"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

type persistenceProvider struct {
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
		return p.totalSingle(ctx, p.resourceTypes[0].Id)
	}
	return p.totalAll(ctx)
}

func (p *persistenceProvider) totalSingle(ctx context.Context, resourceTypeId string) (int64, error) {
	maxTime := p.getMaxTime(ctx)
	n, err := p.collections[resourceTypeId].EstimatedDocumentCount(ctx, &options.EstimatedDocumentCountOptions{
		MaxTime: &maxTime,
	})
	if err != nil {
		return 0, p.errDatabase(err)
	}
	return n, nil
}

func (p *persistenceProvider) totalAll(ctx context.Context) (int64, error) {
	type (
		result struct {
			n   int64
			err error
		}
	)

	var (
		total   int64
		wg      sync.WaitGroup
		resChan = make(chan result, len(p.collections))
	)

	wg.Add(len(p.collections))

	for k := range p.collections {
		go func(ctx context.Context, resourceTypeId string, resChan chan result) {
			defer wg.Done()

			n, err := p.totalSingle(ctx, resourceTypeId)
			select {
			case resChan <- result{n, err}:
				return
			default:
				return
			}
		}(ctx, k, resChan)
	}

	wg.Wait()
	close(resChan)

	for res := range resChan {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			if res.err != nil {
				return 0, res.err
			} else {
				total += res.n
			}
		}
	}

	return total, nil
}

func (p *persistenceProvider) Count(ctx context.Context, scimFilter string) (int64, error) {
	panic("implement me")
}

func (p *persistenceProvider) InsertOne(ctx context.Context, resource *core.Resource) error {
	resourceType := resource.GetResourceType()
	if p.IsResourceTypeSupported(resourceType) {
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
	panic("implement me")
}

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
