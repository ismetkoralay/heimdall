package provider

// ProviderError represents an error that occurs when interacting with a provider. It includes information about the provider, the HTTP status code, whether the error is retryable, and the underlying error.
type ProviderError struct {
	Provider   string
	StatusCode int
	Retryable  bool
	Err        error
}

// Error returns the underlying error associated with the ProviderError.
func (e *ProviderError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return ""
}

// IsRetryable indicates whether the error is retryable based on the Retryable field of the ProviderError.
func (e *ProviderError) IsRetryable() bool {
	return e.Retryable
}

// Unwrap returns the underlying error associated with the ProviderError, allowing for error unwrapping and inspection.
func (e *ProviderError) Unwrap() error {
	return e.Err
}
