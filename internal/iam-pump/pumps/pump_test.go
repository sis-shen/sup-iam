package pumps

import (
	"github.com/sis-shen/sup-iam/internal/iam-pump/analytics"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetPumpByName_Mongo(t *testing.T) {
	pump, err := GetPumpByName("mongo")
	assert.NoError(t, err)
	assert.NotNil(t, pump)
	assert.Equal(t, "MongoDB Pump", pump.GetName())
}

func TestGetPumpByName_Unknown(t *testing.T) {
	pump, err := GetPumpByName("unknown-pump")
	assert.Error(t, err)
	assert.Nil(t, pump)
	assert.Contains(t, err.Error(), "Not found")
}

func TestGetPumpByName_EmptyName(t *testing.T) {
	pump, err := GetPumpByName("")
	assert.Error(t, err)
	assert.Nil(t, pump)
	assert.Contains(t, err.Error(), "Not found")
}

func TestNewPumpReturnsNewInstance(t *testing.T) {
	pump, err := GetPumpByName("mongo")
	assert.NoError(t, err)

	pump2, err := GetPumpByName("mongo")
	assert.NoError(t, err)

	assert.NotSame(t, pump, pump2, "Each call to GetPumpByName should return a new instance")
}

func TestMongoPump_ImplementsInterface(t *testing.T) {
	var _ PumpInterface = (*MongoPump)(nil)
}

func TestCommonPump_GetSetFilters(t *testing.T) {
	p := &CommonPump{}
	assert.Nil(t, p.GetFilters())

	filters := &analytics.AnalyticFilters{
		Usernames:        []string{"user1"},
		SkippedUsernames: []string{"admin"},
	}
	p.SetFilters(filters)
	assert.Equal(t, filters, p.GetFilters())
}

func TestCommonPump_GetSetTimeout(t *testing.T) {
	p := &CommonPump{}
	assert.Equal(t, time.Duration(0), p.GetTimeout())

	p.SetTimeout(30 * time.Second)
	assert.Equal(t, 30*time.Second, p.GetTimeout())
}

func TestCommonPump_GetSetOmitDetail(t *testing.T) {
	p := &CommonPump{}
	assert.Equal(t, false, p.GetOmitDetailEnable())

	p.SetOmitDetailEnable(true)
	assert.Equal(t, true, p.GetOmitDetailEnable())

	p.SetOmitDetailEnable(false)
	assert.Equal(t, false, p.GetOmitDetailEnable())
}

func TestMongoPump_New(t *testing.T) {
	m := &MongoPump{}
	pump := m.New()
	assert.NotNil(t, pump)

	mongoPump, ok := pump.(*MongoPump)
	assert.True(t, ok, "New() should return a *MongoPump")
	assert.Nil(t, mongoPump.session, "New instance should have nil session")
	assert.Nil(t, mongoPump.config, "New instance should have nil config")
}

func TestGetPumpByName_AvailablePumpsMap(t *testing.T) {
	// Reset availablePumps via a call to GetPumpByName which calls initProtoType
	_, _ = GetPumpByName("mongo")
	_, err := GetPumpByName("mongo")
	assert.NoError(t, err)
}
