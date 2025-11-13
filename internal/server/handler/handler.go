package handler

type Service interface {
}
type Handler struct {
	s Service
}

func New(s Service) *Handler {
	return &Handler{s: s}
}
