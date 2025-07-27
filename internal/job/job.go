package job

// Job is an interface that all job types should implement
type Job interface {
	Execute() Result
	GetID() string
	GetType() string // For example: "ParsingJob"
	GetProgress() float64
}

type Result struct {
	JobID    string
	JobType  string
	Data     interface{}
	Error    error
	WorkerID int
}
