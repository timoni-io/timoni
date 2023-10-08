package git

import (
	"encoding/json"
	"time"
)

type GitLog struct {
	Commit      string    `git:"%H"`
	Subject     string    `git:"%s"`
	Body        string    `git:"%b"`
	AuthorName  string    `git:"%aN"`
	AuthorEmail string    `git:"%aE"`
	Timestamp   time.Time `git:"%ct"`
}

func (gl *GitLog) UnmarshalJSON(data []byte) error {
	type Alias GitLog
	x := struct {
		Alias
		Timestamp int64
	}{}

	err := json.Unmarshal(data, &x)
	if err != nil {
		return err
	}

	*gl = GitLog(x.Alias)
	gl.Timestamp = time.Unix(x.Timestamp, 0)

	return nil
}
