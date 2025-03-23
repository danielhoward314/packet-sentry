package dummy

import (
	"math/rand"
	"sync"
	"time"
)

// DummyService tests running a service as a background goroutine
type DummyService interface {
	Start()
	Stop()
}

// Service implements the DummyService interface
type Service struct {
	name            string
	shutdownChannel chan struct{}
	wg              sync.WaitGroup
	workInterval    time.Duration
}

type dummyData struct {
	OwningSvc      string `json:"owningSvc"`
	DummyData      string `json:"dummyData"`
	CollectionTime string `json:"collectionTime"`
}

func NewDummyService(workInterval time.Duration) DummyService {
	return &Service{
		name:            "dummyservice",
		shutdownChannel: make(chan struct{}),
		workInterval:    workInterval,
	}
}

func (svc *Service) Start() {
	svc.wg.Add(1)
	go func() {
		defer svc.wg.Done()
		for {
			select {
			case <-time.After(svc.workInterval):
				svc.doDummyServiceWork()
			case <-svc.shutdownChannel:
				return
			}
		}
	}()
}

// Stop closes the shutdown channel and blocks until wait group is Done.
// The Start goroutine reacts to closed shutdown channel by returning,
// which triggers the deferred `svc.wg.Done()` call and unblocks Stop.
func (svc *Service) Stop() {
	close(svc.shutdownChannel)
	svc.wg.Wait()
}

// randomString is just for demo-ing the dummyService doing work that has dynamic output
func randomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}
