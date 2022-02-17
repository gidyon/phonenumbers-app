package main

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	randomdata "github.com/Pallinder/go-randomdata"
	"github.com/gidyon/jumia-exercise/models"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	db, err := gorm.Open(sqlite.Open("phones.db"), &gorm.Config{})
	handleError(err)

	db = db.Debug()

	// Drop all tables
	handleError(db.Migrator().DropTable(&models.Country{}, &models.Phone{}))

	// Auto migrate
	handleError(db.Migrator().AutoMigrate(&models.Country{}, &models.Phone{}))

	// Add countries
	handleError(addCounties(db))

	// Add phones
	handleError(addRandomPhones(db))

	// Perform autogrations
	handleError(db.AutoMigrate(&models.Phone{}, &models.Country{}))

	router := gin.Default()

	router.SetFuncMap(template.FuncMap{
		"toString": func(v interface{}) string {
			return fmt.Sprint(v)
		},
	})

	router.LoadHTMLGlob("templates/*")

	router.GET("/add", func(c *gin.Context) {
		// Render HTML
		c.HTML(http.StatusOK, "add.html", 1)
	})

	router.GET("/", func(c *gin.Context) {
		// Pagination variables
		pageToken := c.Query("pageToken")
		pageSize := c.Query("pageSize")
		ps := 0

		// Filters in query parameters
		countryFilter := c.Query("countryFilter")
		validStateFilter := c.Query("validStateFilter")
		phoneFilter := c.Query("phoneFilter")

		// Get list of phone numbers from database
		sqlDB := db.Model(&models.Phone{}).Order("id desc")

		// Apply filters
		{
			if pageSize != "" {
				v, err := strconv.Atoi(pageSize)
				if err == nil {
					ps = v
					sqlDB = sqlDB.Limit(v)
				}
			} else {
				sqlDB = sqlDB.Limit(50)
				ps = 50
			}
			if pageToken != "" {
				bs, err := base64.RawStdEncoding.DecodeString(pageToken)
				if err != nil {
					c.AbortWithStatus(http.StatusInternalServerError)
					return
				}
				v, err := strconv.Atoi(string(bs))
				if err == nil {
					sqlDB = sqlDB.Where("id > ?", v)
				}
			}

			if countryFilter != "" {
				sqlDB = sqlDB.Where("country_code = ?", countryFilter)
			}
			if validStateFilter != "" {
				sqlDB = sqlDB.Where("state = ?", validStateFilter)
			}
			if phoneFilter != "" {
				sqlDB = sqlDB.Where("number = ?", phoneFilter)
			}
		}

		if pageToken != "" {
			pt, err := base64.StdEncoding.DecodeString(pageToken)
			if err == nil {
				v, err := strconv.Atoi(string(pt))
				if err == nil {
					sqlDB = sqlDB.Where("id > ?", v)
				}
			}
		}

		phones := make([]*models.Phone, 0, ps+1)

		// Execute the query
		err = sqlDB.Model(&models.Phone{}).Find(&phones).Error
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Update next page token
		var nextPageToken string
		if len(phones) > 0 {
			nextPageToken = base64.StdEncoding.EncodeToString([]byte(fmt.Sprint(phones[0].ID)))
		}

		// Get countries
		countries := make([]*models.Country, 0, 10)
		err = db.Model(&models.Country{}).Find(&countries).Error
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Get last phone

		// Render HTML
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":            "Posts",
			"phones":           phones,
			"countries":        countries,
			"pageSize":         pageSize,
			"pageToken":        pageToken,
			"validStateFilter": validStateFilter,
			"countryFilter":    countryFilter,
			"phoneFilter":      phoneFilter,
			"nextPageToken":    nextPageToken,
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

var states = []string{"VALID", "VALID", "NOT_VALID"}

func randomState() string {
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
			State:  randomState(),
			Number: fmt.Sprint(randomdata.Number(100000000, 999999999)),
		}).Error
		if err != nil {
			return err
		}
	}
	return err
}
