package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/jasonlvhit/gocron"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

type ResponseI struct {
	Message string `json:"message"`
}

type DTO struct {
	Records []Record `json:"records"`
}

type Record struct {
	Name string `json:"name" validate:"required"`
	Addr   string `json:"addr"`
}

var cfApi *cloudflare.API
var cfCtx context.Context
var trueValue = true

func createRecords(records []Record) (error) {

	hostIp, err := getPublicIp()
	if err != nil {
		panic(err.Error())
	}
	println("Host public IP: ", hostIp)

	zoneId := os.Getenv("CLOUDFLARE_ZONE_ID")

	// update the records with cloudflare via the api
	container := cloudflare.ResourceContainer{
		Level: cloudflare.ZoneRouteLevel,
		Identifier: zoneId,
		Type: cloudflare.ZoneType,
	}

	query := cloudflare.ListDNSRecordsParams{
		Type: "A",
	}
	existingRecords, _, err := cfApi.ListDNSRecords(cfCtx, &container, query)
	if err != nil {
		return err
	}

	println("Existing records: ", len(existingRecords))
	println("Records to apply: ", len(records))

	for _, record := range records {
		lowerCaseName := strings.ToLower(record.Name)
		addr := record.Addr
		if addr == "" {
			addr = hostIp
		}

		// check if the record already exists
		updated := false
		for _, existingRecord := range existingRecords {

			// check if the record name is the same, if so update the record
			if existingRecord.Name == lowerCaseName {

				// check if addr is different from the existing record
				if existingRecord.Content == addr {
					updated = true
					break
				}

				// update the record
				record := cloudflare.UpdateDNSRecordParams{
					Type: "A",
					Name: record.Name,
					Content: addr,
					ID: existingRecord.ID,
					Proxied: &trueValue,
				}
				updatedRecord, err := cfApi.UpdateDNSRecord(cfCtx,  &container, record)
				if err != nil {
					return err
				}
				println("Updated record: ", updatedRecord.Name, updatedRecord.ID)
				updated = true
				break
			}
		}

		if !updated {
			// create the record
			record := cloudflare.CreateDNSRecordParams{
				Type: "A",
				Name: record.Name,
				Content: addr,
				Proxied: &trueValue,
			}
			createdRecord, err := cfApi.CreateDNSRecord(cfCtx,  &container, record)
			if err != nil {
				return err
			}
			println("Created record: ", createdRecord.Name, createdRecord.ID)
		}
	}
	
	return nil
}

func handlePost(e echo.Context) error {
	req := e.Request()
	headerToken := req.Header.Get("Authorization")

	if headerToken != os.Getenv("AUTH_HEADER") || headerToken == "" {
		return e.JSON(http.StatusUnauthorized, &ResponseI{Message: "Invalid token"})
	}

	// unmarschal the body into a map
	var dto DTO
	body := req.Body
	err := json.NewDecoder(body).Decode(&dto)
	if err != nil {
		return e.JSON(http.StatusBadRequest, &ResponseI{Message: "Invalid request body"})
	}

	// validate the records
	for _, record := range dto.Records {
		if record.Name == "" {
			return e.JSON(http.StatusBadRequest, &ResponseI{Message: "Invalid record"})
		}
	}

	// create the records\
	err = createRecords(dto.Records)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, &ResponseI{Message: "Failed to create records"})
	}
	
	return nil
}

func handleLocalFile() {
	jsonFileWithRecords := os.Args[1]
	file, err := os.Open(jsonFileWithRecords)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()

	var dto DTO
	err = json.NewDecoder(file).Decode(&dto)
	if err != nil {
		panic(err.Error())
	}

	err = createRecords(dto.Records)
	if err != nil {
		panic(err.Error())
	}
}

func getPublicIp() (string, error) {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	buf := make([]byte, 256)
	n, err := resp.Body.Read(buf)
	if err != nil {
		return "", err
	}

	return string(buf[:n]), nil
}

func main() {	
	err := godotenv.Load()
	if err != nil {
		println("Skipping .env file")
	}

	api, err := cloudflare.NewWithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN"))
	if err != nil {
		panic(err.Error())
	}
	cfApi = api
	cfCtx = context.Background()

	argProvide := len(os.Args) > 1
	
	if !argProvide {
		println("Starting server on port 1338")
		e := echo.New()
		e.POST("/", handlePost)
		err = e.Start(":1338")
		if err != nil {
			panic(err.Error())
		}
	} else {
		handleLocalFile()
		gocron.Every(1).Minute().Do(handleLocalFile)
		<- gocron.Start()
	}
}
