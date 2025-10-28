package logging

import "fmt"

func FormatFlowPrefix(flowUuid string, text string) string {
	return fmt.Sprintf("[flow: %s]: %s", flowUuid, text)
}

func FormatFlowTaskPrefix(flowUuid string, taskKey string, text string) string {
	return fmt.Sprintf("[flow: %s, task: %s]: %s", flowUuid, taskKey, text)
}
