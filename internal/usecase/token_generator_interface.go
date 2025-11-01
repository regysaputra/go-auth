package usecase

type TokenGenerator interface {
	GenerateToken(subject any, purpose string) (string, error)
}
