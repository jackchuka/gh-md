package github

// LabelNode represents a label in the GraphQL response.
type LabelNode struct {
	Name string `json:"name"`
}

// AssigneeNode represents an assignee in the GraphQL response.
type AssigneeNode struct {
	Login string `json:"login"`
}

// extractLabelNames extracts label names from label nodes.
func extractLabelNames(nodes []LabelNode) []string {
	labels := make([]string, 0, len(nodes))
	for _, n := range nodes {
		labels = append(labels, n.Name)
	}
	return labels
}

// extractAssigneeLogins extracts logins from assignee nodes.
func extractAssigneeLogins(nodes []AssigneeNode) []string {
	assignees := make([]string, 0, len(nodes))
	for _, n := range nodes {
		assignees = append(assignees, n.Login)
	}
	return assignees
}
