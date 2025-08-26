package walker

type walkResult int

const (
	resultOk walkResult = iota
	resultFail
	resultCancel
)
