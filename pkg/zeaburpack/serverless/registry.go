package zbserverless

var registry = map[string]ServerlessTransformer{
	"golang": TransformGoServerless,
}

// GetTransformer returns the serverless transformer by name.
//
// The second return value is false if the transformer is not found.
func GetTransformer(transformerName string) (ServerlessTransformer, bool) {
	t, ok := registry[transformerName]
	return t, ok
}
