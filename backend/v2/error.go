package platform

// Error is a domain error encountered while processing CMP requests.
type Error string

// Error returns the string of an error.
func (e Error) Error() string {
	return string(e)
}
