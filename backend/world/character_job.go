package world

import (
	"fmt"
	"strings"

	"github.com/mmcdole/lunar"
)

type JobId int

type JobData struct {
	Name string
	Description string
	Stats StatBlock
	Scaling StatBlock
}

var JOB_DATA []*JobData
var JOB_NAME_TO_ID map[string]JobId

var JOB_ENCHANTER JobId

const JOB_DATA_FOLDER = "jobs"

func JobIdFromString(jobName string) (JobId, error) {
	for jobId, jobData := range JOB_DATA {
		if strings.EqualFold(jobName, jobData.Name) {
			return JobId(jobId), nil
		}
	}

	return 0, fmt.Errorf("'%s' is not a valid job.", jobName)
}

// LOAD

func (world *World) loadJobData() error {
	// Read job folder
	paths, err := scriptGetFilesFrom(world.dataFolder, JOB_DATA_FOLDER)
	if err != nil {
		return fmt.Errorf("Error opening job data folder: %s", err.Error())
	}

	JOB_DATA = make([]*JobData, 0, len(paths))
	JOB_NAME_TO_ID = make(map[string]JobId)

	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			return err
		}

		// Parse job data
		parser := ScriptParser{}
		jobData := parser.parseJob(table)
		if jobData == nil {
			return fmt.Errorf("%s: %s", path, parser.getError().Error())
		}

		// Check for duplicates
		_, duplicateJobName := JOB_NAME_TO_ID[jobData.Name]
		if duplicateJobName {
			return fmt.Errorf("Job %s has name '%s' which is a duplicate of another job.", path, jobData.Name)
		}

		// Store job in JOB_DATA
		JOB_NAME_TO_ID[jobData.Name] = JobId(len(JOB_DATA))
		JOB_DATA = append(JOB_DATA, jobData)
	}

	var exists bool
	JOB_ENCHANTER, exists = JOB_NAME_TO_ID["Enchanter"]
	if !exists {
		return fmt.Errorf("Enchanter job not found!")
	}

	return nil
}

func (parser *ScriptParser) parseJob(table *lua.Table) *JobData {
	jobData := &JobData{}

	jobData.Name = parser.getString(table, "name")
	jobData.Description = parser.getString(table, "description")
	jobData.Stats = parser.getStatBlock(table, "stats", false)
	jobData.Scaling = parser.getStatBlock(table, "scaling", false)

	if len(parser.problems) != 0 {
		return nil
	}

	return jobData
}
