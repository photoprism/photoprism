// Copyright 2015 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package tidb

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/pingcap/errors"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/parser/terror"
	pd "github.com/pingcap/pd/client"
	pumpcli "github.com/pingcap/tidb-tools/tidb-binlog/pump_client"
	"github.com/pingcap/tidb/config"
	"github.com/pingcap/tidb/ddl"
	"github.com/pingcap/tidb/domain"
	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/metrics"
	plannercore "github.com/pingcap/tidb/planner/core"
	"github.com/pingcap/tidb/privilege/privileges"
	"github.com/pingcap/tidb/server"
	"github.com/pingcap/tidb/session"
	"github.com/pingcap/tidb/sessionctx/binloginfo"
	"github.com/pingcap/tidb/sessionctx/variable"
	"github.com/pingcap/tidb/statistics"
	"github.com/pingcap/tidb/store/mockstore"
	"github.com/pingcap/tidb/store/tikv"
	"github.com/pingcap/tidb/store/tikv/gcworker"
	"github.com/pingcap/tidb/util/logutil"
	"github.com/pingcap/tidb/util/printer"
	xserver "github.com/pingcap/tidb/x-server"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var log = event.Log

var (
	cfg      *config.Config
	storage  kv.Storage
	dom      *domain.Domain
	svr      *server.Server
	xsvr     *xserver.Server
	graceful bool
)

// Start the TiDB server using the configuration provided
func Start(ctx context.Context, path string, port uint, host string, debug bool) {
	if err := logutil.SetLevel("fatal"); err != nil {
		log.Error(err)
	}

	registerStores()
	registerMetrics()
	loadConfig()

	cfg.Log.Level = log.GetLevel().String()

	// cfg.Security.SkipGrantTable = true
	if debug {
		cfg.Log.Level = "error"
		host = "0.0.0.0"
	}

	cfg.Path = path
	cfg.Store = "mocktikv"

	if host == "" {
		host = "localhost"
	}

	cfg.Host = host
	cfg.Port = port
	cfg.Status.ReportStatus = false

	validateConfig()
	setGlobalVars()
	setupTracing()

	if debug {
		printInfo()
	}

	setupBinlogClient()
	// setupMetrics()
	createStoreAndDomain()
	createServer()
	go runServer()

	<-ctx.Done()
	serverShutdown(true)
	cleanup()
	log.Info("tidb server shutdown complete")
}

func registerStores() {
	err := session.RegisterStore("tikv", tikv.Driver{})
	terror.MustNil(err)
	tikv.NewGCHandlerFunc = gcworker.NewGCWorker
	err = session.RegisterStore("mocktikv", mockstore.MockDriver{})
	terror.MustNil(err)
}

func registerMetrics() {
	metrics.RegisterMetrics()
}

func createStoreAndDomain() {
	fullPath := fmt.Sprintf("%s://%s", cfg.Store, cfg.Path)
	var err error
	storage, err = session.NewStore(fullPath)
	terror.MustNil(err)
	// Bootstrap a session to load information schema.
	dom, err = session.BootstrapSession(storage)
	terror.MustNil(err)
}

func setupBinlogClient() {
	if !cfg.Binlog.Enable {
		return
	}

	if cfg.Binlog.IgnoreError {
		binloginfo.SetIgnoreError(true)
	}

	client, err := pumpcli.NewPumpsClient(cfg.Path, parseDuration(cfg.Binlog.WriteTimeout), pd.SecurityOption{
		CAPath:   cfg.Security.ClusterSSLCA,
		CertPath: cfg.Security.ClusterSSLCert,
		KeyPath:  cfg.Security.ClusterSSLKey,
	})
	terror.MustNil(err)

	binloginfo.SetPumpsClient(client)
	log.Infof("create pumps client success, ignore binlog error %v", cfg.Binlog.IgnoreError)
}

// Prometheus push.
const zeroDuration = time.Duration(0)

// pushMetric pushes metrics in background.
func pushMetric(addr string, interval time.Duration) {
	if interval == zeroDuration || len(addr) == 0 {
		log.Info("disable Prometheus push client")
		return
	}
	log.Infof("start Prometheus push client with server addr %s and interval %s", addr, interval)
	go prometheusPushClient(addr, interval)
}

// prometheusPushClient pushes metrics to Prometheus Pushgateway.
func prometheusPushClient(addr string, interval time.Duration) {
	// TODO: TiDB do not have uniq name, so we use host+port to compose a name.
	job := "tidb"
	for {
		err := push.AddFromGatherer(
			job,
			map[string]string{"instance": instanceName()},
			addr,
			prometheus.DefaultGatherer,
		)
		if err != nil {
			log.Errorf("could not push metrics to Prometheus Pushgateway: %v", err)
		}
		time.Sleep(interval)
	}
}

func instanceName() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return fmt.Sprintf("%s_%d", hostname, cfg.Port)
}

// parseDuration parses lease argument string.
func parseDuration(lease string) time.Duration {
	dur, err := time.ParseDuration(lease)
	if err != nil {
		dur, err = time.ParseDuration(lease + "s")
	}
	if err != nil || dur < 0 {
		log.Fatalf("invalid lease duration %s", lease)
	}
	return dur
}

func hasRootPrivilege() bool {
	return os.Geteuid() == 0
}

func flagBoolean(name string, defaultVal bool, usage string) *bool {
	if defaultVal == false {
		// Fix #4125, golang do not print default false value in usage, so we append it.
		usage = fmt.Sprintf("%s (default false)", usage)
		return flag.Bool(name, defaultVal, usage)
	}
	return flag.Bool(name, defaultVal, usage)
}

func loadConfig() {
	cfg = config.GetGlobalConfig()
}

func validateConfig() {
	if cfg.Security.SkipGrantTable && !hasRootPrivilege() {
		log.Error("TiDB run with skip-grant-table need root privilege.")
		os.Exit(-1)
	}
	if _, ok := config.ValidStorage[cfg.Store]; !ok {
		nameList := make([]string, 0, len(config.ValidStorage))
		for k, v := range config.ValidStorage {
			if v {
				nameList = append(nameList, k)
			}
		}
		log.Errorf("\"store\" should be in [%s] only", strings.Join(nameList, ", "))
		os.Exit(-1)
	}
	if cfg.Store == "mocktikv" && cfg.RunDDL == false {
		log.Errorf("can't disable DDL on mocktikv")
		os.Exit(-1)
	}
	if cfg.Log.File.MaxSize > config.MaxLogFileSize {
		log.Errorf("log max-size should not be larger than %d MB", config.MaxLogFileSize)
		os.Exit(-1)
	}
	if cfg.XProtocol.XServer {
		log.Error("X Server is not available")
		os.Exit(-1)
	}
	cfg.OOMAction = strings.ToLower(cfg.OOMAction)

	// lower_case_table_names is allowed to be 0, 1, 2
	if cfg.LowerCaseTableNames < 0 || cfg.LowerCaseTableNames > 2 {
		log.Errorf("lower-case-table-names should be 0 or 1 or 2.")
		os.Exit(-1)
	}
}

func setGlobalVars() {
	ddlLeaseDuration := parseDuration(cfg.Lease)
	session.SetSchemaLease(ddlLeaseDuration)
	runtime.GOMAXPROCS(int(cfg.Performance.MaxProcs))
	statsLeaseDuration := parseDuration(cfg.Performance.StatsLease)
	session.SetStatsLease(statsLeaseDuration)
	domain.RunAutoAnalyze = cfg.Performance.RunAutoAnalyze
	statistics.FeedbackProbability = cfg.Performance.FeedbackProbability
	statistics.MaxQueryFeedbackCount = int(cfg.Performance.QueryFeedbackLimit)
	statistics.RatioOfPseudoEstimate = cfg.Performance.PseudoEstimateRatio
	ddl.RunWorker = cfg.RunDDL
	ddl.EnableSplitTableRegion = cfg.SplitTable
	plannercore.AllowCartesianProduct = cfg.Performance.CrossJoin
	privileges.SkipWithGrant = cfg.Security.SkipGrantTable
	variable.ForcePriority = int32(mysql.Str2Priority(cfg.Performance.ForcePriority))

	variable.SysVars[variable.TIDBMemQuotaQuery].Value = strconv.FormatInt(cfg.MemQuotaQuery, 10)
	variable.SysVars["lower_case_table_names"].Value = strconv.Itoa(cfg.LowerCaseTableNames)

	plannercore.SetPreparedPlanCache(cfg.PreparedPlanCache.Enabled)
	if plannercore.PreparedPlanCacheEnabled() {
		plannercore.PreparedPlanCacheCapacity = cfg.PreparedPlanCache.Capacity
	}

	if cfg.TiKVClient.GrpcConnectionCount > 0 {
		tikv.MaxConnectionCount = cfg.TiKVClient.GrpcConnectionCount
	}
	tikv.GrpcKeepAliveTime = time.Duration(cfg.TiKVClient.GrpcKeepAliveTime) * time.Second
	tikv.GrpcKeepAliveTimeout = time.Duration(cfg.TiKVClient.GrpcKeepAliveTimeout) * time.Second

	tikv.CommitMaxBackoff = int(parseDuration(cfg.TiKVClient.CommitTimeout).Seconds() * 1000)
}

func printInfo() {
	// Make sure the TiDB info is always printed.
	printer.PrintTiDBInfo()
}

func createServer() {
	var driver server.IDriver
	driver = server.NewTiDBDriver(storage)
	var err error
	svr, err = server.NewServer(cfg, driver)
	// Both domain and storage have started, so we have to clean them before exiting.
	terror.MustNil(err, closeDomainAndStorage)
	if cfg.XProtocol.XServer {
		xcfg := &xserver.Config{
			Addr:       fmt.Sprintf("%s:%d", cfg.XProtocol.XHost, cfg.XProtocol.XPort),
			Socket:     cfg.XProtocol.XSocket,
			TokenLimit: cfg.TokenLimit,
		}
		xsvr, err = xserver.NewServer(xcfg)
		terror.MustNil(err, closeDomainAndStorage)
	}
}

func serverShutdown(isgraceful bool) {
	if isgraceful {
		log.Info("graceful database shutdown")
		graceful = true
	} else {
		log.Info("database shutdown")
	}

	if xsvr != nil {
		xsvr.Close() // Should close xserver before server.
	}

	svr.Close()

	log.Info("database server closed")
}

func setupTracing() {
	tracingCfg := cfg.OpenTracing.ToTracingConfig()
	tracer, _, err := tracingCfg.New("TiDB")
	if err != nil {
		log.Fatal("cannot initialize Jaeger Tracer", err)
	}
	opentracing.SetGlobalTracer(tracer)
}

func runServer() {
	err := svr.Run()
	if err != nil {
		log.Errorf("Server failed to run: %v", err)
	}
}

func closeDomainAndStorage() {
	dom.Close()
	if err := storage.Close(); err != nil {
		log.Error(errors.Trace(err))
	} else {
		log.Info("database storage closed")
	}
}

func cleanup() {
	if graceful {
		svr.GracefulDown()
	} else {
		svr.KillAllConnections()
	}
	closeDomainAndStorage()
}
