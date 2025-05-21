package errors

var (
	ErrInternalServer     = NewWithInfo(500, "Internal server error, please try again later")
	ErrAdminAccountNotSet = NewWithInfo(501, "Admin account not set, please contact operations team")
	ErrInvalidToken       = NewWithInfo(400, "Current session is invalid, please login first")
	ErrNoPermission       = NewWithInfo(401, "No access permission")
	ErrRequestTimeout     = NewWithInfo(408, "Request timeout")
	ErrArgument           = NewWithInfo(409, "Invalid argument")
	ErrInvalidPlayground  = NewWithInfo(410, "Invalid playground ID")
	BrowserNotInstalled   = NewWithInfo(411, "Browser not installed")
	ErrCurrentPageEmpty   = NewWithInfo(412, "Browser not open any page")
)
