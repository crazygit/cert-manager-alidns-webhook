package alidns

import (
	"errors"
	"fmt"
	"os"
	"testing"

	alidns "github.com/alibabacloud-go/alidns-20150109/v5/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAliDNSClient 是用于测试的 mock 客户端
type MockAliDNSClient struct {
	// 可配置的 mock 行为
	AddDomainRecordFunc       func(request *alidns.AddDomainRecordRequest, runtime *util.RuntimeOptions) (*alidns.AddDomainRecordResponse, error)
	DeleteDomainRecordFunc    func(request *alidns.DeleteDomainRecordRequest, runtime *util.RuntimeOptions) (*alidns.DeleteDomainRecordResponse, error)
	DescribeDomainRecordsFunc func(request *alidns.DescribeDomainRecordsRequest, runtime *util.RuntimeOptions) (*alidns.DescribeDomainRecordsResponse, error)
}

func (m *MockAliDNSClient) AddDomainRecordWithOptions(request *alidns.AddDomainRecordRequest, runtime *util.RuntimeOptions) (*alidns.AddDomainRecordResponse, error) {
	if m.AddDomainRecordFunc != nil {
		return m.AddDomainRecordFunc(request, runtime)
	}
	return &alidns.AddDomainRecordResponse{
		Body: &alidns.AddDomainRecordResponseBody{
			RecordId: tea.String("mock-record-id"),
		},
	}, nil
}

func (m *MockAliDNSClient) DeleteDomainRecordWithOptions(request *alidns.DeleteDomainRecordRequest, runtime *util.RuntimeOptions) (*alidns.DeleteDomainRecordResponse, error) {
	if m.DeleteDomainRecordFunc != nil {
		return m.DeleteDomainRecordFunc(request, runtime)
	}
	return &alidns.DeleteDomainRecordResponse{}, nil
}

func (m *MockAliDNSClient) DescribeDomainRecordsWithOptions(request *alidns.DescribeDomainRecordsRequest, runtime *util.RuntimeOptions) (*alidns.DescribeDomainRecordsResponse, error) {
	if m.DescribeDomainRecordsFunc != nil {
		return m.DescribeDomainRecordsFunc(request, runtime)
	}
	return &alidns.DescribeDomainRecordsResponse{
		Body: &alidns.DescribeDomainRecordsResponseBody{
			TotalCount: tea.Int64(0),
			DomainRecords: &alidns.DescribeDomainRecordsResponseBodyDomainRecords{
				Record: []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord{},
			},
		},
	}, nil
}

func TestAddTXTRecord(t *testing.T) {
	tests := []struct {
		name            string
		existingRecords []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord
		addFuncCalled   bool
		expectError     bool
		errorMsg        string
	}{
		{
			name:            "new record - should create",
			existingRecords: []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord{},
			addFuncCalled:   true,
			expectError:     false,
		},
		{
			name: "record already exists - should return existing",
			existingRecords: []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord{
				{
					RecordId: tea.String("existing-id"),
					Value:    tea.String("test-value"),
				},
			},
			addFuncCalled: false,
			expectError:   false,
		},
		{
			name:            "describe API error",
			existingRecords: nil,
			addFuncCalled:   false,
			expectError:     true,
			errorMsg:        "failed to describe records",
		},
		{
			name:            "add API error",
			existingRecords: []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord{},
			addFuncCalled:   true,
			expectError:     true,
			errorMsg:        "failed to add domain record",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addCalled := false
			mockClient := &MockAliDNSClient{
				DescribeDomainRecordsFunc: func(request *alidns.DescribeDomainRecordsRequest, runtime *util.RuntimeOptions) (*alidns.DescribeDomainRecordsResponse, error) {
					if tt.existingRecords == nil && tt.expectError && tt.name == "describe API error" {
						return nil, errors.New("describe API error")
					}
					return &alidns.DescribeDomainRecordsResponse{
						Body: &alidns.DescribeDomainRecordsResponseBody{
							TotalCount: tea.Int64(int64(len(tt.existingRecords))),
							DomainRecords: &alidns.DescribeDomainRecordsResponseBodyDomainRecords{
								Record: tt.existingRecords,
							},
						},
					}, nil
				},
				AddDomainRecordFunc: func(request *alidns.AddDomainRecordRequest, runtime *util.RuntimeOptions) (*alidns.AddDomainRecordResponse, error) {
					addCalled = true
					if tt.expectError && tt.name == "add API error" {
						return nil, errors.New("add API error")
					}
					return &alidns.AddDomainRecordResponse{
						Body: &alidns.AddDomainRecordResponseBody{
							RecordId: tea.String("new-record-id"),
						},
					}, nil
				},
			}

			provider := &dnsProvider{client: mockClient}
			recordID, err := provider.AddTXTRecord("example.com", "_acme-challenge", "test-value")

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				switch tt.name {
				case "record already exists - should return existing":
					assert.Equal(t, "existing-id", recordID)
				case "new record - should create":
					assert.Equal(t, "new-record-id", recordID)
				}
			}
			assert.Equal(t, tt.addFuncCalled, addCalled)
		})
	}
}

func TestDeleteRecord(t *testing.T) {
	tests := []struct {
		name        string
		recordID    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "successful deletion",
			recordID:    "test-record-id",
			expectError: false,
		},
		{
			name:        "API error",
			recordID:    "test-record-id",
			expectError: true,
			errorMsg:    "failed to delete domain record",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockAliDNSClient{
				DeleteDomainRecordFunc: func(request *alidns.DeleteDomainRecordRequest, runtime *util.RuntimeOptions) (*alidns.DeleteDomainRecordResponse, error) {
					if tt.expectError {
						return nil, errors.New("delete API error")
					}
					return &alidns.DeleteDomainRecordResponse{}, nil
				},
			}

			provider := &dnsProvider{client: mockClient}
			err := provider.DeleteRecord(tt.recordID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteRecordsByKey(t *testing.T) {
	tests := []struct {
		name         string
		records      []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord
		expectDelete int
		expectError  bool
	}{
		{
			name: "single matching record",
			records: []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord{
				{
					RecordId: tea.String("record-1"),
					Value:    tea.String("target-value"),
				},
			},
			expectDelete: 1,
			expectError:  false,
		},
		{
			name: "multiple matching records",
			records: []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord{
				{
					RecordId: tea.String("record-1"),
					Value:    tea.String("target-value"),
				},
				{
					RecordId: tea.String("record-2"),
					Value:    tea.String("target-value"),
				},
			},
			expectDelete: 2,
			expectError:  false,
		},
		{
			name:         "no matching records",
			records:      []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord{},
			expectDelete: 0,
			expectError:  false,
		},
		{
			name:         "describe API error",
			records:      nil,
			expectDelete: 0,
			expectError:  true,
		},
		{
			name: "delete API error",
			records: []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord{
				{
					RecordId: tea.String("record-1"),
					Value:    tea.String("target-value"),
				},
			},
			expectDelete: 1,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleteCalled := 0
			mockClient := &MockAliDNSClient{
				DescribeDomainRecordsFunc: func(request *alidns.DescribeDomainRecordsRequest, runtime *util.RuntimeOptions) (*alidns.DescribeDomainRecordsResponse, error) {
					if tt.name == "describe API error" {
						return nil, errors.New("describe API error")
					}
					return &alidns.DescribeDomainRecordsResponse{
						Body: &alidns.DescribeDomainRecordsResponseBody{
							TotalCount: tea.Int64(int64(len(tt.records))),
							DomainRecords: &alidns.DescribeDomainRecordsResponseBodyDomainRecords{
								Record: tt.records,
							},
						},
					}, nil
				},
				DeleteDomainRecordFunc: func(request *alidns.DeleteDomainRecordRequest, runtime *util.RuntimeOptions) (*alidns.DeleteDomainRecordResponse, error) {
					deleteCalled++
					if tt.name == "delete API error" {
						return nil, errors.New("delete API error")
					}
					return &alidns.DeleteDomainRecordResponse{}, nil
				},
			}

			provider := &dnsProvider{client: mockClient}
			err := provider.DeleteRecordsByKey("example.com", "_acme-challenge", "target-value")

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectDelete, deleteCalled)
		})
	}
}

func TestDescribeRecords(t *testing.T) {
	tests := []struct {
		name           string
		totalCount     int64
		recordsPerPage int // Mock 每次返回的记录数
		expectCalls    int // 期望调用 API 的次数
		expectCount    int // 期望返回的总记录数
		expectError    bool
	}{
		{
			name:           "single page - less than pageSizeRequest",
			totalCount:     3,
			recordsPerPage: 3,
			expectCalls:    1,
			expectCount:    3,
			expectError:    false,
		},
		{
			name:           "exactly one page - equal to pageSizeRequest",
			totalCount:     100,
			recordsPerPage: 100,
			expectCalls:    1,
			expectCount:    100,
			expectError:    false,
		},
		{
			name:           "multiple pages - requires 2 calls",
			totalCount:     150,
			recordsPerPage: 100,
			expectCalls:    2,
			expectCount:    150,
			expectError:    false,
		},
		{
			name:           "multiple pages - requires 3 calls",
			totalCount:     250,
			recordsPerPage: 100,
			expectCalls:    3,
			expectCount:    250,
			expectError:    false,
		},
		{
			name:           "empty result",
			totalCount:     0,
			recordsPerPage: 0,
			expectCalls:    1,
			expectCount:    0,
			expectError:    false,
		},
		{
			name:           "API error",
			totalCount:     0,
			recordsPerPage: 0,
			expectCalls:    0,
			expectCount:    0,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			mockClient := &MockAliDNSClient{
				DescribeDomainRecordsFunc: func(request *alidns.DescribeDomainRecordsRequest, runtime *util.RuntimeOptions) (*alidns.DescribeDomainRecordsResponse, error) {
					if tt.expectError && tt.name == "API error" {
						return nil, errors.New("API error")
					}

					callCount++
					// 验证请求参数
					assert.Equal(t, "example.com", *request.DomainName)
					assert.Equal(t, "_acme-challenge", *request.RRKeyWord)
					assert.Equal(t, "TXT", *request.Type)
					assert.Equal(t, int64(pageSizeRequest), *request.PageSize) // 应该总是 100
					assert.Equal(t, int64(callCount), *request.PageNumber)

					// 计算这次调用应该返回多少条记录
					remaining := tt.totalCount - int64((callCount-1)*tt.recordsPerPage)
					records := []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord{}

					if remaining > 0 && tt.name != "empty result" {
						count := int(remaining)
						if count > tt.recordsPerPage {
							count = tt.recordsPerPage
						}
						for i := 0; i < count; i++ {
							records = append(records, &alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord{
								RecordId: tea.String(fmt.Sprintf("record-%d", (callCount-1)*tt.recordsPerPage+i)),
							})
						}
					}

					return &alidns.DescribeDomainRecordsResponse{
						Body: &alidns.DescribeDomainRecordsResponseBody{
							TotalCount: tea.Int64(tt.totalCount),
							DomainRecords: &alidns.DescribeDomainRecordsResponseBodyDomainRecords{
								Record: records,
							},
						},
					}, nil
				},
			}

			provider := &dnsProvider{client: mockClient}
			records, err := provider.DescribeRecords("example.com", "_acme-challenge")

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectCount, len(records))
				// 验证确实调用了预期的次数
				assert.Equal(t, tt.expectCalls, callCount, "API call count mismatch")
			}
		})
	}
}

func TestGetEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		envRegion      string
		expectedResult string
	}{
		{
			name:           "default endpoint - no env var",
			envRegion:      "",
			expectedResult: "alidns.aliyuncs.com",
		},
		{
			name:           "custom region - cn-hangzhou",
			envRegion:      "cn-hangzhou",
			expectedResult: "alidns.cn-hangzhou.aliyuncs.com",
		},
		{
			name:           "custom region - cn-beijing",
			envRegion:      "cn-beijing",
			expectedResult: "alidns.cn-beijing.aliyuncs.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original env var
			originalValue := os.Getenv("ALIBABA_CLOUD_REGION_ID")
			defer func() {
				if originalValue != "" {
					mustSetEnv(t, "ALIBABA_CLOUD_REGION_ID", originalValue)
				} else {
					mustUnsetEnv(t, "ALIBABA_CLOUD_REGION_ID")
				}
			}()

			// Set test env var
			if tt.envRegion != "" {
				mustSetEnv(t, "ALIBABA_CLOUD_REGION_ID", tt.envRegion)
			} else {
				mustUnsetEnv(t, "ALIBABA_CLOUD_REGION_ID")
			}

			result := getEndpoint()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func mustSetEnv(t *testing.T, key, value string) {
	t.Helper()
	require.NoError(t, os.Setenv(key, value))
}

func mustUnsetEnv(t *testing.T, key string) {
	t.Helper()
	require.NoError(t, os.Unsetenv(key))
}
