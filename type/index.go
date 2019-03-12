package _type

import "errors"

type FailedResponse struct {
	Recommend string
	Message   string
	Code      string
}
type DomainRecords struct {
	Record []Record
}
type Record struct {
	RR         string
	Value      string
	RecordId   string
	Type       string
	DomainName string
}
type SuccessedResponse struct {
	PageNumber    float64
	TotalCount    float64
	PageSize      float64
	DomainRecords DomainRecords
}
type Config struct {
	AccessKeyId string
	DomainName  string
	RR          string
	Signature   string
}
type PublicQuery struct {
	AccessKeyId      string
	Format           string
	SignatureMethod  string
	SignatureNonce   int32
	SignatureVersion string
	Timestamp        string
	Version          string
}
type EditQuery struct {
	PublicQuery
	DomainName string
	RR         string
	Action     string
	RecordId   string
	Type       string
	Value      string
}
type ListQuery struct {
	PublicQuery
	DomainName string
	Action     string
	PageSize   int8
	PageNumber int8
	RRKeyWord  string
}

func (this Config) Check() error {
	if this.AccessKeyId == "" {
		return errors.New("config.json中AccessKeyId字段不能为空")
	}
	if this.DomainName == "" {
		return errors.New("config.json中DomainName字段不能为空")
	}
	if this.RR == "" {
		return errors.New("config.json中RR字段不能为空")
	}
	if this.Signature == "" {
		return errors.New("config.json中Signature字段不能为空")
	}
	return nil
}
