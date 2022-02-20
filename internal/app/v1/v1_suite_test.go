package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/Pallinder/go-randomdata"
	phonebook_v1 "github.com/gidyon/jumia-exercise/pkg/api/phonebook/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestV1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "App V1 Suite")
}

var phoneBookAPI phonebook_v1.PhoneBookService

var _ = BeforeSuite(func() {
	gormDB, err := gorm.Open(sqlite.Open("phones.db"))
	Expect(err).ShouldNot(HaveOccurred())

	// We use mocks for database
	phoneBookAPI, err = NewPhoneBookService(context.Background(), &Options{
		SqlDB:  gormDB,
		Logger: &zerolog.Logger{},
	})
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = Describe("Phone Record", func() {

	Context("Creating a phone record", func() {
		var (
			req *phonebook_v1.PhoneRecord
			ctx context.Context
		)

		BeforeEach(func() {
			req = &phonebook_v1.PhoneRecord{
				CustId:      fmt.Sprint(randomdata.Number(1, 999)),
				CountryName: randomdata.Country(randomdata.FullCountry),
				CountryCode: 0,
				Number:      randomdata.PhoneNumber(),
				PhoneValid:  false,
			}
			ctx = context.Background()
		})

		When("Creating a phone record with incorrect or missing data", func() {
			It("should fail when number is missing", func() {
				req.Number = ""
				_, err := phoneBookAPI.CreatePhoneRecord(ctx, req)
				Expect(err).Should(HaveOccurred())
			})
			It("should fail when country is missing", func() {
				req.CountryName = ""
				_, err := phoneBookAPI.CreatePhoneRecord(ctx, req)
				Expect(err).Should(HaveOccurred())
			})
		})

		When("Creating a phone record with valid data", func() {
			It("should successfully create phone record", func() {
				pb, err := phoneBookAPI.CreatePhoneRecord(ctx, req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(pb).ShouldNot(BeNil())
			})
		})
	})

	Context("Getting a phone record", func() {
		var (
			req *phonebook_v1.GetPhoneRecordRequest
			ctx context.Context
		)

		BeforeEach(func() {
			req = &phonebook_v1.GetPhoneRecordRequest{
				RecordId: "",
			}
			ctx = context.Background()
		})

		When("Getting a phone record with missing record id", func() {
			It("should fail", func() {
				req.RecordId = ""
				_, err := phoneBookAPI.GetPhoneRecord(ctx, req)
				Expect(err).Should(HaveOccurred())
			})
		})

		When("Getting a phone record with correct details", func() {
			var (
				pb  *phonebook_v1.PhoneRecord
				err error
			)

			Context("Lets create a phone record", func() {
				It("should succeed", func() {
					pb, err = phoneBookAPI.CreatePhoneRecord(ctx, &phonebook_v1.PhoneRecord{
						CustId:      fmt.Sprint(randomdata.Number(1, 999)),
						CountryName: randomdata.Country(randomdata.FullCountry),
						CountryCode: 0,
						Number:      randomdata.PhoneNumber(),
						PhoneValid:  false,
					})
					Expect(err).ShouldNot(HaveOccurred())
					Expect(pb).ShouldNot(BeNil())
				})
			})

			Context("Getting the created phone record", func() {
				It("should succeed", func() {
					req.RecordId = pb.Id
					record, err := phoneBookAPI.GetPhoneRecord(ctx, req)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(record.CountryName).To(Equal(pb.CountryName))
					Expect(record.CountryCode).To(Equal(pb.CountryCode))
					Expect(record.Number).To(Equal(pb.Number))
					Expect(record.PhoneValid).To(Equal(pb.PhoneValid))
				})
			})
		})
	})

	Context("Deleting a phone record", func() {
		var (
			req *phonebook_v1.DeletePhoneRecordRequest
			ctx context.Context
		)

		BeforeEach(func() {
			req = &phonebook_v1.DeletePhoneRecordRequest{
				RecordId: "",
			}
			ctx = context.Background()
		})

		When("Deleting a phone record with missing record id", func() {
			It("should fail", func() {
				req.RecordId = ""
				err := phoneBookAPI.DeletePhoneRecord(ctx, req)
				Expect(err).Should(HaveOccurred())
			})
		})
	})

	Context("Listing phone records", func() {
		var (
			req *phonebook_v1.ListPhoneRecordsRequest
			ctx context.Context
		)

		BeforeEach(func() {
			req = &phonebook_v1.ListPhoneRecordsRequest{}
			ctx = context.Background()
		})

		When("Listing phone records with no filters", func() {
			It("should succeed", func() {
				_, err := phoneBookAPI.ListPhoneRecords(ctx, req)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})
})
