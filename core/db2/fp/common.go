package fp

type requestS struct {
	Action  string // login, logout, create, delete, set, get, map, ...
	UserID  string // kto wykonuje ta akcje
	Payload []byte
}

// ---
// get / delete one

type req_oneS struct {
	ID string
}

type req_listQueryS struct {
	Where  string
	Order  string
	Offset int
	Limit  int
}

// ---

type req_setStringS struct {
	ID       string
	NewValue string
}

type req_setBoolS struct {
	ID       string
	NewValue bool
}

type req_setBytesS struct {
	ID       string
	NewValue []byte
}

type req_setInt64S struct {
	ID       string
	NewValue int64
}

type req_setUInt16S struct {
	ID       string
	NewValue uint16
}

type req_setRelationS struct {
	ID     string
	NewKey string
}

// ---

type req_sendMailS struct {
	ToEmail     string
	Subject     string
	HtmlMessage string
}

// ---
