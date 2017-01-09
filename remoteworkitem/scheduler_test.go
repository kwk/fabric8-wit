package remoteworkitem

import (
	"fmt"
	"os"
	"testing"

	"github.com/almighty/almighty-core/configuration"
	"github.com/almighty/almighty-core/resource"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	var err error

	if err = configuration.Setup(""); err != nil {
		panic(fmt.Errorf("Failed to setup the configuration: %s", err.Error()))
	}

	if _, c := os.LookupEnv(resource.Database); c {
		db, err = gorm.Open("postgres", configuration.GetPostgresConfigString())
		if err != nil {
			panic("Failed to connect database: " + err.Error())
		}
		defer db.Close()
	}
	os.Exit(m.Run())
}

func TestNewScheduler(t *testing.T) {
	resource.Require(t, resource.Database)

	s := NewScheduler(db)
	if s.db != db {
		t.Error("DB not set as an attribute")
	}
	s.Stop()
}

func TestLookupProvider(t *testing.T) {
	resource.Require(t, resource.Database)
	ts1 := trackerSchedule{TrackerType: ProviderGithub}
	tp1 := lookupProvider(ts1)
	assert.NotNil(t, tp1, "nil provider")
	ts2 := trackerSchedule{TrackerType: ProviderJira}
	tp2 := lookupProvider(ts2)
	assert.NotNil(t, tp2, "nil provider")
	ts3 := trackerSchedule{TrackerType: "unknown"}
	tp3 := lookupProvider(ts3)
	assert.Nil(t, tp3, "non-nil provider")
}
