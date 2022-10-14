package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
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

// all of these objects map back to the FHIR valueset schema found at:
// https://build.fhir.org/valueset.html

// PhinvadsLink is a link for walking the structure of PHINVADS data
type PhinvadsLink struct {
	Relation string
	Url      string
}

// Identifier comes from the FHIR datatypes. Schema is here:
// https://build.fhir.org/datatypes.html#Identifier
type Identifier struct {
	Use      string
	Type     string
	System   string
	value    string
	period   string
	assigner string
}

// Concept is the actual data elements that make up a valueset, for example "U": "Unknown", etc
type Concept struct {
	Code    string
	Display string
}

// ValueSet is the collection of concepts that make up a valueset
type ValueSet struct {
	Version string
	System  string
	Concept []Concept
}

type Composition struct {
	LockedDate string
	Inactive   bool
	Include    []ValueSet
}

type PhinvadsResource struct {
	ResourceType string
	Id           string
	Version      string
	Name         string
	Title        string
	Status       string
	Date         string
	Description  string
	Publisher    string
	Identifier   []Identifier
	Compose      Composition
}

type PhinvadsEntry struct {
	FullUrl  string
	Resource PhinvadsResource
}

type PhinvadsResponse struct {
	ResourceType string
	Id           string
	Type         string
	Total        int64
	Link         []PhinvadsLink
	Entry        []PhinvadsEntry
}

// check - checks for an error on a result, like reading a file, etc
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func fetchThatShit(url string) PhinvadsResponse {
	var response PhinvadsResponse
	resp, err := http.Get(url)
	check(err)
	decoder := json.NewDecoder(resp.Body)
	if decoder.More() {
		err := decoder.Decode(&response)
		check(err)
	}
	return response
}

// getResponseLinks - Given a response, grabs the links available
func getResponseLinks(response *PhinvadsResponse) (string, string) {
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
func getNextResponse(nextUrl string) (*PhinvadsResponse, error) {
	// sanity check the url
	if len(strings.TrimSpace(nextUrl)) == 0 {
		return nil, errors.New("cannot fetch an empty URL")
	}
	// we might need to retry
	retryCount := 0
	for {
		fmt.Printf("Processed %d of %d records so far...\n", countEntries, totalEntries)
		fmt.Printf("Fetching %s\n\n", nextUrl)
		response := fetchThatShit(nextUrl)
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
func writeEntriesToCsv(response *PhinvadsResponse, writer *csv.Writer) string {
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

func getAllResponses() []PhinvadsResponse {
	var responses = make([]PhinvadsResponse, 0, 0)
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
	//var response *PhinvadsResponse
	//fetchPages := "?_getpages="
	fmt.Println("Starting")
	date := time.Now()
	fmt.Println("Getting all responses")
	responses := getAllResponses()
	// write out the results to a CSV
	file, err := os.Create(fmt.Sprintf("results-%s.csv", date.Format("2006-01-02")))
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
