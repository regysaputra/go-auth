package usecase

// TokenGenerator interface for token generation
type TokenGenerator interface {
	GenerateToken(subject any, purpose string) (string, error)
}
