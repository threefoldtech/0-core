package pm

const (
	//StateSuccess successs exit status
	StateSuccess JobState = "SUCCESS"
	//StateError error exist status
	StateError JobState = "ERROR"
	//StateTimeout timeout exit status
	StateTimeout JobState = "TIMEOUT"
	//StateKilled killed exit status
	StateKilled JobState = "KILLED"
	//StateUnknownCmd unknown cmd exit status
	StateUnknownCmd JobState = "UNKNOWN_CMD"
	//StateDuplicateID dublicate id exit status
	StateDuplicateID JobState = "DUPILICATE_ID"
)

const (
	//LevelStdout stdout message
	LevelStdout uint16 = 1 // stdout
	//LevelStderr stderr message
	LevelStderr uint16 = 2 // stderr
	//LevelPublic public message
	LevelPublic uint16 = 3 // message for endusers / public message
	//LevelOperator operator message
	LevelOperator uint16 = 4 // message for operator / internal message
	//LevelUnknown unknown message
	LevelUnknown uint16 = 5 // log msg (unstructured = level5, cat=unknown)
	//LevelStructured structured message
	LevelStructured uint16 = 6 // log msg structured
	//LevelWarning warning message
	LevelWarning uint16 = 7 // warning message
	//LevelOpsError ops error message
	LevelOpsError uint16 = 8 // ops error
	//LevelCritical critical message
	LevelCritical uint16 = 9 // critical error
	//LevelStatsd statsd message
	LevelStatsd uint16 = 10 // statsd message(s) AVG
	//LevelDebug debug message
	LevelDebug uint16 = 11 // debug message
	//LevelResultJSON json result message
	LevelResultJSON uint16 = 20 // result message, json
	//LevelResultYAML yaml result message
	LevelResultYAML uint16 = 21 // result message, yaml
	//LevelResultTOML toml result message
	LevelResultTOML uint16 = 22 // result message, toml
	//LevelResultHRD hrd result message
	LevelResultHRD uint16 = 23 // result message, hrd
	//LevelResultJob job result message
	LevelResultJob uint16 = 30 // job, json (full result of a job)
)

var (
	//ResultMessageLevels known results types
	ResultMessageLevels = []uint16{LevelResultJSON,
		LevelResultYAML, LevelResultTOML, LevelResultHRD, LevelResultJob}
)

//JobState of a job
type JobState string

//Streams holds stdout and stderr of a job
type Streams []string

//Stdout getter for stdout
func (s Streams) Stdout() string {
	if len(s) >= 1 {
		return s[0]
	}
	return ""
}

//Stderr getter for stderr
func (s Streams) Stderr() string {
	if len(s) >= 2 {
		return s[1]
	}
	return ""
}

//JobResult represents a result of a job
type JobResult struct {
	ID        string   `json:"id"`
	Command   string   `json:"command"`
	Data      string   `json:"data"`
	Streams   Streams  `json:"streams,omitempty"`
	Critical  string   `json:"critical,omitempty"`
	Level     uint16   `json:"level"`
	State     JobState `json:"state"`
	Code      uint32   `json:"code"`
	StartTime int64    `json:"starttime"`
	Time      int64    `json:"time"`
	Tags      Tags     `json:"tags"`
	Container uint64   `json:"container"`
}

//NewJobResult creates a new job result from command
func NewJobResult(cmd *Command) *JobResult {
	return &JobResult{
		ID:      cmd.ID,
		Command: cmd.Command,
		Tags:    cmd.Tags,
	}
}
