package transformers

type Transformer interface {
	ValidateWith(with map[string]string) error
	Transform(value any, with map[string]string) any
}
