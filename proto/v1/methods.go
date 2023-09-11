package v1

func (s *Error) Error() string {
	return s.Message
}
