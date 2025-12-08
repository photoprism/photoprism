package vision

// PriorityFromTopicality converts topicality scores to our priority scale (-2..5).
func PriorityFromTopicality(topicality float32) int {
	switch {
	case topicality >= 0.95:
		return 5
	case topicality >= 0.90:
		return 4
	case topicality >= 0.75:
		return 3
	case topicality >= 0.6:
		return 2
	case topicality >= 0.4:
		return 1
	case topicality >= 0.15:
		return -1
	default:
		return -2
	}
}
