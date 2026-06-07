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
	assert.Nil(t, mongoPump.client, "New instance should have nil client")
	assert.Nil(t, mongoPump.config, "New instance should have nil config")
}

func TestExtractDatabaseName(t *testing.T) {
	tests := []struct {
		name string
		uri  string
		want string
	}{
		{
			name: "database in path",
			uri:  "mongodb://root:pass@host:27017/admin?replicaSet=rs0",
			want: "admin",
		},
		{
			name: "authSource fallback when no db in path",
			uri:  "mongodb://root:pass@host:27017/?authSource=admin&replicaSet=rs0",
			want: "admin",
		},
		{
			name: "authSource fallback with no trailing slash db",
			uri:  "mongodb://host:27017,host2:27017/?authSource=testdb&replicaSet=rs0",
			want: "testdb",
		},
		{
			name: "path takes priority over authSource",
			uri:  "mongodb://host:27017/mydb?authSource=admin&replicaSet=rs0",
			want: "mydb",
		},
		{
			name: "no database and no authSource",
			uri:  "mongodb://host:27017",
			want: "",
		},
		{
			name: "no database and no authSource with trailing slash",
			uri:  "mongodb://host:27017/",
			want: "",
		},
		{
			name: "srv scheme with db in path",
			uri:  "mongodb+srv://host.example.com/mydb",
			want: "mydb",
		},
		{
			name: "srv scheme with authSource fallback",
			uri:  "mongodb+srv://host.example.com/?authSource=admin",
			want: "admin",
		},
		{
			name: "with username and password",
			uri:  "mongodb://user:pass@host:27017/dbname?authSource=admin",
			want: "dbname",
		},
		{
			name: "password with special chars",
			uri:  "mongodb://user:IAMpass123@host:27017,host2:27017/test?replicaSet=rs0",
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, extractDatabaseName(tt.uri))
		})
	}
}

func TestGetPumpByName_AvailablePumpsMap(t *testing.T) {
	// Reset availablePumps via a call to GetPumpByName which calls initProtoType
	_, _ = GetPumpByName("mongo")
	_, err := GetPumpByName("mongo")
	assert.NoError(t, err)
}
