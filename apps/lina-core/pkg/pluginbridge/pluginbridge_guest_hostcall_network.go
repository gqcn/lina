//go:build wasip1

// This file provides guest-side helpers for the governed outbound HTTP host service.

package pluginbridge

// HTTPHostService exposes guest-side helpers for the governed outbound HTTP host service.
type HTTPHostService struct{}

// HTTP returns the outbound HTTP host service guest client.
func HTTP() *HTTPHostService {
	return &HTTPHostService{}
}

// Request executes one governed outbound HTTP request through the host.
func (s *HTTPHostService) Request(
	targetURL string,
	request *HostServiceNetworkRequest,
) (*HostServiceNetworkResponse, error) {
	payload, err := invokeHostService(
		HostServiceNetwork,
		HostServiceMethodNetworkRequest,
		targetURL,
		"",
		MarshalHostServiceNetworkRequest(request),
	)
	if err != nil {
		return nil, err
	}
	return UnmarshalHostServiceNetworkResponse(payload)
}
