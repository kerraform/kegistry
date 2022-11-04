package request

type Request[T any, K comparable] struct {
	Data *Data[T, K] `json:"data"`
}

type Data[T any, K comparable] struct {
	Attributes *T `json:"attributes"`
	Type       K  `json:"type"`
}
