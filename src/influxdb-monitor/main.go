package main

import (
	"configuration"
	"statistic"

	"flag"
	"fmt"
	"os"
	"util"

	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	//	"runtime"
	//	"syscall"

	"github.com/cloudfoundry/gosigar"
	"github.com/influxdb/influxdb-go"
	"time"
)

var config *configuration.Configuration

func getInfluxDBMetric(host, port, user, password string) (*statistic.Statistic, error) {
	body, err := http.Get(fmt.Sprintf("http://%s:%s/statistic?u=%s&p=%s", host, port, user, password))
	if err != nil {
		fmt.Printf("Failed %+v\n", err)
		return nil, err
	}
	defer body.Body.Close()

	data, err := ioutil.ReadAll(body.Body)
	if err != nil {
		return nil, err
	}

	stat := statistic.Statistic{}
	err = json.Unmarshal(data, &stat)
	if err != nil {
		fmt.Printf("Unmarshal Error: %s\n", err)
		return nil, err
	}

	return &stat, nil
}

func getClient(user, pass, database string) *influxdb.Client {
	client, err := influxdb.NewClient(&influxdb.ClientConfig{
		Username: user,
		Password: pass,
		Database: database,
	})

	if err != nil {
		fmt.Printf("Client Error: %+v\n", err)
	}

	return client
}

func getLeveldbOpened(pid string) int {
	lsof := exec.Command("lsof", "-p", pid)
	grep := exec.Command("grep", "shard_db")
	wc := exec.Command("wc", "-l")

	var b2 bytes.Buffer
	r, w := io.Pipe()
	r2, w2 := io.Pipe()

	lsof.Stdout = w
	grep.Stdin = r
	grep.Stdout = w2
	wc.Stdin = r2

	wc.Stdout = &b2

	lsof.Start()
	grep.Start()
	wc.Start()

	lsof.Wait()
	w.Close()

	grep.Wait()
	w2.Close()

	wc.Wait()

	cnt, err := strconv.Atoi(strings.Trim(b2.String(), " \t\n\r"))
	if err != nil {
		fmt.Printf("Atoi Error: %s\n", err)
		return -1
	}

	return cnt
}

func getHttpStatusCount(status int, metric *statistic.Statistic) uint64 {
	for _, s := range metric.Api.Http.Response.Status {
		if s.Code == status {
			return uint64(s.Count)
		}
	}

	return uint64(0)
}

func loop() {
	client := getClient(config.Output.UserName, config.Output.Password, config.Output.Database)

	fmt.Printf("Client: %+v\n", client)
	last := statistic.Statistic{}
	lastCpu := sigar.Cpu{}
	http_response_200 := uint64(0)
	http_response_400 := uint64(0)
	http_response_401 := uint64(0)
	http_response_409 := uint64(0)
	http_response_500 := uint64(0)
	leveldb_opend := 0
	container := []*influxdb.Series{}
	cpu := sigar.Cpu{}

	for {
		metric, err := getInfluxDBMetric(config.Collect.Host, config.Collect.Port, config.Collect.UserName, config.Collect.Password)
		if err != nil {
			goto sleep
		}
		leveldb_opend = getLeveldbOpened(strconv.Itoa(metric.Pid))
		fmt.Printf("Last: %+v\n", last)
		fmt.Printf("Metric: %+v\n", metric)

		if last.Pid != 0 {
			// calculate difference
			http_response_200 = getHttpStatusCount(200, metric) - getHttpStatusCount(200, &last)
			http_response_400 = getHttpStatusCount(400, metric) - getHttpStatusCount(400, &last)
			http_response_401 = getHttpStatusCount(401, metric) - getHttpStatusCount(401, &last)
			http_response_409 = getHttpStatusCount(409, metric) - getHttpStatusCount(409, &last)
			http_response_500 = getHttpStatusCount(500, metric) - getHttpStatusCount(500, &last)

			container = append(container, &influxdb.Series{
				Name: fmt.Sprintf("influxdb.%s.api.http", metric.Name),
				Columns: []string{
					"bytes_written",
					"bytes_read",
					"current_connections",
					"connections",
					"response_min",
					"response_max",
					"response_avg",
					"response_status_200",
					"response_status_400",
					"response_status_401",
					"response_status_409",
					"response_status_500",
				},
				Points: [][]interface{}{
					[]interface{}{
						(metric.Api.Http.BytesWritten - last.Api.Http.BytesWritten),
						(metric.Api.Http.BytesRead - last.Api.Http.BytesRead),
						metric.Api.Http.Connections - last.Api.Http.Connections,
						metric.Api.Http.CurrentConnections,
						metric.Api.Http.Response.Min,
						metric.Api.Http.Response.Max,
						metric.Api.Http.Response.Avg,
						http_response_200,
						http_response_400,
						http_response_401,
						http_response_409,
						http_response_500,
					},
				},
			})
			container = append(container, &influxdb.Series{
				Name: fmt.Sprintf("influxdb.%s.wal", metric.Name),
				Columns: []string{
					"opening",
					"append_entry",
					"commit_entry",
					"bookmark_entry",
				},
				Points: [][]interface{}{
					[]interface{}{
						metric.Wal.Opening,
						metric.Wal.AppendEntry - last.Wal.AppendEntry,
						metric.Wal.CommitEntry - last.Wal.CommitEntry,
						metric.Wal.BookmarkEntry - last.Wal.BookmarkEntry,
					},
				},
			})
			container = append(container, &influxdb.Series{
				Name: fmt.Sprintf("influxdb.%s.shard", metric.Name),
				Columns: []string{
					"opening",
					"delete",
				},
				Points: [][]interface{}{
					[]interface{}{
						metric.Shard.Opening,
						metric.Shard.Delete - last.Shard.Delete,
					},
				},
			})

			container = append(container, &influxdb.Series{
				Name: fmt.Sprintf("influxdb.%s.coordinator", metric.Name),
				Columns: []string{
					"cmd_query",
					"cmd_select",
					"cmd_write_series",
					"cmd_delete",
					"cmd_drop",
					"cmd_list_series",
				},
				Points: [][]interface{}{
					[]interface{}{
						metric.Coordinator.CmdQuery - last.Coordinator.CmdQuery,
						metric.Coordinator.CmdSelect - last.Coordinator.CmdSelect,
						metric.Coordinator.CmdWriteSeries - last.Coordinator.CmdWriteSeries,
						metric.Coordinator.CmdDelete - last.Coordinator.CmdDelete,
						metric.Coordinator.CmdDrop - last.Coordinator.CmdDrop,
						metric.Coordinator.CmdListSeries - last.Coordinator.CmdListSeries,
					},
				},
			})

			container = append(container, &influxdb.Series{
				Name: fmt.Sprintf("influxdb.%s.net", metric.Name),
				Columns: []string{
					"current_connections",
					"connections",
					"bytes_written",
					"bytes_read",
				},
				Points: [][]interface{}{
					[]interface{}{
						metric.Net.CurrentConnections,
						metric.Net.Connections,
						metric.Net.BytesWritten - last.Net.BytesWritten,
						metric.Net.BytesRead - last.Net.BytesRead,
					},
				},
			})

			container = append(container, &influxdb.Series{
				Name: fmt.Sprintf("influxdb.%s.leveldb", metric.Name),
				Columns: []string{
					"points_read",
					"points_write",
					"points_delete",
					"opening",
					"writetime_min",
					"writetime_max",
					"writetime_avg",
					"bytes_written",
				},
				Points: [][]interface{}{
					[]interface{}{
						metric.LevelDB.PointsRead - last.LevelDB.PointsRead,
						metric.LevelDB.PointsWrite - last.LevelDB.PointsWrite,
						metric.LevelDB.PointsDelete - last.LevelDB.PointsDelete,
						leveldb_opend,
						metric.LevelDB.WriteTimeMin,
						metric.LevelDB.WriteTimeMax,
						metric.LevelDB.WriteTimeAvg,
						metric.LevelDB.BytesWritten - last.LevelDB.BytesWritten,
					},
				},
			})

			container = append(container, &influxdb.Series{
				Name: fmt.Sprintf("influxdb.%s.go", metric.Name),
				Columns: []string{
					"current_goroutines",
					"goroutines_avg",
					"cgo_call",
					"goroutines_per_conn",
				},
				Points: [][]interface{}{
					[]interface{}{
						metric.Go.CurrentGoroutines,
						metric.Go.GoroutinesAvg,
						metric.Go.CgoCall - last.Go.CgoCall,
						uint64(metric.Go.CurrentGoroutines) / metric.Net.CurrentConnections,
					},
				},
			})

			container = append(container, &influxdb.Series{
				Name: fmt.Sprintf("influxdb.%s.sys", metric.Name),
				Columns: []string{
					"rusage_user",
					"rusage_sys",
					"sys_bytes",
					"alloc",
				},
				Points: [][]interface{}{
					[]interface{}{
						(float64(metric.Sys.Rusage.User.Sec) + float64(metric.Sys.Rusage.User.Usec/1000000)) - (float64(last.Sys.Rusage.User.Sec) + float64(last.Sys.Rusage.User.Usec/1000000)),
						(float64(metric.Sys.Rusage.System.Sec) + float64(metric.Sys.Rusage.System.Usec/1000000)) - (float64(last.Sys.Rusage.System.Sec) + float64(last.Sys.Rusage.System.Usec/1000000)),
						metric.Sys.Alloc,
						metric.Sys.SysBytes,
					},
				},
			})

			mem := sigar.Mem{}
			mem.Get()
			container = append(container, &influxdb.Series{
				Name:    fmt.Sprintf("influxdb.%s.mem", metric.Name),
				Columns: []string{"free", "used", "actualfree", "actualused", "total"},
				Points: [][]interface{}{
					[]interface{}{
						mem.Free,
						mem.Used,
						mem.ActualFree,
						mem.ActualUsed,
						mem.Total,
					},
				},
			})

			load := sigar.LoadAverage{}
			load.Get()
			container = append(container, &influxdb.Series{
				Name:    fmt.Sprintf("influxdb.%s.load", metric.Name),
				Columns: []string{"one", "five", "fifteen"},
				Points: [][]interface{}{
					[]interface{}{
						load.One,
						load.Five,
						load.Fifteen,
					},
				},
			})

			cpu.Get()
			container = append(container, &influxdb.Series{
				Name:    fmt.Sprintf("influxdb.%s.cpu", metric.Name),
				Columns: []string{"id", "user", "nice", "sys", "idle", "wait", "irq", "softirq", "stolen", "total"},
				Points: [][]interface{}{
					[]interface{}{
						"cpu",
						cpu.User - lastCpu.User,
						cpu.Nice - lastCpu.Nice,
						cpu.Sys - lastCpu.Sys,
						cpu.Idle - lastCpu.Idle,
						cpu.Wait - lastCpu.Wait,
						cpu.Irq - lastCpu.Irq,
						cpu.SoftIrq - lastCpu.SoftIrq,
						cpu.Stolen - lastCpu.Stolen,
						cpu.Total() - lastCpu.Total(),
					},
				},
			})

			if err := client.WriteSeriesWithTimePrecision(container, "s"); err != nil {
				fmt.Printf("Error: %s\n", err)
			}
			container = container[:0]
		}

		last = *metric
		lastCpu = cpu
	sleep:
		time.Sleep(1 * time.Second)
	}
}

func main() {
	var err error

	backGround := flag.Bool("background", false, "run as background")
	configFile := flag.String("config", "config.toml", "the config file")
	pidFile := flag.String("pidfile", "", "the pid file")
	flag.Parse()

	config, err = configuration.LoadConfiguration(*configFile)
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(-1)
	}

	if *backGround {
		err := util.Daemonize(0, 0)
		if err != 0 {
			os.Exit(-1)
		}
	}

	if pidFile != nil && *pidFile != "" {
		util.WritePid(*pidFile)
	}

	go loop()
	select {}
}
