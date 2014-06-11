package statistic

// Statistic structs should support Protobuf.
type Statistic struct {
	Name        string               `json:"name"`
	Version     string               `json:"version"`
	Pid         int                  `json:"pid"`
	Uptime      float64              `json:"uptime"`
	Time        int64                `json:"time"`
	Api         StatisticApi         `json:"api"`
	Wal         StatisticWal         `json:"wal"`
	Shard       StatisticShard       `json:"shard"`
	Coordinator StatisticCoordinator `json:"coordinator"`
	Net         StatisticNet         `json:"net"`
	LevelDB     StatisticLevelDB     `json:"leveldb"`
	Go          StatisticGo          `json:"go"`
	Sys         StatisticSys         `json:"sys"`
	Raft        StatisticRaft        `json:"raft"`
	Protobuf    StatisticProtobuf    `json:"protobuf"`
}

type StatisticWal struct {
	Opening       uint64 `json:"opening"`
	AppendEntry   uint64 `json:"append_entry"`
	CommitEntry   uint64 `json:"commit_entry"`
	BookmarkEntry uint64 `json:"bookmark_entry"`
}

type StatisticShard struct {
	Opening uint   `json:"opening"`
	Delete  uint64 `json:"delete"`
}

type StatisticCoordinator struct {
	CmdQuery       uint64 `json:"cmd_query"`
	CmdSelect      uint64 `json:"cmd_select"`
	CmdWriteSeries uint64 `json:"cmd_write_series"`
	CmdDelete      uint64 `json:"cmd_delete"`
	CmdDrop        uint64 `json:"cmd_drop"`
	CmdListSeries  uint64 `json:"cmd_list_series"`
}

type StatisticNet struct {
	CurrentConnections uint64 `json:"current_connections"`
	Connections        uint64 `json:"connections"`
	BytesWritten       uint64 `json:"bytes_written"`
	BytesRead          uint64 `json:"bytes_read"`
}

type StatisticLevelDB struct {
	PointsRead   uint64  `json:"points_read"`
	PointsWrite  uint64  `json:"points_write"`
	PointsDelete uint64  `json:"points_delete"`
	WriteTimeMin float64 `json:"min"`
	WriteTimeAvg float64 `json:"avg"`
	WriteTimeMax float64 `json:"max"`
	BytesWritten uint64  `json:"bytes_written"`
}

type StatisticGo struct {
	CurrentGoroutines int     `json:"current_goroutines"`
	GoroutinesAvg     float64 `json:"goroutines_avg"`
	CgoCall           int64   `json:"cgo_call"`
}

type StatisticTimeVal struct {
	Sec  int64 `json:"sec"`
	Usec int32 `json:"usec"`
}

type StatisticRusage struct {
	User   StatisticTimeVal `json:"user"`
	System StatisticTimeVal `json:"sys"`
}

type StatisticSys struct {
	Rusage   StatisticRusage `json:"rusage"`
	SysBytes uint64          `json:"sys_bytes"`
	Alloc    uint64          `json:"alloc"`
}

type StatisticRaft struct {
	//TODO
}

type StatisticProtobuf struct {
	//TODO
}

type StatisticApi struct {
	Http StatisticApiHTTP `json:"http"`
}

type StatisticApiHTTPResponseStatus struct {
	Code  int    `json:"code"`
	Count uint64 `json:"count"`
}
type StatisticApiHTTPResponse struct {
	Min    float64                          `json:"min"`
	Avg    float64                          `json:"avg"`
	Max    float64                          `json:"max"`
	Status []StatisticApiHTTPResponseStatus `json:"status"`
}

type StatisticApiHTTP struct {
	BytesWritten       uint64                   `json:"bytes_written"`
	BytesRead          uint64                   `json:"bytes_read"`
	CurrentConnections uint64                   `json:"current_connections"`
	Connections        uint64                   `json:"connections"`
	Response           StatisticApiHTTPResponse `json:"response"`
}
