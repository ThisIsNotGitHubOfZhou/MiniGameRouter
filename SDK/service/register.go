package service

type RegisterService interface {
	Register(username, password string) error

	Deregister(username, password string) error
}
