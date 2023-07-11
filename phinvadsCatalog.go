package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"phinvads-catalog/phinvads"
	"regexp"
	"strings"
	"time"
)

const MaxRetries = 10
const url = "https://phinvads.cdc.gov/baseStu3/ValueSet/"
const fetchPages = "?_getpages="

var totalEntries = int64(0)
var countEntries = int64(0)
var durationInSeconds = 3

// a regex to strip out the newlines
var re = regexp.MustCompile(`\r?\n`)

// check - checks for an error on a result, like reading a file, etc
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// writeJsonToFile - writes the JSON we receive to a file
func writeJsonToFile(response *phinvads.Response) {
	// create the output folder
	jsonFolder := "results/valuesets"
	if _, err := os.Stat(jsonFolder); os.IsNotExist(err) {
		err := os.Mkdir(jsonFolder, 0700)
		check(err)
	}
	// loop through each entry in the response
	for _, entry := range response.Entry {
		// write out the file given the json contents and the name
		// this is done in an anonymous function so defer works correctly
		func() {
			// create the file to write out to
			f, err := os.Create(jsonFolder + "/" + entry.Resource.Name + ".json")
			check(err)
			defer f.Close()
			// marshall out the JSON
			asJson, err := json.MarshalIndent(entry, "", "\t")
			check(err)
			f.Write(asJson)
		}()
	}
}

// fetchUri pulls the data at url for PHINVADS and turns it into
// the requisite FHIR struct that we are interested in
func fetchUri(url string) phinvads.Response {
	// this it the object that will be reconstituted
	var response phinvads.Response
	// get the value and check the error
	resp, err := http.Get(url)
	// check for an error
	check(err)
	// decode the contents of the body
	decoder := json.NewDecoder(resp.Body)
	if decoder.More() {
		err := decoder.Decode(&response)
		check(err)
	}

	// write the data out
	writeJsonToFile(&response)

	return response
}

// getResponseLinks - Given a response, grabs the links available
// the body of the response from PHINVADS actually is supposed to support paging
// so there is a link for the current page and a link for the next set of values
func getResponseLinks(response *phinvads.Response) (string, string) {
	nextUrl := ""
	selfUrl := ""
	fmt.Println(response.Id)
	for _, link := range response.Link {
		if link.Relation == "next" {
			nextUrl = strings.TrimSpace(link.Url)
		}
		if link.Relation == "self" {
			selfUrl = strings.TrimSpace(link.Url)
		}
	}
	return nextUrl, selfUrl
}

// getNextResponse - gets the response object from the URL nextUrl
func getNextResponse(nextUrl string) (*phinvads.Response, error) {
	// sanity check the url
	if len(strings.TrimSpace(nextUrl)) == 0 {
		return nil, errors.New("cannot fetch an empty URL")
	}
	// we might need to retry
	retryCount := 0
	for {
		fmt.Printf("Processed %d of %d records so far...\n", countEntries, totalEntries)
		fmt.Printf("Fetching %s\n\n", nextUrl)
		response := fetchUri(nextUrl)
		// we may have exceeded some throttling threshold. let's try to back off a bit
		if len(response.Entry) == 0 {
			// increment our retry count
			retryCount += 1
			// oh you done messed up now
			if retryCount > MaxRetries {
				return nil, errors.New(fmt.Sprintf("Max retries exceeded. Cannot retrieve %s", nextUrl))
			}
			duration := durationInSeconds * retryCount
			fmt.Printf("Sleeping for %d seconds\n", duration)
			time.Sleep(time.Duration(duration) * time.Second)
		} else {
			return &response, nil
		}
	}
}

// writeEntriesToCsv takes the response object and the CSV writer object and writes the data out
func writeEntriesToCsv(response *phinvads.Response, writer *csv.Writer) string {
	lastId := ""
	for _, entry := range response.Entry {
		selfUrl := fmt.Sprintf("%s%s", url, entry.Resource.Id)
		publisher := strings.TrimSpace(re.ReplaceAllString(entry.Resource.Publisher, " "))
		lastId = entry.Resource.Id
		writer.Write([]string{
			selfUrl,
			lastId,
			entry.Resource.Name,
			entry.Resource.Title,
			publisher,
			entry.Resource.Date,
		})
	}
	writer.Flush()

	return lastId
}

func getAllResponses() []phinvads.Response {
	var responses = make([]phinvads.Response, 0, 0)
	// fetch our first result
	response, err := getNextResponse(url)
	// append responses
	responses = append(responses, *response)
	// set our marker for the total number of responses
	totalEntries = response.Total
	// loop beautifully, we love a big beautiful loop don't we folks
	for {
		nextUrl, _ := getResponseLinks(response)
		// increment our counts
		countEntries += int64(len(response.Entry))
		// getting the last id because we may need it to fetch data later
		lastId := response.Entry[len(response.Entry)-1].Resource.Id
		// fetch the next, kind of like a lookahead
		response, err = getNextResponse(nextUrl)
		if err != nil {
			// the canonical nextUrl failed. let's try to reach back one Id value and see if that works instead
			fmt.Fprintf(os.Stderr, "Unable to fetch based on canonical nextUrl %s. Maybe trying saving throw!\n", nextUrl)
			if countEntries < totalEntries {
				// output if we have a problem
				fmt.Fprintf(
					os.Stderr,
					"Danger Will Robinson! We only got %d of %d entries\n",
					countEntries,
					totalEntries,
				)
				// build a new URL
				nextUrl = fmt.Sprintf("%s%s%s", url, fetchPages, lastId)
				fmt.Fprintf(os.Stderr, "Retry fetching %s instead", nextUrl)
				// try this new URL
				response, err = getNextResponse(nextUrl)
				// did it also fail? well shit
				if err != nil {
					// log the error
					fmt.Fprintf(os.Stderr, "Error trying saving throw: %s! Exiting\n\n", nextUrl)
					break
				}
				// we are good. let's loop
				if response != nil {
					// append responses
					responses = append(responses, *response)
					continue
				}
			}
		} else {
			// append responses
			responses = append(responses, *response)
		}
		// sanity check to exit loop
		if countEntries >= totalEntries || (response != nil && len(response.Entry) == 0) {
			// exit the loop
			break
		}
	}
	return responses
}

func main() {
	fmt.Println("Starting")
	date := time.Now()
	fmt.Println("Getting all responses")
	// get all our responses
	responses := getAllResponses()
	// write out the results to a CSV
	if _, err := os.Stat("results"); os.IsNotExist(err) {
		err := os.Mkdir("results", 0700)
		check(err)
	}
	file, err := os.Create(fmt.Sprintf("results/results-%s.csv", date.Format("2006-01-02")))
	check(err)
	// defer calls this at the end of the main function
	defer file.Close()
	// get our CSV writer
	writer := csv.NewWriter(file)
	// defer calls this at the end of the main function
	defer writer.Flush()
	// write some headers
	writer.Write([]string{"selfUrl", "id", "name", "title", "publisher", "date"})
	// write them out
	fmt.Println("Writing out responses")
	for _, r := range responses {
		writeEntriesToCsv(&r, writer)
	}
	fmt.Println("Done!")
}
