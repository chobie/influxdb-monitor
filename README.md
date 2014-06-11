# Influxdb Monitoring Agent

This is example monitoring agent for InfluxDB statistic api.

```
export GOPATH=`pwd`
go get github.com/cloudfoundry/gosigar
go get github.com/influxdb/influxdb-go
go get github.com/BurntSushi/toml
go run src/influxdb-monitor/main.go
```

You can create charts easily with graphana like this.

![screenshot](https://raw.githubusercontent.com/chobie/influxdb-monitor/master/content/screenshot.png)

