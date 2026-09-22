package world

import (
	"fmt"
	"strings"
)

type JobId int
const (
	JOB_ALCHEMIST = iota
	JOB_BLACKSMITH
	JOB_ENCHANTER
)

type JobData struct {
	Name string
	Stats StatBlock
	Scaling StatBlock
}

func JobIdFromString(jobName string) (JobId, error) {
	for jobId, jobData := range JOB_DATA {
		if strings.EqualFold(jobName, jobData.Name) {
			return JobId(jobId), nil
		}
	}

	return 0, fmt.Errorf("'%s' is not a valid job.", jobName)
}

var JOB_DATA []*JobData = []*JobData {
	JOB_ALCHEMIST: {
		Name: "Alchemist",
		Stats: StatBlock{},
		Scaling: StatBlock{},
	},

	JOB_BLACKSMITH: {
		Name: "Blacksmith",
		Stats: StatBlock{},
		Scaling: StatBlock{},
	},

	JOB_ENCHANTER: {
		Name: "Enchanter",
		Stats: StatBlock{},
		Scaling: StatBlock{},
	},
}
