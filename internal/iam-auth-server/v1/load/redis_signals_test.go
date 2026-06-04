package load

import (
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
	"sync"
	"testing"
	"time"
)

// drainPolicyChan reads from policyChan in a goroutine to prevent blocking
func drainPolicyChan(t *testing.T) (chan func(), *sync.WaitGroup) {
	t.Helper()
	received := make(chan func(), 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case fn := <-policyChan:
			received <- fn
		case <-time.After(3 * time.Second):
			t.Error("timeout draining policyChan")
		}
		close(received)
	}()
	return received, &wg
}

// drainSecretChan reads from secretChan in a goroutine to prevent blocking
func drainSecretChan(t *testing.T) (chan func(), *sync.WaitGroup) {
	t.Helper()
	received := make(chan func(), 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case fn := <-secretChan:
			received <- fn
		case <-time.After(3 * time.Second):
			t.Error("timeout draining secretChan")
		}
		close(received)
	}()
	return received, &wg
}

func TestHandleRedisEvent_PolicyChanged(t *testing.T) {
	notif := &Notification{
		Command: NoticePolicyChanged,
		Payload: "test-payload",
	}
	data, err := msgpack.Marshal(notif)
	assert.NoError(t, err)

	msg := &redis.Message{
		Payload: string(data),
	}

	var receivedCmd NotificationCommand
	handleFunc := func(cmd NotificationCommand) {
		receivedCmd = cmd
	}

	received, wg := drainPolicyChan(t)
	handleRedisEvent(msg, handleFunc, nil)
	wg.Wait()

	assert.Equal(t, NoticePolicyChanged, receivedCmd, "handle function should receive PolicyChanged command")

	// Verify callback was sent through the channel
	fn, ok := <-received
	assert.True(t, ok, "should receive callback from policyChan")
	assert.Nil(t, fn, "callback should be nil")
}

func TestHandleRedisEvent_SecretChanged(t *testing.T) {
	notif := &Notification{
		Command: NoticeSecretChanged,
		Payload: "test-payload",
	}
	data, err := msgpack.Marshal(notif)
	assert.NoError(t, err)

	msg := &redis.Message{
		Payload: string(data),
	}

	var receivedCmd NotificationCommand
	handleFunc := func(cmd NotificationCommand) {
		receivedCmd = cmd
	}

	received, wg := drainSecretChan(t)
	handleRedisEvent(msg, handleFunc, nil)
	wg.Wait()

	assert.Equal(t, NoticeSecretChanged, receivedCmd, "handle function should receive SecretChanged command")

	fn, ok := <-received
	assert.True(t, ok, "should receive callback from secretChan")
	assert.Nil(t, fn, "callback should be nil")
}

func TestHandleRedisEvent_UnknownCommand(t *testing.T) {
	notif := &Notification{
		Command: NotificationCommand("UnknownCommand"),
		Payload: "",
	}
	data, err := msgpack.Marshal(notif)
	assert.NoError(t, err)

	msg := &redis.Message{
		Payload: string(data),
	}

	var receivedCmd NotificationCommand
	handleFunc := func(cmd NotificationCommand) {
		receivedCmd = cmd
	}

	// Unknown command should not send to any channel, so no drain needed
	handleRedisEvent(msg, handleFunc, nil)

	assert.Equal(t, NotificationCommand("UnknownCommand"), receivedCmd, "handle function should forward unknown commands too")
}

func TestHandleRedisEvent_InvalidMessageType(t *testing.T) {
	// Passing a non-redis.Message type should silently return
	handleRedisEvent("not a redis message", nil, nil)
	// No assertions needed - just verify it doesn't panic
}

func TestHandleRedisEvent_NilMessagePayload(t *testing.T) {
	msg := &redis.Message{
		Payload: "",
	}

	// Empty payload should not panic
	handleRedisEvent(msg, nil, nil)
}

func TestHandleRedisEvent_InvalidMsgpack(t *testing.T) {
	msg := &redis.Message{
		Payload: "not valid msgpack data",
	}

	handleRedisEvent(msg, nil, nil)
	// Should not panic, just log error
}

func TestHandleRedisEvent_NilHandle(t *testing.T) {
	notif := &Notification{
		Command: NoticePolicyChanged,
	}
	data, err := msgpack.Marshal(notif)
	assert.NoError(t, err)

	msg := &redis.Message{
		Payload: string(data),
	}

	received, wg := drainPolicyChan(t)
	// Nil handle function should not panic
	handleRedisEvent(msg, nil, nil)
	wg.Wait()

	_, ok := <-received
	assert.True(t, ok, "should receive callback from policyChan")
}

func TestHandleRedisEvent_PolicyChanWithCallback(t *testing.T) {
	notif := &Notification{
		Command: NoticePolicyChanged,
		Payload: "policy-id-1",
	}
	data, err := msgpack.Marshal(notif)
	assert.NoError(t, err)

	msg := &redis.Message{
		Payload: string(data),
	}

	callbackCalled := make(chan struct{}, 1)
	received, wg := drainPolicyChan(t)

	handleRedisEvent(msg, nil, func() {
		callbackCalled <- struct{}{}
	})

	wg.Wait()

	// The callback was sent through policyChan, so drainPolicyChan received it
	fn, ok := <-received
	assert.True(t, ok, "should receive callback from policyChan")
	assert.NotNil(t, fn, "callback should not be nil")

	// Call the received function to verify it's the right one
	fn()
	select {
	case <-callbackCalled:
		// Success - the callback was invoked
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for callback invocation")
	}
}

func TestHandleRedisEvent_SecretChanWithCallback(t *testing.T) {
	notif := &Notification{
		Command: NoticeSecretChanged,
		Payload: "secret-id-1",
	}
	data, err := msgpack.Marshal(notif)
	assert.NoError(t, err)

	msg := &redis.Message{
		Payload: string(data),
	}

	callbackCalled := make(chan struct{}, 1)
	received, wg := drainSecretChan(t)

	handleRedisEvent(msg, nil, func() {
		callbackCalled <- struct{}{}
	})

	wg.Wait()

	fn, ok := <-received
	assert.True(t, ok, "should receive callback from secretChan")
	assert.NotNil(t, fn, "callback should not be nil")

	fn()
	select {
	case <-callbackCalled:
		// Success - the callback was invoked
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for callback invocation")
	}
}
