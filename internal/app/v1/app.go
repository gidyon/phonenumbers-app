// Package app implements phonebook service(v1)
package app

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gidyon/jumia-exercise/internal/models"
	phonebook_v1 "github.com/gidyon/jumia-exercise/pkg/api/phonebook/v1"
	"github.com/gidyon/jumia-exercise/pkg/utils/phoneutils"
	"github.com/gidyon/micro/utils/errs"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

type Options struct {
	SqlDB  *gorm.DB
	Logger *zerolog.Logger
}

func NewPhoneBookService(ctx context.Context, opt *Options) (phonebook_v1.PhoneBookService, error) {
	switch {
	case opt == nil:
		return nil, errors.New("missing opts")
	case opt.SqlDB == nil:
		return nil, errors.New("missing sql db")
	case opt.Logger == nil:
		return nil, errors.New("missing logger")
	}
	pb := &phoneBookAPIServer{
		Options: opt,
	}
	return pb, nil
}

type phoneBookAPIServer struct {
	*Options
}

func (pb *phoneBookAPIServer) CreatePhoneRecord(
	ctx context.Context, req *phonebook_v1.PhoneRecord,
) error {
	// Validate fields
	switch {
	case req == nil:
		return errors.New("missing phonebook")
	case req.CountryName == "":
		return errors.New("missing country")
	case req.Number == "":
		return errors.New("missing phone number")
	}

	// Validate phone
	phoneutils.ValidatePhone(req)

	// Create phone
	err := pb.SqlDB.Create(&models.Phone{
		ID: 0,
		Country: models.Country{
			CountryCode: req.CountryCode,
			CountryName: req.CountryName,
		},
		Number:     req.Number,
		CustId:     req.CustId,
		PhoneValid: req.PhoneValid,
	}).Error
	if err != nil {
		pb.Logger.Error().Str("method", "CreatePhoneRecord").Str("error", err.Error()).Msg("failed to create phone record")
		return errors.New("creating phone record failed")
	}
	return nil
}

func (pb *phoneBookAPIServer) GetPhoneRecord(
	ctx context.Context, req *phonebook_v1.GetPhoneRecordRequest,
) (*phonebook_v1.PhoneRecord, error) {
	if req.RecordId == "" {
		return nil, errors.New("missing phone record id")
	}

	db := &models.Phone{}

	// Get from db
	err := pb.SqlDB.First(db, "id = ?", req.RecordId).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errors.New("record not found")
	default:
		pb.Logger.Error().Str("method", "GetPhoneRecord").Str("error", err.Error()).Msg("failed to get phone record")
		return nil, errors.New("getting phone record failed")
	}

	return &phonebook_v1.PhoneRecord{
		Id:          fmt.Sprint(db.ID),
		CustId:      db.CustId,
		CountryName: db.Country.CountryName,
		CountryCode: db.Country.CountryCode,
		Number:      db.Number,
		PhoneValid:  db.PhoneValid,
		CreateDate:  db.CreateDate.UTC().Format(time.RFC3339),
	}, nil
}

const defaultPageSize = 50

func (pb *phoneBookAPIServer) ListPhoneRecords(
	ctx context.Context, req *phonebook_v1.ListPhoneRecordsRequest,
) (*phonebook_v1.ListPhoneRecordsResponse, error) {
	var (
		pageSize  = req.PageSize
		pageToken = req.PageToken

		ID  uint
		err error
	)

	switch {
	case pageSize <= 0:
		pageSize = defaultPageSize
	case pageSize > defaultPageSize:
		pageSize = defaultPageSize
	}

	if pageToken != "" {
		bs, err := base64.StdEncoding.DecodeString(pageToken)
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to parse page token")
		}
		v, err := strconv.ParseUint(string(bs), 10, 64)
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "incorrect page token")
		}
		ID = uint(v)
		fmt.Println(ID)
	}

	// Default db settings
	db := pb.SqlDB.Unscoped().Limit(int(pageSize + 1)).Order("id DESC").Model(&models.Phone{})
	if ID != 0 {
		db = db.Where("id<?", ID)
	}

	// Apply filters
	if req.Filters != nil {
		if req.Filters.PhoneNumber != "" {
			db = db.Where("number  = ?", req.Filters.PhoneNumber)
		}
		if req.Filters.CountryCode != "" {
			db = db.Where("country_code  = ?", req.Filters.CountryCode)
		}
		if req.Filters.ValidOnly && req.Filters.NotValidOnly {
		} else if req.Filters.ValidOnly {
			db = db.Where("phone_valid  = ?", true)
		} else if req.Filters.NotValidOnly {
			db = db.Where("phone_valid  = ?", false)
		}
	}

	var collectionCount int64

	if pageToken == "" {
		err = db.Count(&collectionCount).Error
		if err != nil {
			return nil, errs.SQLQueryFailed(err, "count")
		}
	}

	dbs := make([]*models.Phone, 0, pageSize+1)
	err = db.Find(&dbs).Error
	switch {
	case err == nil:
	default:
		return nil, errs.SQLQueryFailed(err, "LIST")
	}

	pbs := make([]*phonebook_v1.PhoneRecord, 0, len(dbs))

	for i, db := range dbs {
		if i == int(pageSize) {
			break
		}
		pbs = append(pbs, &phonebook_v1.PhoneRecord{
			Id:          fmt.Sprint(db.ID),
			CustId:      db.CustId,
			CountryName: db.Country.CountryName,
			CountryCode: db.Country.CountryCode,
			Number:      db.Number,
			PhoneValid:  db.PhoneValid,
			CreateDate:  db.CreateDate.UTC().Format(time.RFC3339),
		})
		fmt.Println(db.ID)
		ID = db.ID
	}

	var token string
	if len(dbs) > int(pageSize) {
		// Next page token
		token = base64.StdEncoding.EncodeToString([]byte(fmt.Sprint(ID)))
	}

	fmt.Println(token)

	return &phonebook_v1.ListPhoneRecordsResponse{
		NextPageToken:   token,
		PhoneRecords:    pbs,
		CollectionCount: int32(collectionCount),
	}, nil
}

func (pb *phoneBookAPIServer) DeletePhoneRecord(
	ctx context.Context, req *phonebook_v1.DeletePhoneRecordRequest,
) error {
	if req.RecordId == "" {
		return errors.New("missing phone record id")
	}

	// Delete from db
	err := pb.SqlDB.Delete(&models.Phone{}, "id = ?", req.RecordId).Error
	if err != nil {
		pb.Logger.Error().Str("method", "DeletePhoneRecord").Str("error", err.Error()).Msg("failed to delete phone record")
		return errors.New("deleting phone record failed")
	}

	return nil
}
