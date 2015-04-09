package ledis

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/astaxie/beego"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"

	"github.com/containerops/wharf/models"
	"github.com/containerops/wharf/utils"
)

const (
	driver = "ledis"
)

const (
	GLOBAL_USER_INDEX         = "GLOBAL_USER_INDEX"
	GLOBAL_REPOSITORY_INDEX   = "GLOBAL_REPOSITORY_INDEX"
	GLOBAL_ORGANIZATION_INDEX = "GLOBAL_ORGANIZATION_INDEX"
	GLOBAL_TEAM_INDEX         = "GLOBAL_TEAM_INDEX"
	GLOBAL_IMAGE_INDEX        = "GLOBAL_IMAGE_INDEX"
	GLOBAL_TARSUM_INDEX       = "GLOBAL_TARSUM_INDEX"
	GLOBAL_TAG_INDEX          = "GLOBAL_TAG_INDEX"
	GLOBAL_COMPOSE_INDEX      = "GLOBAL_COMPOSE_INDEX"
	GLOBAL_ADMIN_INDEX        = "GLOBAL_ADMIN_INDEX"
	GLOBAL_PRIVILEGE_INDEX    = "GLOBAL_PRIVILEGE_INDEX"
	GLOBAL_LOG_INDEX          = "GLOBAL_LOG_INDEX"
)

var (
	ledisOnce sync.Once
	nowLedis  *ledis.Ledis
	LedisDB   *ledis.DB
)

type ledisDriverFactory struct{}

func (l *ledisDriverFactory) Create(parameters map[string]string) (models.DatabaseDriver, error) {
	return initLedis(parameters), nil
}

type LedisDriver struct{}

func initLedis(parameters map[string]string) *LedisDriver {

}
