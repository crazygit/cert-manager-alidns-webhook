package alidns

import (
	"fmt"
	"os"

	alidns "github.com/alibabacloud-go/alidns-20150109/v5/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
)

// Reference:
// https://api.aliyun.com/product/Alidns
const (
	defaultEndpoint = "alidns.aliyuncs.com"
	pageSizeRequest = 100
	recordType      = "TXT"
)

type DNSProvider interface {
	AddTXTRecord(domain, rr, value string) (string, error)
	DeleteRecordsByKey(domain, rr, value string) error
}

// dnsProvider 是 AliDNS 的客户端封装
type dnsProvider struct {
	client *alidns.Client
}

// DNSProvider defines the interface for DNS operations

// NewDNSProvider 创建一个新的 AliDNS 客户端
func NewDNSProvider() (DNSProvider, error) {
	credential, err := credential.NewCredential(nil)
	if err != nil {
		return nil, err
	}
	endpoint := getEndpoint()

	config := &openapi.Config{
		Credential: credential,
		Endpoint:   tea.String(endpoint),
	}
	alidnsClient, err := alidns.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &dnsProvider{
		client: alidnsClient,
	}, nil
}

// AddTXTRecord 添加 TXT 记录
func (p *dnsProvider) AddTXTRecord(domain, rr, value string) (string, error) {
	// 查询现有记录
	records, err := p.DescribeRecords(domain, rr)
	if err != nil {
		return "", fmt.Errorf("failed to describe records: %w", err)
	}

	// 检查是否已存在相同值的记录
	for _, record := range records {
		if record.Value != nil && *record.Value == value {
			// 记录已存在，直接返回
			return *record.RecordId, nil
		}
	}

	// 添加新记录
	request := &alidns.AddDomainRecordRequest{
		DomainName: tea.String(domain),
		RR:         tea.String(rr),
		Type:       tea.String("TXT"),
		Value:      tea.String(value),
	}

	runtime := &util.RuntimeOptions{}
	response, err := p.client.AddDomainRecordWithOptions(request, runtime)
	if err != nil {
		return "", fmt.Errorf("failed to add domain record: %w", err)
	}

	recordId := *response.Body.RecordId
	return recordId, nil
}

// DeleteRecord 删除 TXT 记录
func (p *dnsProvider) DeleteRecord(recordId string) error {
	request := &alidns.DeleteDomainRecordRequest{
		RecordId: tea.String(recordId),
	}

	runtime := &util.RuntimeOptions{}
	_, err := p.client.DeleteDomainRecordWithOptions(request, runtime)
	if err != nil {
		return fmt.Errorf("failed to delete domain record: %w", err)
	}

	return nil
}

// DeleteRecordsByKey 根据 domain、rr、value 删除记录
func (p *dnsProvider) DeleteRecordsByKey(domain, rr, value string) error {
	// 查询记录
	records, err := p.DescribeRecords(domain, rr)
	if err != nil {
		return fmt.Errorf("failed to describe records: %w", err)
	}

	// 删除匹配的记录
	for _, record := range records {
		if record.Value != nil && *record.Value == value {
			if err := p.DeleteRecord(*record.RecordId); err != nil {
				return err
			}
		}
	}

	return nil
}

// DescribeRecords 查询记录
func (p *dnsProvider) DescribeRecords(domain, rr string) ([]*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord, error) {
	var allRecords []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord
	pageNumber := int64(1)
	pageSize := int64(pageSizeRequest)

	for {
		request := &alidns.DescribeDomainRecordsRequest{
			DomainName: tea.String(domain),
			RRKeyWord:  tea.String(rr),
			Type:       tea.String(recordType),
			PageNumber: tea.Int64(pageNumber),
			PageSize:   tea.Int64(pageSize),
		}

		runtime := &util.RuntimeOptions{}
		response, err := p.client.DescribeDomainRecordsWithOptions(request, runtime)
		if err != nil {
			return nil, fmt.Errorf("failed to describe domain records: %w", err)
		}

		if response.Body.DomainRecords != nil && response.Body.DomainRecords.Record != nil {
			allRecords = append(allRecords, response.Body.DomainRecords.Record...)
		}

		// 如果没有更多记录，退出循环
		if response.Body.TotalCount == nil || int64(len(allRecords)) >= *response.Body.TotalCount {
			break
		}
		pageNumber++
	}

	return allRecords, nil
}

func getEndpoint() string {
	region := os.Getenv("ALIBABA_CLOUD_REGION_ID")
	if region == "" {
		return defaultEndpoint
	}
	return fmt.Sprintf("alidns.%s.aliyuncs.com", region)
}
