/**
 * Copyright 2015 @ to2.net.
 * name : testing
 * author : jarryliu
 * date : 2016-06-15 08:31
 * description :
 * history :
 */
package ti

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/ixre/gof"
	"github.com/ixre/gof/db"
	"github.com/ixre/gof/db/orm"
	"github.com/ixre/gof/log"
	"github.com/ixre/gof/storage"
	"go.etcd.io/etcd/clientv3"
	"go2o/core"
	"go2o/core/repos"
	"go2o/core/service/impl"
	"time"
)

var (
	app     *testingApp
	Factory *repos.RepoFactory
)
var (
	REDIS_DB = "1"
)

func GetApp() gof.App {
	return gof.CurrentApp
}

var _ gof.App = new(testingApp)

// application context
// implement of web.Application
type testingApp struct {
	Loaded        bool
	_confFilePath string
	_config       *gof.Config
	_redis        *redis.Pool
	_dbConnector  db.Connector
	_debugMode    bool
	_template     *gof.Template
	_logger       log.ILogger
	_storage      storage.Interface
	_registry     *gof.Registry
}

func newMainApp(confPath string) *testingApp {
	return &testingApp{
		_confFilePath: confPath,
	}
}

func (t *testingApp) Registry() *gof.Registry {
	if t._registry == nil {
		t._registry, _ = gof.NewRegistry("../conf", ":")
	}
	return t._registry
}

func (t *testingApp) Db() db.Connector {
	if t._dbConnector == nil {
		connStr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
			t._config.GetString(core.DbUsr),
			t._config.GetString(core.DbPwd),
			t._config.GetString(core.DbServer),
			t._config.GetString(core.DbPort),
			t._config.GetString(core.DbName))
		conn, _ := db.NewConnector("postgresql", connStr, t.Log(), t._debugMode)
		conn.SetMaxIdleConns(10000)
		conn.SetMaxIdleConns(5000)
		conn.SetConnMaxLifetime(time.Second * 10)
		t._dbConnector = conn
	}
	return t._dbConnector
}

func (t *testingApp) Storage() storage.Interface {
	if t._storage == nil {
		t._storage = storage.NewRedisStorage(t.Redis())
	}
	return t._storage
}

func (t *testingApp) Config() *gof.Config {
	if t._config == nil {
		if t._confFilePath == "" {
			t._config = gof.NewConfig()
		} else {
			if cfg, err := gof.LoadConfig(t._confFilePath); err == nil {
				t._config = cfg
			} else {
				log.Fatalln(err)
			}
		}
	}
	return t._config
}

func (t *testingApp) Source() interface{} {
	return t
}

func (t *testingApp) Debug() bool {
	return t._debugMode
}

func (t *testingApp) Log() log.ILogger {
	if t._logger == nil {
		var flag = 0
		if t._debugMode {
			flag = log.LOpen | log.LESource | log.LStdFlags
		}
		t._logger = log.NewLogger(nil, " O2O", flag)
	}
	return t._logger
}

func (t *testingApp) Redis() *redis.Pool {
	if t._redis == nil {
		t._redis = core.CreateRedisPool(t.Config())
	}
	return t._redis
}

func (t *testingApp) Init(debug, trace bool) bool {
	t._debugMode = debug
	t.Loaded = true
	return true
}

func init() {
	// 默认的ETCD端点
	etcdEndPoints := []string{"http://127.0.0.1:2379"}
	cfg := clientv3.Config{
		Endpoints:   etcdEndPoints,
		DialTimeout: 5 * time.Second,
	}
	app := core.NewApp("../app_dev.conf", &cfg)
	gof.CurrentApp = app
	core.Init(app, false, false)
	conn := app.Db()
	sto := app.Storage()
	o := orm.NewOrm(conn.Driver(), conn.Raw())
	Factory = (&repos.RepoFactory{}).Init(o, sto)
	impl.InitTestService(app, conn, o, sto)
}
