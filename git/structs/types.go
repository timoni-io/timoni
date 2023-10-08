package git

type Action uint8

const (
	Unknown Action = iota
	New
	Delete
	Branches
	Log
	Diff
	Files
	Open
)

type Control struct {
	Action   Action
	Branch   string
	Commit   string
	Filename string
	Diff     CommitDiff
}

type GitStatus string

const (
	Added    GitStatus = "A"
	Modified GitStatus = "M"
	Deleted  GitStatus = "D"
)

type GitDiff struct {
	Filename string
	Status   GitStatus
}

type CommitDiff struct {
	A string
	B string
}
