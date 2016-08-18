package domfinder

import "fmt"

// Err represents the custom error methods
type Err interface {
	error
	GetCode() int
}

// NewErr is a custom error struct representing the error with additional
// information
type NewErr struct {
	Code    int
	value   string
	deepErr error
}

// GetCode returns the code of the error, useful to reference errMsg
func (e NewErr) GetCode() int {
	return e.Code
}

// Error replaces the default Error method
func (e NewErr) Error() string {
	switch {
	case e.Code == ErrUpgradedError && e.value == "" && e.deepErr != nil:
		return e.deepErr.Error()
	case e.deepErr == nil && e.value == "":
		return errMsg[e.Code]
	case e.deepErr == nil && e.value != "":
		return fmt.Sprintf(errMsg[e.Code], e.value)
	case e.value == "" && e.deepErr != nil:
		return fmt.Sprintf(errMsg[e.Code], e.deepErr)
	default:
		return fmt.Sprintf(errMsg[e.Code], e.value, e.deepErr)
	}
}

// UpgradeErr takes a standard error interface and upgrades it to our
// custom error types
func UpgradeErr(e error) *NewErr {
	return &NewErr{Code: ErrUpgradedError, deepErr: e}
}

// map each error name to a unique id
const (
	ErrUpgradedError = 1 << iota
	ErrNoWebservers
	ErrApacheFetchVhosts
	ErrApacheInvalidVhosts
	ErrApacheParseVhosts
	ErrApacheNoEntries
	ErrNotImplemented
)

// errMsg contains a map of error name id keys and error/deep error pairs
var errMsg = map[int]string{
	ErrUpgradedError:  "Not a real error",
	ErrNoWebservers:   "Did not find any webservers running",
	ErrNotImplemented: "The webserver %s is not implemented at this time",

	// Apache specific
	ErrApacheFetchVhosts:   "Unable to obtain vhost data from Apache: %s",
	ErrApacheInvalidVhosts: "Apache didn't return valid vhost entries when checking %s",
	ErrApacheParseVhosts:   "Unable to parse Apache vhost: %s",
	ErrApacheNoEntries:     "No Apache vhost entries found",
}
