package main

import (
	"context"
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

func main() {
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

	// Drop all tables
	handleError(db.Migrator().DropTable(&models.Country{}, &models.Phone{}))

	// Auto migrate
	handleError(db.Migrator().AutoMigrate(&models.Country{}, &models.Phone{}))

	// Add countries
	handleError(addCounties(db))

	// Add phones
	handleError(addRandomPhones(db))

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

	router.GET("/", func(c *gin.Context) {
		var (
			// Pagination variables
			pageToken   = c.Query("pageToken")
			pageSize    = c.Query("pageSize")
			pageSizeInt int

			// Filters in query parameters
			countryCodeFilter = c.Query("countryCodeFilter")
			validStateFilter  = c.Query("validStateFilter")
			phoneFilter       = c.Query("phoneFilter")
		)

		if pageSize != "" {
			pageSizeInt, err = strconv.Atoi(pageSize)
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		// Get phone numbers
		records, err := appV1.ListPhoneRecords(c.Request.Context(), &phonebook_v1.ListPhoneRecordsRequest{
			PageSize:  int32(pageSizeInt),
			PageToken: pageToken,
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

		// Render HTML
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":             "Posts",
			"phones":            records.PhoneRecords,
			"countries":         countries,
			"pageSize":          pageSize,
			"pageToken":         pageToken,
			"validStateFilter":  validStateFilter,
			"countryCodeFilter": countryCodeFilter,
			"phoneFilter":       phoneFilter,
			"nextPageToken":     records.NextPageToken,
		})
	})

	handleError(router.Run(":8080"))
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

var countries = []*models.Country{
	{
		Code: 237,
		Name: "Cameroon",
	},
	{
		Code: 251,
		Name: "Ethiopia",
	},
	{
		Code: +212,
		Name: "Morocco",
	},
	{
		Code: 258,
		Name: "Mozambique",
	},
	{
		Code: +256,
		Name: "Uganda",
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
				Code: country.Code,
				Name: country.Name,
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
