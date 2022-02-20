package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	randomdata "github.com/Pallinder/go-randomdata"
	app_v1 "github.com/gidyon/jumia-exercise/internal/app/v1"
	"github.com/gidyon/jumia-exercise/internal/models"
	phonebook_v1 "github.com/gidyon/jumia-exercise/pkg/api/phonebook/v1"
	"github.com/gidyon/jumia-exercise/pkg/utils/phoneutils"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	port = flag.String("port", ":8080", "Port for server")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if gin.IsDebugging() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Logger instance (Singleton)
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Seed random generators
	rand.Seed(time.Now().UnixNano())

	// Db connection (Pool)
	db, err := gorm.Open(sqlite.Open("phones.db"), &gorm.Config{})
	handleError(err)

	db = db.Debug()

	// This block is for easier demostration purposes as it performs auto-migrations and populates data each time service starts
	// It is not intended for a serious production application
	{
		// Drop all tables
		handleError(db.Migrator().DropTable(&models.Country{}, &models.Phone{}))

		// Auto migrate
		handleError(db.Migrator().AutoMigrate(&models.Country{}, &models.Phone{}))

		// Add countries
		handleError(addCounties(db))

		// Add phones
		handleError(addRandomPhones(db))
	}

	// Singleton instance of phone book service
	appV1, err := app_v1.NewPhoneBookService(ctx, &app_v1.Options{
		SqlDB:  db,
		Logger: &log,
	})
	handleError(err)

	router := gin.Default()

	router.SetFuncMap(template.FuncMap{
		"toString": func(v interface{}) string {
			return fmt.Sprint(v)
		},
	})

	router.LoadHTMLGlob("../../web/templates/*")

	router.POST("/addPhone", func(c *gin.Context) {
		var (
			// Pagination variables
			countryName = c.PostForm("country")
			number      = c.PostForm("phone")
			err         error
		)

		// Create record
		_, err = appV1.CreatePhoneRecord(c.Request.Context(), &phonebook_v1.PhoneRecord{
			CustId:      "",
			CountryName: countryName,
			CountryCode: 0,
			Number:      number,
			PhoneValid:  false,
		})
		if err != nil {
			log.Error().Msg(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Redirect to home
		c.Redirect(http.StatusFound, "/")
	})

	// Pagination API
	pagination := phoneutils.NewPaginationAPI()

	router.GET("/", func(c *gin.Context) {
		var (
			// Pagination variables
			prevPageToken = c.Query("prevPageToken")
			nextPageToken = c.Query("nextPageToken")
			pageSize      = c.Query("pageSize")
			sessionId     = c.Query("sessionId")
			pageSizeInt   = 20
			pageInfo      = &phoneutils.PageInfo{}

			// Filters in query parameters
			countryCodeFilter = c.Query("countryCodeFilter")
			validStateFilter  = c.Query("validStateFilter")
			phoneFilter       = c.Query("phoneFilter")
		)

		// Page token
		if prevPageToken != "" {
			pageInfo = pagination.GetBackPageInfo(sessionId, prevPageToken)
		} else {
			pageInfo.PageToken = nextPageToken
		}

		// Page size
		if pageSize != "" {
			pageSizeInt, err = strconv.Atoi(pageSize)
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		// Get phone numbers
		listRes, err := appV1.ListPhoneRecords(c.Request.Context(), &phonebook_v1.ListPhoneRecordsRequest{
			PageSize:  int32(pageSizeInt),
			PageToken: pageInfo.PageToken,
			Filters: &phonebook_v1.PhoneRecordsFilters{
				CountryCode:  countryCodeFilter,
				ValidOnly:    validStateFilter == phoneutils.ValidState,
				NotValidOnly: validStateFilter == phoneutils.NotValidState,
				PhoneNumber:  phoneFilter,
			},
		})
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Get all countries
		countries := make([]*models.Country, 0, 10)
		err = db.Model(&models.Country{}).Find(&countries).Error
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Update some values for pagination
		if pagination.SessionExist(sessionId) {
			pageInfo = pagination.SetNextPageInfo(sessionId, listRes.NextPageToken, pageInfo.PageToken)
		} else {
			sessionId = pagination.SetNewSession(listRes.CollectionCount, listRes.NextPageToken)
		}

		// Render HTML
		c.HTML(http.StatusOK, "index.html", gin.H{
			"phones":            listRes.PhoneRecords,
			"countries":         countries,
			"pageSize":          pageSize,
			"validStateFilter":  validStateFilter,
			"countryCodeFilter": countryCodeFilter,
			"phoneFilter":       phoneFilter,
			"nextPageToken":     listRes.NextPageToken,
			"collectionCount":   pageInfo.CollectionCount,
			"pageNumber":        pageInfo.PageNumber,
			"prevPageToken":     pageInfo.PageToken,
			"sessionId":         sessionId,
		})
	})

	handleError(router.Run(*port))
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

var countries = []*models.Country{
	{
		CountryCode: 237,
		CountryName: "Cameroon",
	},
	{
		CountryCode: 251,
		CountryName: "Ethiopia",
	},
	{
		CountryCode: +212,
		CountryName: "Morocco",
	},
	{
		CountryCode: 258,
		CountryName: "Mozambique",
	},
	{
		CountryCode: +256,
		CountryName: "Uganda",
	},
}

func addCounties(db *gorm.DB) error {
	return db.CreateInBatches(countries, 10).Error
}

func randomCountry() *models.Country {
	return countries[rand.Intn(len(countries))]
}

var states = []bool{true, true, false}

func randomState() bool {
	return states[rand.Intn(len(states))]
}

func addRandomPhones(db *gorm.DB) error {
	var err error
	for i := 0; i < 100; i++ {
		country := randomCountry()
		err = db.Create(&models.Phone{
			Country: models.Country{
				CountryCode: country.CountryCode,
				CountryName: country.CountryName,
			},
			PhoneValid: randomState(),
			Number:     fmt.Sprint(randomdata.Number(100000000, 999999999)),
		}).Error
		if err != nil {
			return err
		}
	}
	return err
}
