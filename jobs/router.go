package jobs

func RunJob(jobName string) {
	switch jobName {
	case "daily":
		dailyMode()
	case "weekly":
		weeklyJob()
	default:
		panic("No job found with the name: " + jobName)
	}
}
