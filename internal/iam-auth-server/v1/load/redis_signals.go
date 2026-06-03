package load

import (
	"github.com/redis/go-redis/v9"
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	"github.com/vmihailenco/msgpack/v5"
)

type NotificationCommand string

const (
	RedisPubSubChannel                      = "iam.cluster.notification"
	NoticePolicyChanged NotificationCommand = "PolicyChanged"
	NoticeSecretChanged NotificationCommand = "SecretChanged"
)

type Notification struct {
	Command NotificationCommand `mapstructure:"command"`
	Payload string              `mapstructure:"payload"`
}

func handleRedisEvent(v interface{}, handle func(command NotificationCommand), reloadedAsync func()) {
	message, ok := v.(*redis.Message)
	if !ok {
		return
	}

	notif := &Notification{}
	if err := msgpack.Unmarshal([]byte(message.Payload), notif); err != nil {
		log.Errorf("Error decoding notification: %v", err)
	}
	log.Debugf("Notification received: %v", notif)
	switch notif.Command {
	case NoticePolicyChanged:
		log.Info("Notice policy changed")
		policyChan <- reloadedAsync
	case NoticeSecretChanged:
		log.Info("Notice secret changed")
		secretChan <- reloadedAsync
	default:
		log.Warnf("Unknown notification command: %v", notif.Command)
	}

	//消息处理回调,目前尚未使用
	if handle != nil {
		handle(notif.Command)
	}
}
