package build

// Service provide necessary services for f8-build needed by webhook
type Service interface {
	GetEnvironmentType(giturl string) (string, error)
}

type service struct{}

// New gives an instance of build service
func New() Service {
	return &service{}
}

func (*service) GetEnvironmentType(giturl string) (string, error) {
	return "OSIO", nil
}
