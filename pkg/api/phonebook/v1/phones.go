// Package phonebook is the interface for managing phone records
package phonebook

import "context"

type PhoneBookService interface {
	CreatePhoneRecord(context.Context, PhoneRecord) error
	GetPhoneRecord(context.Context, GetPhoneRecordRequest) (PhoneRecord, error)
	ListPhoneRecords(context.Context, ListPhoneRecordsRequest) (ListPhoneRecordsResponse, error)
	DeletePhoneRecord(context.Context, DeletePhoneRecordRequest) error
}

type PhoneRecord struct {
	Id          string `json:"id,omitempty"`
	CustName    string `json:"cust_name,omitempty"`
	CustId      string `json:"cust_id,omitempty"`
	CountryName string `json:"country_name,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
	Number      string `json:"number,omitempty"`
	PhoneValid  bool   `json:"phone_valid,omitempty"`
	CreateDate  string `json:"create_date,omitempty"`
}

type GetPhoneRecordRequest struct {
	RecordId string `json:"record_id,omitempty"`
}

type ListPhoneRecordsRequest struct {
	PageSize  int32                `json:"page_size,omitempty"`
	PageToken string               `json:"page_token,omitempty"`
	Filters   *PhoneRecordsFilters `json:"filters,omitempty"`
}

type PhoneRecordsFilters struct {
	CountryCode  int    `json:"country_code,omitempty"`
	ValidOnly    int    `json:"valid_only,omitempty"`
	NotValidOnly bool   `json:"not_valid_only,omitempty"`
	PhoneNumber  string `json:"phone_number,omitempty"`
}

type ListPhoneRecordsResponse struct {
	PhoneRecords  []*PhoneRecord `json:"phone_records,omitempty"`
	NextPageToken string         `json:"next_page_token,omitempty"`
}

type DeletePhoneRecordRequest struct {
	RecordId string `json:"record_id,omitempty"`
}
