package mongo

import (
	"context"
	"github.com/imulab/go-scim/core"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

type persistenceProvider struct {
	resourceType	*core.ResourceType
	collections		[]*mongo.Collection
	maxTimePercent	int
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

func (p *persistenceProvider) ResourceType() *core.ResourceType {
	return p.resourceType
}

func (p *persistenceProvider) Total(ctx context.Context) (int64, error) {
	if len(p.collections) == 1 {
		return p.totalSingle(ctx, 0)
	}
	return p.totalAll(ctx)
}

func (p *persistenceProvider) totalSingle(ctx context.Context, collectionIndex int) (int64, error) {
	maxTime := p.getMaxTime(ctx)
	n, err := p.collections[collectionIndex].EstimatedDocumentCount(ctx, &options.EstimatedDocumentCountOptions{
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
			n int64
			err error
		}
	)

	var (
		total int64
		wg sync.WaitGroup
		resChan = make(chan result, len(p.collections))
	)

	wg.Add(len(p.collections))

	for i := range p.collections {
		go func(ctx context.Context, index int, resChan chan result) {
			defer wg.Done()

			n, err := p.totalSingle(ctx, index)
			select {
			case resChan <- result{n, err}:
				return
			default:
				return
			}
		}(ctx, i, resChan)
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
	panic("implement me")
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
