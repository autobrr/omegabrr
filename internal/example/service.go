package example

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s Service) Hello() string {
	return "hello world"
}
