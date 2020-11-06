package cron

import (
	"encoding/json"
	"strings"
	batch "k8s.io/api/batch/v1beta1"
)

// JobSet is a hashed set of cron jobs
type JobSet struct {
	cronJobMap map[string]bool
}

func getJobHash(job batch.CronJob) string {
	hashString, _ := json.Marshal(job)
	return string(hashString)
}

// In returns true if cronJob is in the set else false
func (jobSet JobSet) In(job batch.CronJob) bool {
	if _, ok := jobSet.cronJobMap[getJobHash(job)]; ok {
		return true
	}
	return false
}

// Add adds a cronJob to the set
func (jobSet JobSet) Add(job batch.CronJob) {
	jobSet.cronJobMap[getJobHash(job)] = true
}

// map1InMap2 returns true if all elements in set1 are in the set2
func map1InMap2(map1 map[string]bool, map2 map[string]bool) bool {
	for key := range map1 {
		_, inSet := map2[key]
		if !inSet {
			return false
		}
	}
	return true
}

// Equals returns true if the other set is equal to this set
func (jobSet JobSet) Equals(otherJobSet *JobSet) bool {
	return map1InMap2(jobSet.cronJobMap, otherJobSet.cronJobMap) && map1InMap2(otherJobSet.cronJobMap, jobSet.cronJobMap)
}

// String returns the string json representation of the JobSet
func (jobSet JobSet) String() string {
	jobStringSlice := make([]string, 0, len(jobSet.cronJobMap))
	for key := range jobSet.cronJobMap {
		jobStringSlice = append(jobStringSlice, key)
	}
	return strings.Join(jobStringSlice, ", ")
}

// NewSet creates a new job set
func NewSet() JobSet {
	return JobSet{make(map[string]bool)}
}

// NewSetFromList creates a new job set from a list of cron jobs
func NewSetFromList(jobList []batch.CronJob) JobSet {
	newSet := NewSet()
	for _, job := range jobList {
		newSet.Add(job)
	}
	return newSet
}
