package model

// Context is the business context for each request.
type Context struct {
	UserId   int    `json:"userId"`
	Username string `json:"username"`
	Status   int    `json:"status"`
}
