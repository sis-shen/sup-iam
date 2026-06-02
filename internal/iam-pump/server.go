package iampump

import (
	"context"
	"fmt"
	"github.com/go-redsync/redsync/v4"
	redsyncredis "github.com/go-redsync/redsync/v4/redis/goredis/v9"
	goredis "github.com/redis/go-redis/v9"
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	"sync"

	"github.com/sis-shen/sup-iam/internal/iam-pump/analytics"
	"github.com/sis-shen/sup-iam/internal/iam-pump/options"
	"github.com/sis-shen/sup-iam/internal/iam-pump/pumps"
	"github.com/sis-shen/sup-iam/internal/iam-pump/store"
	"github.com/sis-shen/sup-iam/internal/iam-pump/store/redis"
	"github.com/vmihailenco/msgpack/v5"

	"time"
)

var pmps []pumps.PumpInterface

type pumpServer struct {
	options        *options.Options
	purgeDelay     time.Duration
	omitDetail     bool
	mutex          *redsync.Mutex
	analyticsStore storage.AnalyticStoreInterface
	mapPumpConfig  map[string]options.PumpOptions
}

// preparedGenericAPIServer is a private wrapper that enforces a call of PrepareRun() before Run can be invoked.

type preparedPumpServer struct {
	*pumpServer
}

func createPumpServer(o *options.Options) (*pumpServer, error) {
	// use the same redis database with authorization log history
	client := goredis.NewClient(&goredis.Options{
		Addr:     fmt.Sprintf("%s:%d", o.RedisOptions.Host, o.RedisOptions.Port),
		Username: o.RedisOptions.Username,
		Password: o.RedisOptions.Password,
	})

	rs := redsync.New(redsyncredis.NewPool(client))
	server := &pumpServer{
		options:        o,
		purgeDelay:     o.PurgeInterval,
		omitDetail:     o.OmitDetailRecoding,
		mutex:          rs.NewMutex("iam-pump", redsync.WithExpiry(10*time.Minute)),
		analyticsStore: &redis.RedisClusterStorageManager{},
		mapPumpConfig:  o.Pumps,
	}

	if err := server.analyticsStore.Init(o.RedisOptions); err != nil {
		return nil, err
	}
	return server, nil
}

func (p *pumpServer) PrepareRun() preparedPumpServer {
	p.initialize()
	return preparedPumpServer{p}
}

func (p preparedPumpServer) Run(stopChan <-chan struct{}) error {
	trigger := time.NewTicker(p.purgeDelay)
	defer trigger.Stop()
	for {
		select {
		case <-trigger.C:
			p.pumpServer.pump()
		case <-stopChan:
			log.Info("stopping pump server")
			return nil
		}
	}
}

func (p *pumpServer) pump() {
	//获取租期
	if err := p.mutex.Lock(); err != nil {
		log.Info("there is already an iam-pump instance running")
		return
	}

	defer func() {
		if _, err := p.mutex.Unlock(); err != nil {
			log.Errorf("failed to unlock: %v", err)
		}
	}()

	analyticsValues := p.analyticsStore.GetAndDeleteSet(storage.AnalyticsKeyName)
	keys := make([]interface{}, len(analyticsValues))
	for i, v := range analyticsValues {
		var decode analytics.AnalyticsRecode
		err := msgpack.Unmarshal([]byte(v.(string)), &decode)
		if err != nil {
			log.Errorf("failed to unmarshal analytics: %v", err)
			continue
		}
		if p.omitDetail {
			decode.Reason = ""
		}
		keys[i] = interface{}(decode)
	}

	writeToPumps(keys, p.purgeDelay)
}

func writeToPumps(keys []interface{}, purgeDelay time.Duration) {
	if pmps != nil {
		var wg sync.WaitGroup
		for _, pump := range pmps {
			go execPumpWriting(&wg, pump, &keys, purgeDelay)
		}
	}
}

func execPumpWriting(wg *sync.WaitGroup, pump pumps.PumpInterface, keys *[]interface{}, purgeDelay time.Duration) {
	timer := time.AfterFunc(purgeDelay, func() {
		if pump.GetTimeout() == 0 {
			log.Warnf(
				"Pump %s is taking more time than the value configured of purge_delay. You should try to set a timeout for this pump.",
				pump.GetName(),
			)
		} else if pump.GetTimeout() > pump.GetTimeout() {
			log.Warnf("Pump %s is taking more time than the value configured of purge_delay. You should try lowering the timeout configured for this pump.", pump.GetName())
		}

	})
	defer timer.Stop()
	defer wg.Done()

	ch := make(chan error, 1)
	var ctx context.Context
	var cancel context.CancelFunc

	if to := pump.GetTimeout(); to > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), to)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}

	defer cancel()
	go func(ch chan error, ctx context.Context, pump pumps.PumpInterface, keys *[]interface{}) {
		filteredKeys := filterData(pump, *keys)

		ch <- pump.WriteData(ctx, filteredKeys)
	}(ch, ctx, pump, keys)

	select {
	case err := <-ch:
		if err != nil {
			log.Errorf("failed to write to pump: %v", err)
		}
	case <-ctx.Done():
		switch ctx.Err() {
		case context.Canceled:
			log.Warnf("The writing to %s have got canceled.", pump.GetName())
		case context.DeadlineExceeded:
			log.Warnf("Timeout Writing to: %s", pump.GetName())
		}
	}
}

func filterData(pump pumps.PumpInterface, keys []interface{}) []interface{} {
	filters := pump.GetFilters()
	if !filters.HasFilter() && !pump.GetOmitDetailEnable() {
		return keys
	}

	filteredKeys := keys[:]
	newLength := 0
	for _, key := range filteredKeys {
		decoded, _ := key.(analytics.AnalyticsRecode)
		if pump.GetOmitDetailEnable() {
			decoded.Reason = ""
		}
		if filters.ShouldFilter(decoded.Username) {
			continue
		}
		filteredKeys[newLength] = interface{}(decoded)
		newLength++
	}
	filteredKeys = filteredKeys[:newLength]
	return filteredKeys
}

func (p *pumpServer) initialize() {
	pmps = make([]pumps.PumpInterface, 0)
	for key, pmp := range p.mapPumpConfig {
		pumpTypeName := pmp.Type
		if pumpTypeName == "" {
			pumpTypeName = key
		}

		thisPump, err := pumps.GetPumpByName(pumpTypeName)
		if err != nil {
			log.Errorf("Error getting pump %s: %s, now skipping", pumpTypeName, err)
		} else {
			err = thisPump.Init(pmp.Meta)
			if err != nil {
				log.Errorf("Error initializing pump %s: %s, now skipping", pumpTypeName, err)
				continue
			}
			log.Infof("Initialized pump %s", pumpTypeName)
			thisPump.SetFilters(pmp.Filters)
			thisPump.SetTimeout(pmp.Timeout)
			thisPump.SetOmitDetailEnable(pmp.OmitDetailRecoding)
		}
	}
}
