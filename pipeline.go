package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"errors"
	"github.com/mohae/deepcopy"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Pipeline struct {
	Steps []interface{} `yaml:"steps"`
}

func generateProjectSteps(step interface{}, projects []Project) []interface{} {
	projectSteps := make([]interface{}, 0)
	for _, project := range projects {
		stepCopy := deepcopy.Copy(step)
		stepCopyMap := stepCopy.(map[interface{}]interface{})

		if project.checkProjectRules(stepCopyMap) {
			stepCopyMap["label"] = fmt.Sprintf("%s %s", stepCopyMap["label"], project.Label)
			env := stepCopyMap["env"].(map[interface{}]interface{})
			env["BUILDPIPE_PROJECT_LABEL"] = project.Label
			env["BUILDPIPE_PROJECT_PATH"] = project.getMainPath()

			projectSteps = append(projectSteps, stepCopy)
		}
	}

	return projectSteps
}

func generateDistinctProjectSteps(step interface{}, projects []Project) []interface{} {
	projectSteps := make([]interface{}, 0)
	for _, project := range projects {
		stepCopy := deepcopy.Copy(step)
		stepCopyMap := stepCopy.(map[interface{}]interface{})

		if project.checkProjectRules(stepCopyMap) {
			stepCopyMap["label"] = stepCopyMap["label"]
			env := stepCopyMap["env"].(map[interface{}]interface{})
			env["BUILDPIPE_PROJECT_LABEL"] = project.Label
			env["BUILDPIPE_PROJECT_PATH"] = project.getMainPath()

			projectSteps = append(projectSteps, stepCopy)
		}
	}

	return projectSteps
}

func generatePipeline(steps []interface{}, projects []Project) *Pipeline {
	generatedSteps := make([]interface{}, 0)

	for _, step := range steps {
		stepMap, _ := step.(map[interface{}]interface{})
		env, _ := stepMap["env"].(map[interface{}]interface{})
		value, ok := env["BUILDPIPE_SCOPE"]
		if ok && value == "project" {
			projectSteps := generateProjectSteps(step, projects)
			generatedSteps = append(generatedSteps, projectSteps...)
		} else if ok && value == "distinct" {
			projectSteps := generateProjectSteps(step, projects)
			generatedSteps = append(generatedSteps, projectSteps...)
			// projectSteps := generateDistinctProjectSteps(step, projects)
			// for _, ps := range projectSteps {
			// 	skip := false
			// 	psMap, _ := ps.(map[interface{}]interface{})
			// 	psLabel, _ := psMap["label"]
			// 	for _, gs := range generatedSteps {
			// 		gsMap, _ := gs.(map[interface{}]interface{})
			// 		gsLabel, _ := gsMap["label"]
			// 		if gsLabel == psLabel {
			// 			skip = true
			//         }
			// 	} 
			// 	if(!skip) {
			// 		generatedSteps = append(generatedSteps, ps)
			// 	}
			// }
		} else {
			errors.New("Must be a project or distinct")
		}
	}

	return &Pipeline{
		Steps: generatedSteps,
	}
}

func Index(vs []interface{}, t interface{}) int {
    for i, v := range vs {
        if v == t {
            return i
        }
    }
    return -1
}

func Includes(vs []interface{}, t interface{}) bool {
    return Index(vs, t) >= 0
}

func uploadPipeline(pipeline Pipeline) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "buildpipe-")
	if err != nil {
		log.Fatalf("Cannot create temporary file: %s\n", err)
	}
	defer os.Remove(tmpFile.Name())

	data, err := yaml.Marshal(&pipeline)

	fmt.Printf("Pipeline:\n%s", string(data))

	err = ioutil.WriteFile(tmpFile.Name(), data, 0644)
	if err != nil {
		log.Fatalf("Error writing outfile: %s\n", err)
	}

	execCommand("buildkite-agent", []string{"pipeline", "upload", tmpFile.Name()})
}
