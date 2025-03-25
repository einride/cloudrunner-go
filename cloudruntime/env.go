package cloudruntime

import (
	"os"
	"path"
	"strconv"
)

// Service returns the service name of the current runtime.
func Service() (string, bool) {
	if service, ok := os.LookupEnv("K_SERVICE"); ok {
		return service, true
	}
	// Default to the name of the entrypoint command.
	return path.Base(os.Args[0]), true
}

// Revision returns the service revision of the current runtime.
func Revision() (string, bool) {
	return os.LookupEnv("K_REVISION")
}

// Configuration returns the service configuration of the current runtime.
func Configuration() (string, bool) {
	return os.LookupEnv("K_CONFIGURATION")
}

// Port returns the service port of the current runtime.
func Port() (int, bool) {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	return port, err == nil
}

// Job returns the name of the Cloud Run job being run.
func Job() (string, bool) {
	return os.LookupEnv("CLOUD_RUN_JOB")
}

// Execution returns the name of the Cloud Run job execution being run.
func Execution() (string, bool) {
	return os.LookupEnv("CLOUD_RUN_EXECUTION")
}

// TaskIndex returns the index of the Cloud Run job task being run.
// Starts at 0 for the first task and increments by 1 for every successive task,
// up to the maximum number of tasks minus 1.
func TaskIndex() (int, bool) {
	taskIndex, err := strconv.Atoi(os.Getenv("CLOUD_RUN_TASK_INDEX"))
	return taskIndex, err == nil
}

// TaskAttempt returns the number of time this Cloud Run job task tas been retried.
// Starts at 0 for the first attempt and increments by 1 for every successive retry,
// up to the maximum retries value.
func TaskAttempt() (int, bool) {
	taskIndex, err := strconv.Atoi(os.Getenv("CLOUD_RUN_TASK_ATTEMPT"))
	return taskIndex, err == nil
}

// TaskCount returns the number of tasks in the current Cloud Run job.
func TaskCount() (int, bool) {
	taskIndex, err := strconv.Atoi(os.Getenv("CLOUD_RUN_TASK_COUNT"))
	return taskIndex, err == nil
}

// EnablePubsubTracing returns a boolean indicating whether Pub/Sub tracing is enabled (false by default).
func EnablePubsubTracing() (bool, bool) {
	enablePubsubTracing, err := strconv.ParseBool(os.Getenv("ENABLE_PUBSUB_TRACING"))
	return enablePubsubTracing, err == nil
}
