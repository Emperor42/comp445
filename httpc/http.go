package http

type HTTP struct {
	operation string
	verbose   bool
	headers   map[string]string
	body      string
	file      bool
	inline    bool
}
