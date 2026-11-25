package dto

import "time"

// Worker states the admin home reports.
const (
	WorkerUp      = "up"
	WorkerLate    = "late"
	WorkerUnknown = "unknown"
)

// AdminDashboardResponse is the admin home: the accounts, and whether the
// system runs as it should.
type AdminDashboardResponse struct {
	Accounts AdminDashboardAccounts `json:"accounts"`
	Health   AdminDashboardHealth   `json:"health"`
}

// AdminDashboardAccounts counts the accounts: those in use by role, those
// disabled, customers new since Monday, and customers yet to confirm their email.
type AdminDashboardAccounts struct {
	ByRole               []AdminRoleCount `json:"by_role"`
	Disabled             int64            `json:"disabled"`
	NewCustomersThisWeek int64            `json:"new_customers_this_week"`
	UnconfirmedEmails    int64            `json:"unconfirmed_emails"`
}

// AdminRoleCount is how many accounts in use hold a role.
type AdminRoleCount struct {
	Role  string `json:"role"`
	Count int64  `json:"count"`
}

// AdminDashboardHealth is how the system runs: the database and Redis (up or
// down), the worker (up, late or unknown) with its last run, events the outbox
// has held more than five minutes, emails given up on (null when the queue
// cannot be read), and refunds asked for more than an hour ago and not made.
type AdminDashboardHealth struct {
	Database          string     `json:"database"`
	Redis             string     `json:"redis"`
	Worker            string     `json:"worker"`
	WorkerLastRunAt   *time.Time `json:"worker_last_run_at"`
	UndeliveredEvents int64      `json:"undelivered_events"`
	FailedEmails      *int64     `json:"failed_emails"`
	StuckRefunds      int64      `json:"stuck_refunds"`
}
