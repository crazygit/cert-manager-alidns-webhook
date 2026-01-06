package alidns

import (
	"fmt"
	"strings"

	"log/slog"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"k8s.io/client-go/rest"

	"github.com/cert-manager/cert-manager/pkg/issuer/acme/dns/util"
)

// Solver implements the provider-specific logic needed to
// 'present' an ACME challenge TXT record for your own DNS provider.
// To do so, it must implement the `github.com/cert-manager/cert-manager/pkg/acme/webhook.Solver`
// interface.
// 实现 cert-manager webhook Solver interface
type Solver struct {
	// If a Kubernetes 'clientset' is needed, you must:
	// 1. uncomment the additional `dnsProvider` field in this structure below
	// 2. uncomment the "k8s.io/dnsProvider-go/kubernetes" import at the top of the file
	// 3. uncomment the relevant code in the Initialize method below
	// 4. ensure your webhook's service account has the required RBAC role
	//    assigned to it for interacting with the Kubernetes APIs you need.
	//dnsProvider kubernetes.Clientset
	dnsProvider DNSProvider
}

func NewSolver(dnsProvider DNSProvider) *Solver {
	return &Solver{
		dnsProvider: dnsProvider,
	}
}

// Config is a structure that is used to decode into when
// solving a DNS01 challenge.
// This information is provided by cert-manager, and may be a reference to
// additional configuration that's needed to solve the challenge for this
// particular certificate or issuer.
// This typically includes references to Secret resources containing DNS
// provider credentials, in cases where a 'multi-tenant' DNS solver is being
// created.
// If you do *not* require per-issuer or per-certificate configuration to be
// provided to your webhook, you can skip decoding altogether in favour of
// using CLI flags or similar to provide configuration.
// You should not include sensitive information here. If credentials need to
// be used by your provider here, you should reference a Kubernetes Secret
// resource and fetch these credentials using a Kubernetes clientset.

type Config struct {
	// Change the two fields below according to the format of the configuration
	// to be decoded.
	// These fields will be set by users in the
	// `issuer.spec.acme.dns01.providers.webhook.config` field.

	//Email           string `json:"email"`
	//APIKeySecretRef v1alpha1.SecretKeySelector `json:"apiKeySecretRef"`
}

// Name is used as the name for this DNS solver when referencing it on the ACME
// Issuer resource.
// This should be unique **within the group name**, i.e. you can have two
// solvers configured with the same Name() **so long as they do not co-exist
// within a single webhook deployment**.
// For example, `cloudflare` may be used as the name of a solver.

func (s *Solver) Name() string {
	return "alidns"
}

// Present is responsible for actually presenting the DNS record with the
// DNS provider.
// This method should tolerate being called multiple times with the same value.
// cert-manager itself will later perform a self check to ensure that the
// solver has correctly configured the DNS provider.
func (s *Solver) Present(ch *v1alpha1.ChallengeRequest) error {
	if s.dnsProvider == nil {
		return fmt.Errorf("alidns client not initialized")
	}

	// not required in this solver
	// cfg, err := loadConfig(ch.Config)
	// if err != nil {
	// 	return fmt.Errorf("failed to load config: %w", err)
	// }

	// 解析域名和记录名
	domain, rr := s.extractDomainAndRR(ch.ResolvedFQDN, ch.ResolvedZone)

	// 添加 TXT 记录
	recordId, err := s.dnsProvider.AddTXTRecord(domain, rr, ch.Key)
	if err != nil {
		return fmt.Errorf("failed to add TXT record: %w", err)
	}

	slog.Info("Successfully added TXT record",
		"domain", domain,
		"rr", rr,
		"value", ch.Key,
		"recordId", recordId,
	)

	return nil
}

// CleanUp should delete the relevant TXT record from the DNS provider console.
// If multiple TXT records exist with the same record name (e.g.
// _acme-challenge.example.com) then **only** the record with the same `key`
// value provided on the ChallengeRequest should be cleaned up.
// This is in order to facilitate multiple DNS validations for the same domain
// concurrently.
func (s *Solver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	if s.dnsProvider == nil {
		return fmt.Errorf("alidns client not initialized")
	}

	// not required in this solver
	// cfg, err := loadConfig(ch.Config)
	// if err != nil {
	// 	return fmt.Errorf("failed to load config: %w", err)
	// }

	// 解析域名和记录名
	domain, rr := s.extractDomainAndRR(ch.ResolvedFQDN, ch.ResolvedZone)

	// 删除记录（根据 key 值匹配）
	err := s.dnsProvider.DeleteRecordsByKey(domain, rr, ch.Key)
	if err != nil {
		return fmt.Errorf("failed to delete TXT record: %w", err)
	}

	slog.Info("Successfully deleted TXT record",
		"domain", domain,
		"rr", rr,
		"value", ch.Key,
	)
	return nil
}

// Initialize will be called when the webhook first starts.
// This method can be used to instantiate the webhook, i.e. initialising
// connections or warming up caches.
// Typically, the kubeClientConfig parameter is used to build a Kubernetes
// client that can be used to fetch resources from the Kubernetes API, e.g.
// Secret resources containing credentials used to authenticate with DNS
// provider accounts.
// The stopCh can be used to handle early termination of the webhook, in cases
// where a SIGTERM or similar signal is sent to the webhook process.
func (s *Solver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	// 如果需要 Kubernetes 客户端（例如从 Secret 读取配置），可以在这里初始化
	// if kubeClientConfig != nil {
	// 	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to create kubernetes client: %w", err)
	// 	}
	// 	_ = cl // 避免未使用变量警告
	// }
	client, err := NewDNSProvider()
	if err != nil {
		return fmt.Errorf("failed to create alidns client: %w", err)
	}
	s.dnsProvider = client
	return nil
}

// loadConfig is a small helper function that decodes JSON configuration into
// the typed config struct.
// func loadConfig(cfgJSON *extapi.JSON) (*Config, error) {
// 	cfg := &Config{}
// 	if cfgJSON == nil {
// 		return cfg, nil
// 	}

// 	if err := json.Unmarshal(cfgJSON.Raw, cfg); err != nil {
// 		return nil, fmt.Errorf("error decoding solver config: %w", err)
// 	}

// 	return cfg, nil
// }

// extractDomainAndRR 从 FQDN 和 Zone 中提取域名和记录名
// 例如：
//
//	ResolvedFQDN: _acme-challenge.example.com.example.com.
//	ResolvedZone: example.com.
//
// 返回：
//
//	domain: example.com
//	rr: _acme-challenge.example.com
func (s *Solver) extractDomainAndRR(fqdn, zone string) (string, string) {
	fqdn = util.UnFqdn(fqdn)
	zone = util.UnFqdn(zone)

	// 从 FQDN 中移除 zone 部分，得到记录名
	rr := strings.TrimSuffix(fqdn, zone)

	// 移除 rr 可能的结尾点
	rr = strings.TrimSuffix(rr, ".")

	return zone, rr
}
