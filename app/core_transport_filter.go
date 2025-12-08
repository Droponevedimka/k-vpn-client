package main

// Transport Filter - filters unsupported transport types from subscriptions
// Currently sing-box does not support xhttp transport (Xray-core specific)

// UnsupportedTransports lists transport types not supported by sing-box
var UnsupportedTransports = []string{
	"xhttp",      // Xray-core specific transport
	"splithttp",  // Old name for xhttp
}

// IsTransportSupported checks if a transport type is supported by sing-box
func IsTransportSupported(transport string) bool {
	for _, unsupported := range UnsupportedTransports {
		if transport == unsupported {
			return false
		}
	}
	return true
}

// FilterResult contains information about filtered proxies
type FilterResult struct {
	Supported   []ProxyConfig // Proxies with supported transports
	Filtered    []ProxyConfig // Proxies with unsupported transports (filtered out)
	Message     string        // Human-readable message about filtered proxies
	AllFiltered bool          // True if ALL proxies were filtered (none supported)
}

// FilterUnsupportedTransports filters out proxies with unsupported transport types
// Returns supported proxies and information about filtered ones
func FilterUnsupportedTransports(proxies []ProxyConfig) FilterResult {
	result := FilterResult{
		Supported: make([]ProxyConfig, 0),
		Filtered:  make([]ProxyConfig, 0),
	}

	filteredInfo := []string{}

	for _, proxy := range proxies {
		if IsTransportSupported(proxy.Network) {
			result.Supported = append(result.Supported, proxy)
		} else {
			result.Filtered = append(result.Filtered, proxy)
			// Create human-readable info
			info := proxy.Name
			if info == "" {
				info = proxy.Server
			}
			filteredInfo = append(filteredInfo, info+" (транспорт: "+proxy.Network+")")
		}
	}

	// Set AllFiltered flag
	result.AllFiltered = len(result.Supported) == 0 && len(result.Filtered) > 0

	// Generate message
	if len(result.Filtered) > 0 {
		if result.AllFiltered {
			result.Message = "Все серверы в подписке используют неподдерживаемый транспорт (xhttp). " +
				"Этот протокол пока не поддерживается. Ожидайте обновлений или попросите " +
				"провайдера предоставить серверы с другим транспортом (ws, grpc, tcp)."
		} else {
			result.Message = "Некоторые серверы (" +
				joinStrings(filteredInfo, ", ") +
				") используют неподдерживаемый транспорт и были пропущены."
		}
	}

	return result
}

// joinStrings joins strings with separator (simple implementation to avoid importing strings)
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
