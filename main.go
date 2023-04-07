package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"golang.org/x/text/encoding/charmap"
	"leads_atlas_2/files/models"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Fields struct {
	ServiceData        struct{} `json:"serviceData"`
	RawData            string   `json:"rawData"`
	PartnerLeadId      string   `json:"partnerLeadId"`
	PartnerWebmasterId string   `json:"partnerWebmasterId"`
	Phone              string   `json:"phone"`
}

type ResponseAtl struct {
	Status      string `json:"status"`
	InboxLeadId uint32 `json:"inboxLeadId"`
}

func main() {
	paths := []string{"files/input/s_1.txt", "files/input/s_2.txt"}
	for _, path := range paths {

		parsed := parceFile(path)
		saveToFile(parsed)
	}
}

func parceFile(filePath string) []models.Lead {
	var parsed []models.Lead
	f, err := os.Open(filePath)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	var id = 1
	lead := models.Lead{}

	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}

		if checkNewLead(scanner.Text()) {
			lead = models.Lead{}
			lead.ID = uint32(id)
			date, _ := time.Parse("2006-01-02 15:04:05", scanner.Text())
			lead.Date = date.Format("2006-01-02 15:04:05")
			id++
		}
		indexF := strings.Index(scanner.Text(), "FIELDS:")
		if indexF == 0 {
			fields, err := getFields(scanner.Text())
			if err != nil {
				log.Println("68", filePath, err, scanner.Text())
				panic("STOP")
			}
			lead.RawData = fields.RawData
			num, err := strconv.ParseUint(fields.PartnerLeadId, 10, 32)
			if err != nil {
				//fmt.Println("Ошибка преобразования:", filePath, err)
				num = 0
			}
			lead.VzId = uint32(num)
			lead.Phone = fields.Phone
		}

		if strings.Contains(scanner.Text(), "RESPONSE: ") {
			respAtl, err := getResponseAtl(scanner.Text())
			if err != nil {
				log.Println("84", filePath, err, scanner.Text())
				lead.AtlStatus = ""
				lead.AtlId = 0
				//panic("STOP")
			} else {
				lead.AtlId = respAtl.InboxLeadId
				lead.AtlStatus = respAtl.Status
			}

			parsed = append(parsed, lead)

		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return parsed

}

func getFields(str string) (Fields, error) {
	fields := Fields{}
	str = strings.Replace(str, "FIELDS:", "", 1)
	err := json.Unmarshal([]byte(str), &fields)
	if err != nil {
		return Fields{}, err
	}
	return fields, nil
}

func getResponseAtl(str string) (ResponseAtl, error) {
	respAtl := ResponseAtl{}
	str = strings.Replace(str, "RESPONSE: ", "", 1)
	err := json.Unmarshal([]byte(str), &respAtl)
	if err != nil {
		return ResponseAtl{}, err
	}
	return respAtl, nil
}

func checkNewLead(str string) bool {
	_, err := time.Parse("2006-01-02 15:04:05", str)
	if err != nil {
		return false
	}
	return true
}

func saveToFile(leads []models.Lead) {
	file, err := os.OpenFile("files/output/sent.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	file1251, err := os.OpenFile("files/output/sent_1251.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	defer file1251.Close()

	encoder := charmap.Windows1251.NewEncoder()

	writer := csv.NewWriter(file)
	writer.Comma = ';'
	writer1251 := csv.NewWriter(encoder.Writer(file1251))
	writer1251.Comma = ';'

	for _, lead := range leads {
		record := []string{lead.Date, strconv.Itoa(int(lead.VzId)), strconv.Itoa(int(lead.AtlId)), lead.Phone, lead.RawData, lead.AtlStatus}
		err := writer.Write(record)
		if err != nil {
			panic(err)
		}
		err = writer1251.Write(record)
		if err != nil {
			panic(err)
		}
	}
	writer.Flush()
	writer1251.Flush()
}
