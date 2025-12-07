package main

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// UserWireGuardConfig представляет пользовательскую конфигурацию WireGuard (из стандартного WG конфига)
type UserWireGuardConfig struct {
	Tag            string   `json:"tag"`              // Уникальный тег (латиница, без пробелов)
	Name           string   `json:"name"`             // Отображаемое имя
	PrivateKey     string   `json:"private_key"`      // [Interface] PrivateKey
	LocalAddress   []string `json:"local_address"`    // [Interface] Address
	DNS            string   `json:"dns,omitempty"`    // [Interface] DNS (опционально)
	MTU            int      `json:"mtu,omitempty"`    // [Interface] MTU (опционально)
	PublicKey      string   `json:"public_key"`       // [Peer] PublicKey
	PresharedKey   string   `json:"preshared_key,omitempty"` // [Peer] PresharedKey (опционально)
	AllowedIPs     []string `json:"allowed_ips"`      // [Peer] AllowedIPs
	Endpoint       string   `json:"endpoint"`         // [Peer] Endpoint (host без порта)
	EndpointPort   int      `json:"endpoint_port"`    // Порт из Endpoint
	PersistentKeepalive int `json:"persistent_keepalive,omitempty"` // [Peer] PersistentKeepalive
}

// ParseWireGuardConfig парсит стандартный WireGuard конфиг
func ParseWireGuardConfig(config string) (*UserWireGuardConfig, error) {
	wg := &UserWireGuardConfig{
		LocalAddress: []string{},
		AllowedIPs:   []string{},
		MTU:          1280, // Default MTU
	}

	lines := strings.Split(config, "\n")
	currentSection := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Пропускаем пустые строки и комментарии
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Определяем секцию
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(strings.Trim(line, "[]"))
			continue
		}

		// Парсим key = value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch currentSection {
		case "interface":
			switch key {
			case "privatekey":
				wg.PrivateKey = value
			case "address":
				// Может быть несколько адресов через запятую
				addresses := strings.Split(value, ",")
				for _, addr := range addresses {
					addr = strings.TrimSpace(addr)
					if addr != "" {
						wg.LocalAddress = append(wg.LocalAddress, addr)
					}
				}
			case "dns":
				// Берём только первый DNS сервер
				dnsServers := strings.Split(value, ",")
				if len(dnsServers) > 0 {
					wg.DNS = strings.TrimSpace(dnsServers[0])
				}
			case "mtu":
				if mtu, err := strconv.Atoi(value); err == nil {
					wg.MTU = mtu
				}
			}

		case "peer":
			switch key {
			case "publickey":
				wg.PublicKey = value
			case "presharedkey":
				wg.PresharedKey = value
			case "allowedips":
				// Может быть несколько IP через запятую
				ips := strings.Split(value, ",")
				for _, ip := range ips {
					ip = strings.TrimSpace(ip)
					if ip != "" {
						wg.AllowedIPs = append(wg.AllowedIPs, ip)
					}
				}
			case "endpoint":
				wg.Endpoint = value
				// Извлекаем порт
				if idx := strings.LastIndex(value, ":"); idx != -1 {
					if port, err := strconv.Atoi(value[idx+1:]); err == nil {
						wg.EndpointPort = port
						wg.Endpoint = value[:idx] // Только хост
					}
				}
			case "persistentkeepalive":
				if keepalive, err := strconv.Atoi(value); err == nil {
					wg.PersistentKeepalive = keepalive
				}
			}
		}
	}

	// Валидация обязательных полей
	if wg.PrivateKey == "" {
		return nil, fmt.Errorf("отсутствует PrivateKey")
	}
	if len(wg.LocalAddress) == 0 {
		return nil, fmt.Errorf("отсутствует Address")
	}
	if wg.PublicKey == "" {
		return nil, fmt.Errorf("отсутствует PublicKey")
	}
	if len(wg.AllowedIPs) == 0 {
		return nil, fmt.Errorf("отсутствует AllowedIPs")
	}
	if wg.Endpoint == "" {
		return nil, fmt.Errorf("отсутствует Endpoint")
	}

	return wg, nil
}

// ValidateTag проверяет корректность тега (латиница, без пробелов)
func ValidateTag(tag string) error {
	if tag == "" {
		return fmt.Errorf("тег не может быть пустым")
	}
	if len(tag) > 32 {
		return fmt.Errorf("тег слишком длинный (макс. 32 символа)")
	}
	
	// Только латиница, цифры, дефис и подчёркивание
	validTag := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	if !validTag.MatchString(tag) {
		return fmt.Errorf("тег должен начинаться с буквы и содержать только латинские буквы, цифры, дефис или подчёркивание")
	}

	return nil
}

// ToSingboxOutbound конвертирует WireGuard конфиг в sing-box outbound (deprecated но работает до 1.13.0)
// Используем outbound вместо endpoint чтобы работал detour в DNS серверах
func (wg *UserWireGuardConfig) ToSingboxOutbound() map[string]interface{} {
	// Резолвим hostname в IP если нужно
	endpointAddr := wg.Endpoint
	if net.ParseIP(endpointAddr) == nil {
		// Это hostname, пробуем резолвить
		ips, err := net.LookupIP(endpointAddr)
		if err == nil && len(ips) > 0 {
			// Предпочитаем IPv4
			for _, ip := range ips {
				if ipv4 := ip.To4(); ipv4 != nil {
					endpointAddr = ipv4.String()
					break
				}
			}
			if net.ParseIP(endpointAddr) == nil && len(ips) > 0 {
				endpointAddr = ips[0].String()
			}
		}
	}

	// Deprecated WireGuard outbound format (без вложенного peers)
	// peer_public_key и pre_shared_key на верхнем уровне
	outbound := map[string]interface{}{
		"type":            "wireguard",
		"tag":             wg.Tag,
		"server":          endpointAddr,
		"server_port":     wg.EndpointPort,
		"local_address":   wg.LocalAddress,
		"private_key":     wg.PrivateKey,
		"peer_public_key": wg.PublicKey,
	}

	if wg.PresharedKey != "" {
		outbound["pre_shared_key"] = wg.PresharedKey
	}

	if wg.MTU > 0 {
		outbound["mtu"] = wg.MTU
	}

	return outbound
}

// ToSingboxEndpoint конвертирует WireGuard конфиг в sing-box endpoint (новый формат 1.12+)
// ПРИМЕЧАНИЕ: endpoints пока не поддерживают detour в DNS, поэтому используем ToSingboxOutbound()
func (wg *UserWireGuardConfig) ToSingboxEndpoint() map[string]interface{} {
	// Резолвим hostname в IP если нужно
	endpointAddr := wg.Endpoint
	if net.ParseIP(endpointAddr) == nil {
		// Это hostname, пробуем резолвить
		ips, err := net.LookupIP(endpointAddr)
		if err == nil && len(ips) > 0 {
			// Предпочитаем IPv4
			for _, ip := range ips {
				if ipv4 := ip.To4(); ipv4 != nil {
					endpointAddr = ipv4.String()
					break
				}
			}
			if net.ParseIP(endpointAddr) == nil && len(ips) > 0 {
				endpointAddr = ips[0].String()
			}
		}
	}

	// Формируем peer
	peer := map[string]interface{}{
		"public_key":  wg.PublicKey,
		"allowed_ips": wg.AllowedIPs,
		"address":     endpointAddr,
		"port":        wg.EndpointPort,
	}

	if wg.PresharedKey != "" {
		peer["pre_shared_key"] = wg.PresharedKey
	}

	if wg.PersistentKeepalive > 0 {
		peer["persistent_keepalive_interval"] = wg.PersistentKeepalive
	}

	endpoint := map[string]interface{}{
		"type":        "wireguard",
		"tag":         wg.Tag,
		"system":      false,
		"name":        wg.Tag,
		"private_key": wg.PrivateKey,
		"address":     wg.LocalAddress,
		"peers":       []interface{}{peer},
	}

	if wg.MTU > 0 {
		endpoint["mtu"] = wg.MTU
	}

	return endpoint
}

// GenerateRouteRulesForWireGuard генерирует правила маршрутизации для WireGuard
func GenerateRouteRulesForWireGuard(configs []UserWireGuardConfig) []map[string]interface{} {
	rules := []map[string]interface{}{}

	for _, wg := range configs {
		// Правило для AllowedIPs -> этот WireGuard
		if len(wg.AllowedIPs) > 0 {
			rules = append(rules, map[string]interface{}{
				"ip_cidr":  wg.AllowedIPs,
				"outbound": wg.Tag,
			})
		}
	}

	return rules
}

// WireGuardInfo информация для UI
type WireGuardInfo struct {
	Tag        string   `json:"tag"`
	Name       string   `json:"name"`
	Endpoint   string   `json:"endpoint"`
	AllowedIPs []string `json:"allowed_ips"`
}

// ToInfo конвертирует в структуру для UI
func (wg *UserWireGuardConfig) ToInfo() WireGuardInfo {
	endpoint := wg.Endpoint
	if wg.EndpointPort > 0 {
		endpoint = fmt.Sprintf("%s:%d", wg.Endpoint, wg.EndpointPort)
	}
	
	return WireGuardInfo{
		Tag:        wg.Tag,
		Name:       wg.Name,
		Endpoint:   endpoint,
		AllowedIPs: wg.AllowedIPs,
	}
}

// WireGuardConfig is the format used by NativeWireGuardManager
type WireGuardConfig struct {
	PrivateKey string
	Address    []string
	DNS        string
	MTU        int
	Peers      []WireGuardPeer
}

// WireGuardPeer represents a WireGuard peer configuration
type WireGuardPeer struct {
	PublicKey           string
	PresharedKey        string
	Endpoint            string
	Port                int
	AllowedIPs          []string
	PersistentKeepalive int
}

// ToWireGuardConfig converts UserWireGuardConfig to WireGuardConfig for native manager
func (wg *UserWireGuardConfig) ToWireGuardConfig() *WireGuardConfig {
	return &WireGuardConfig{
		PrivateKey: wg.PrivateKey,
		Address:    wg.LocalAddress,
		DNS:        wg.DNS,
		MTU:        wg.MTU,
		Peers: []WireGuardPeer{
			{
				PublicKey:           wg.PublicKey,
				PresharedKey:        wg.PresharedKey,
				Endpoint:            wg.Endpoint,
				Port:                wg.EndpointPort,
				AllowedIPs:          wg.AllowedIPs,
				PersistentKeepalive: wg.PersistentKeepalive,
			},
		},
	}
}
