package groupsync

type Message struct {
	GroupID  string `json:"group_id"`
	MemberID string `json:"member_id"`
	Trial    int    `json:"trial"`
}

// Retry increments Trial by one
func (m *Message) Retry() {
	m.Trial++
}

// Fields returns the structure fields in a map, for easy logging.
func (m *Message) Fields() map[string]interface{} {
	return map[string]interface{}{
		"groupId":  m.GroupID,
		"memberId": m.MemberID,
		"trial":    m.Trial,
	}
}

// ExceededTrialLimit returns true if Trial is greater than limit, given limit is positive.
func (m *Message) ExceededTrialLimit(limit int) bool {
	return limit > 0 && m.Trial > limit
}
