package traefik

import (
	"fmt"
	"log"

	"github.com/lite-dokploy/backend/internal/types"
	"gopkg.in/yaml.v3"
)

type Manager struct {
	enabled bool
}

func NewManager() *Manager {
	return &Manager{enabled: false}
}

func (m *Manager) Enable() {
	m.enabled = true
}

func (m *Manager) InjectDomainLabels(composeBytes []byte, appName string, domains []types.AppDomain) ([]byte, error) {
	if !m.enabled || len(domains) == 0 {
		return composeBytes, nil
	}

	var compose map[string]interface{}
	if err := yaml.Unmarshal(composeBytes, &compose); err != nil {
		return nil, fmt.Errorf("unmarshal compose: %w", err)
	}

	servicesRaw, ok := compose["services"]
	if !ok {
		return composeBytes, nil
	}
	servicesMap, ok := servicesRaw.(map[string]interface{})
	if !ok {
		return composeBytes, nil
	}

	m.ensureTraefikNetwork(compose)

	for i, d := range domains {
		targetService := d.ServiceName
		if targetService == "" {
			if len(servicesMap) == 1 {
				for name := range servicesMap {
					targetService = name
				}
			} else {
				log.Printf("domain[%d]: serviceName empty and multiple services exist, skipping", i)
				continue
			}
		}

		svcRaw, ok := servicesMap[targetService]
		if !ok {
			log.Printf("domain[%d]: service %q not found in compose", i, targetService)
			continue
		}
		svcMap, ok := svcRaw.(map[string]interface{})
		if !ok {
			continue
		}

		labels := m.buildLabels(d, appName, i)
		if existing, ok := svcMap["labels"].([]interface{}); ok {
			labels = append(existing, labels...)
		}
		svcMap["labels"] = labels
	}

	out, err := yaml.Marshal(compose)
	if err != nil {
		return nil, fmt.Errorf("marshal compose: %w", err)
	}

	return out, nil
}

func (m *Manager) ensureTraefikNetwork(compose map[string]interface{}) {
	existing := map[string]interface{}{}
	if raw, ok := compose["networks"]; ok {
		if m, ok := raw.(map[string]interface{}); ok {
			existing = m
		}
	}
	if _, ok := existing["traefik"]; !ok {
		existing["traefik"] = map[string]interface{}{
			"external": true,
		}
	}
	compose["networks"] = existing
}

func (m *Manager) buildLabels(d types.AppDomain, appName string, idx int) []string {
	host := d.Host
	labels := []string{"traefik.enable=true"}

	entrypointWeb := fmt.Sprintf("%s-web-%d", appName, idx)
	entrypointSecure := fmt.Sprintf("%s-websecure-%d", appName, idx)

	rule := fmt.Sprintf("Host(`%s`)", host)
	if d.Path != "" && d.Path != "/" {
		rule += fmt.Sprintf(" && PathPrefix(`%s`)", d.Path)
	}

	labels = append(labels,
		fmt.Sprintf("traefik.http.routers.%s.rule=%s", entrypointWeb, rule),
		fmt.Sprintf("traefik.http.routers.%s.entrypoints=web", entrypointWeb),
		fmt.Sprintf("traefik.http.routers.%s.service=%s", entrypointWeb, entrypointWeb),
		fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port=%d", entrypointWeb, d.Port),
	)

	if d.HTTPS {
		labels = append(labels,
			fmt.Sprintf("traefik.http.routers.%s.rule=%s", entrypointSecure, rule),
			fmt.Sprintf("traefik.http.routers.%s.entrypoints=websecure", entrypointSecure),
			fmt.Sprintf("traefik.http.routers.%s.service=%s", entrypointSecure, entrypointSecure),
			fmt.Sprintf("traefik.http.routers.%s.tls.certresolver=letsencrypt", entrypointSecure),
			fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port=%d", entrypointSecure, d.Port),
		)
		redirectName := fmt.Sprintf("%s-redirect-%d", appName, idx)
		labels = append(labels,
			fmt.Sprintf("traefik.http.middlewares.%s.redirectscheme.scheme=https", redirectName),
			fmt.Sprintf("traefik.http.middlewares.%s.redirectscheme.permanent=true", redirectName),
			fmt.Sprintf("traefik.http.routers.%s.middlewares=%s", entrypointWeb, redirectName),
		)
	}

	var mids []string

	if d.StripPath && d.Path != "" && d.Path != "/" {
		stripName := fmt.Sprintf("%s-stripprefix-%d", appName, idx)
		labels = append(labels,
			fmt.Sprintf("traefik.http.middlewares.%s.stripprefix.prefixes=%s", stripName, d.Path),
		)
		mids = append(mids, stripName)
	}

	if d.InternalPath != "" && d.InternalPath != "/" && d.InternalPath != d.Path {
		prefixName := fmt.Sprintf("%s-addprefix-%d", appName, idx)
		labels = append(labels,
			fmt.Sprintf("traefik.http.middlewares.%s.addprefix.prefix=%s", prefixName, d.InternalPath),
		)
		mids = append(mids, prefixName)
	}

	if len(mids) > 0 {
		midsStr := ""
		for i, m := range mids {
			if i > 0 {
				midsStr += ","
			}
			midsStr += m
		}
		labels = append(labels, fmt.Sprintf("traefik.http.routers.%s.middlewares=%s", entrypointWeb, midsStr))
		if d.HTTPS {
			labels = append(labels, fmt.Sprintf("traefik.http.routers.%s.middlewares=%s", entrypointSecure, midsStr))
		}
	}

	return labels
}

func (m *Manager) EnsureTraefikNetwork() error {
	return nil
}
