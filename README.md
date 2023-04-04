# go-phinvads
A Go utility to get data from the PHINVADS API at the CDC.

PHINVADS is a massive dataset of LOINC codes and associated metadata that is stored at the CDC and available via an API. Unfortunately the API is not as intutitive as I would like, and the next and last pointers they provide for paging don't always work as expected. So I built this script to connect to the API and pull down the FHIR bundles for further processing.

More information can be found in the comments in the code.

