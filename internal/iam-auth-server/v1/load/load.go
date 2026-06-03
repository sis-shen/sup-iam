package load

import (
	"context"
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/storage/redis"
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	"sync"
	"time"
)

type Loader interface {
	ReloadSecrets() error
	ReloadPolicies() error
	Reload() error
}

var (
	secretMutex sync.Mutex
	secretChan  = make(chan func())
	secretQueue []func()
	policyMutex sync.Mutex
	policyChan  = make(chan func())
	policyQueue []func()
)

type LoadManager struct {
	ctx        context.Context
	lock       *sync.RWMutex
	store      Loader
	reloadTick time.Duration
	wg         *sync.WaitGroup
}

func NewLoadManager(ctx context.Context, store Loader, reloadTick time.Duration) *LoadManager {
	return &LoadManager{
		ctx:        ctx,
		lock:       new(sync.RWMutex),
		store:      store,
		reloadTick: reloadTick,
		wg:         new(sync.WaitGroup),
	}
}

func (l *LoadManager) Start() error {
	l.wg.Add(3)
	go l.pubSubLoop()
	go l.reloadQueueLoop()
	go l.reloadLoop()

	return l.store.Reload()
}

func (l *LoadManager) ShutDown() {
	l.ctx.Done()
	l.wg.Wait()
}

func (l *LoadManager) pubSubLoop() {
	defer l.wg.Done()
	redisCli := redis.GetRedisClusterSingleton()
	err := redisCli.EnsureConnection()
	if err != nil {
		log.Errorf("Error connecting to redis: %v", err)
		return
	}
	for {
		err := redisCli.RegisterPubSubHandler(RedisPubSubChannel, func(v interface{}) {
			handleRedisEvent(v, nil, nil)
		})
		// 先判断是不是要关停了
		select {
		case <-l.ctx.Done():
			return
		default:
		}
		if err != nil {
			log.Errorf("Error subscribing to Redis: %v\n now try reconnecting...", err)
			time.Sleep(10 * time.Second)
			_ = redisCli.EnsureConnection()
		}
	}
}

func secretShouldReload() ([]func(), bool) {
	secretMutex.Lock()
	defer secretMutex.Unlock()

	if len(secretQueue) == 0 {
		return nil, false
	}
	res := secretQueue[:]
	secretQueue = secretQueue[:0]
	return res, true
}

func policyShouldReload() ([]func(), bool) {
	policyMutex.Lock()
	defer policyMutex.Unlock()

	if len(policyQueue) == 0 {
		return nil, false
	}
	res := policyQueue[:]
	policyQueue = policyQueue[:0]
	return res, true
}

// move messages from chan to queue
func (l *LoadManager) reloadQueueLoop() {
	defer l.wg.Done()
	for {
		select {
		case <-l.ctx.Done():
			return
		case fn := <-policyChan:
			policyMutex.Lock()
			policyQueue = append(policyQueue, fn)
			policyMutex.Unlock()
		case fn := <-secretChan:
			secretMutex.Lock()
			secretQueue = append(secretQueue, fn)
			secretMutex.Unlock()
		}
	}
}

func (l *LoadManager) reloadLoop() {
	defer l.wg.Done()
	ticker := time.NewTicker(l.reloadTick)
	for {
		select {
		case <-l.ctx.Done():
			return
		case <-ticker.C:
			fns, should := secretShouldReload()
			if should {
				err := l.store.ReloadSecrets()
				nowTime := time.Now()
				if err != nil {
					log.Errorf("Error reloading secrets: %v", err)
				}
				if len(fns) > 0 && fns[0] != nil {
					fns[0]()
				}
				log.Infof("reloaded secrets in %v", time.Since(nowTime))
			}

			fns, should = policyShouldReload()
			if should {
				nowTime := time.Now()
				err := l.store.ReloadPolicies()
				if err != nil {
					log.Errorf("Error reloading policies: %v", err)
				}
				if len(fns) > 0 && fns[0] != nil {
					fns[0]()
				}
				log.Infof("reloaded policies in %v", time.Since(nowTime))
			}
		}
	}
}
