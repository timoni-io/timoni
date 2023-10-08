package terrors

type Error int

const (
	Success Error = 0

	BadRequest          Error = 1
	InvalidEmail        Error = 2
	MissingTotpToken    Error = 3
	InvalidTotpToken    Error = 4
	InvalidLicenseKey   Error = 5
	InternalServerError Error = 6
	Timeout             Error = 7
	Forbidden           Error = 8

	ValidationError Error = 10
	DatabaseError   Error = 11

	EmailSend              Error = 30
	UserIsBlockedOrInvalid Error = 31

	SessionIDInvalid  Error = 32
	SessionIDNotFound Error = 33

	InstalationNotFound Error = 40
	InstalationBlocked  Error = 41

	ResourceNotFound    Error = 50
	ResourceCreateError Error = 51
	ResourceDeleteError Error = 52
	ResourceExists      Error = 53

	ActionNotFound       Error = 60
	ClusterConfigIsEmpty Error = 61
)
