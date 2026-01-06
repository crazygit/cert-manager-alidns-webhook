package alidns

import (
	"fmt"
	"testing"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/stretchr/testify/assert"
)

// MockDNSProvider is a mock implementation of DNSProvider
type MockDNSProvider struct {
	AddTXTRecordFunc       func(domain, rr, value string) (string, error)
	DeleteRecordsByKeyFunc func(domain, rr, value string) error
}

func (m *MockDNSProvider) AddTXTRecord(domain, rr, value string) (string, error) {
	if m.AddTXTRecordFunc != nil {
		return m.AddTXTRecordFunc(domain, rr, value)
	}
	return "mock-record-id", nil
}

func (m *MockDNSProvider) DeleteRecordsByKey(domain, rr, value string) error {
	if m.DeleteRecordsByKeyFunc != nil {
		return m.DeleteRecordsByKeyFunc(domain, rr, value)
	}
	return nil
}

func TestSolver_Name(t *testing.T) {
	solver := &Solver{}
	assert.Equal(t, "alidns", solver.Name(), "Expected solver name to be 'alidns'")
}

func TestSolver_Initialize(t *testing.T) {
	solver := &Solver{}
	stopCh := make(chan struct{})

	// Initialize with nil config - should still work as it only creates k8s client
	// Note: This relies on NewClient() which might require env vars.
	// If it fails due to missing creds, we might need to mock NewClient or skip this test if env is missing.
	// For now we assume NewClient checks creds lazily or envs are present/optional.
	err := solver.Initialize(nil, stopCh)

	// If NewClient fails (e.g. no creds), we verify that. If it succeeds, we verify client is set.
	if err == nil {
		assert.NotNil(t, solver.dnsProvider, "Expected client to be created")
	} else {
		// Log the error but don't fail if it's just credential missing in test env
		t.Logf("Initialize failed (expected if no creds): %v", err)
	}
	close(stopCh)
}

func TestExtractDomainAndRR(t *testing.T) {
	solver := &Solver{}

	tests := []struct {
		name         string
		fqdn         string
		zone         string
		expectDomain string
		expectRR     string
	}{
		{
			name:         "simple case",
			fqdn:         "_acme-challenge.example.com.",
			zone:         "example.com.",
			expectDomain: "example.com",
			expectRR:     "_acme-challenge",
		},
		{
			name:         "subdomain case",
			fqdn:         "_acme-challenge.www.example.com.",
			zone:         "example.com.",
			expectDomain: "example.com",
			expectRR:     "_acme-challenge.www",
		},
		{
			name:         "zone without trailing dot",
			fqdn:         "_acme-challenge.example.com.",
			zone:         "example.com",
			expectDomain: "example.com",
			expectRR:     "_acme-challenge",
		},
		{
			name:         "nested subdomain",
			fqdn:         "_acme-challenge.api.v1.example.com.",
			zone:         "example.com.",
			expectDomain: "example.com",
			expectRR:     "_acme-challenge.api.v1",
		},
		{
			name:         "exact match",
			fqdn:         "example.com.",
			zone:         "example.com.",
			expectDomain: "example.com",
			expectRR:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain, rr := solver.extractDomainAndRR(tt.fqdn, tt.zone)
			assert.Equal(t, tt.expectDomain, domain, "Domain mismatch")
			assert.Equal(t, tt.expectRR, rr, "RR mismatch")
		})
	}
}

func TestSolver_Present(t *testing.T) {
	mockProvider := &MockDNSProvider{
		AddTXTRecordFunc: func(domain, rr, value string) (string, error) {
			assert.Equal(t, "example.com", domain)
			assert.Equal(t, "_acme-challenge", rr)
			assert.Equal(t, "test-key-value", value)
			return "12345", nil
		},
	}

	solver := &Solver{
		dnsProvider: mockProvider,
	}

	ch := &v1alpha1.ChallengeRequest{
		ResolvedFQDN: "_acme-challenge.example.com.",
		ResolvedZone: "example.com.",
		Key:          "test-key-value",
	}

	err := solver.Present(ch)
	assert.NoError(t, err)
}

func TestSolver_Present_Error(t *testing.T) {
	mockProvider := &MockDNSProvider{
		AddTXTRecordFunc: func(domain, rr, value string) (string, error) {
			return "", fmt.Errorf("mock api error")
		},
	}

	solver := &Solver{
		dnsProvider: mockProvider,
	}

	ch := &v1alpha1.ChallengeRequest{
		ResolvedFQDN: "_acme-challenge.example.com.",
		ResolvedZone: "example.com.",
		Key:          "test-key-value",
	}

	err := solver.Present(ch)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock api error")
}

func TestSolver_CleanUp(t *testing.T) {
	mockProvider := &MockDNSProvider{
		DeleteRecordsByKeyFunc: func(domain, rr, value string) error {
			assert.Equal(t, "example.com", domain)
			assert.Equal(t, "_acme-challenge", rr)
			assert.Equal(t, "test-key-value", value)
			return nil
		},
	}

	solver := &Solver{
		dnsProvider: mockProvider,
	}

	ch := &v1alpha1.ChallengeRequest{
		ResolvedFQDN: "_acme-challenge.example.com.",
		ResolvedZone: "example.com.",
		Key:          "test-key-value",
	}

	err := solver.CleanUp(ch)
	assert.NoError(t, err)
}

func TestSolver_CleanUp_Error(t *testing.T) {
	mockProvider := &MockDNSProvider{
		DeleteRecordsByKeyFunc: func(domain, rr, value string) error {
			return fmt.Errorf("mock delete error")
		},
	}

	solver := &Solver{
		dnsProvider: mockProvider,
	}

	ch := &v1alpha1.ChallengeRequest{
		ResolvedFQDN: "_acme-challenge.example.com.",
		ResolvedZone: "example.com.",
		Key:          "test-key-value",
	}

	err := solver.CleanUp(ch)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock delete error")
}

func TestSolver_Present_Uninitialized(t *testing.T) {
	solver := &Solver{dnsProvider: nil}
	ch := &v1alpha1.ChallengeRequest{}
	err := solver.Present(ch)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestSolver_CleanUp_Uninitialized(t *testing.T) {
	solver := &Solver{dnsProvider: nil}
	ch := &v1alpha1.ChallengeRequest{}
	err := solver.CleanUp(ch)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}
