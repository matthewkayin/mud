package world

import (
	"fmt"
	"log"
	"strings"

	"github.com/mmcdole/lunar"
)

type JobId int

type JobData struct {
	Name string
	Stats StatBlock
	Scaling StatBlock
}

var JOB_DATA []*JobData
var JOB_NAME_TO_ID map[string]JobId

const JOB_DATA_FOLDER = WORLD_DATA_FOLDER + "/jobs"

func JobIdFromString(jobName string) (JobId, error) {
	for jobId, jobData := range JOB_DATA {
		if strings.EqualFold(jobName, jobData.Name) {
			return JobId(jobId), nil
		}
	}

	return 0, fmt.Errorf("'%s' is not a valid job.", jobName)
}

// LOAD

func (world *World) loadJobData() {
	// Read job folder
	log.Printf("Loading job data...")
	paths, err := scriptGetFilesFrom(JOB_DATA_FOLDER)
	if err != nil {
		log.Fatalf("Error opening job data folder: %s", err.Error())
	}

	JOB_DATA = make([]*JobData, 0, len(paths))
	JOB_NAME_TO_ID = make(map[string]JobId)

	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			log.Fatalf("%s: %s", path, err.Error())
		}

		// Parse job data
		parser := ScriptParser{}
		jobData := parser.parseJob(table)
		if jobData == nil {
			log.Fatalf("%s: %s", path, parser.getError().Error())
		}

		// Check for duplicates
		_, duplicateJobName := JOB_NAME_TO_ID[jobData.Name]
		if duplicateJobName {
			log.Fatalf("Job %s has name '%s' which is a duplicate of another job.", path, jobData.Name)
		}

		// Store job in JOB_DATA
		JOB_NAME_TO_ID[jobData.Name] = JobId(len(JOB_DATA))
		JOB_DATA = append(JOB_DATA, jobData)
		log.Printf("Loaded job '%s'.", jobData.Name)
	}

	log.Printf("All job data has been loaded.")
}

func (parser *ScriptParser) parseJob(table *lua.Table) *JobData {
	jobData := &JobData{}

	jobData.Name = parser.getString(table, "name")
	jobData.Stats = parser.getStatBlock(table, "stats", false)
	jobData.Scaling = parser.getStatBlock(table, "scaling", false)

	if len(parser.problems) != 0 {
		return nil
	}

	return jobData
}
