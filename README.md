# go-phinvads
A Go utility to get data from the PHINVADS API at the CDC.

PHINVADS is a massive dataset of terminology codes and associated metadata that is stored at the CDC and available via an API. Unfortunately the API is not as intutitive as I would like, and the next and last pointers they provide for paging don't always work as expected. So I built this script to connect to the API and pull down the FHIR bundles for further processing. The codes can come from many different sources including LOINC, SNOMED, HHS, and others. The goal of the system is to host and make available common codesets in a public and accessible way.

More information can be found in the comments in the code.

An example of a valueset in PHINVADS looks like this
```json
{
    "resourceType": "ValueSet",
    "id": "2.16.840.1.114222.4.11.3213",
    "text": {
        "status": "generated",
        "div": "<div xmlns=\"http://www.w3.org/1999/xhtml\">Acute Hepatitis B Signs and Symptoms value set  is primarily based upon the SNOMED concepts that are defined in the CSTE standardized reporting definition for Acute Hepatitis B.</div>"
    },
    "url": "https://phinvads.cdc.gov/baseStu3/ValueSet/2.16.840.1.114222.4.11.3213",
    "identifier": [
        {
            "system": "urn:oid:2.16.840.1.114222.4.11.3213"
        }
    ],
    "version": "1",
    "name": "PHVS_SignsSymptoms_AcuteHepatitisB",
    "title": "Signs and Symptoms (Acute Hepatitis B)",
    "status": "active",
    "experimental": false,
    "date": "2009-07-07",
    "publisher": "CDC PHIN Vocab and Msg (PHINVS@cdc.gov)",
    "contact": [
        {
            "telecom": [
                {
                    "system": "url",
                    "value": "https://www.cdc.gov/phin/resources/standards/index.html"
                }
            ]
        }
    ],
    "description": "<div xmlns=\"http://www.w3.org/1999/xhtml\">Acute Hepatitis B Signs and Symptoms value set  is primarily based upon the SNOMED concepts that are defined in the CSTE standardized reporting definition for Acute Hepatitis B.</div>",
    "compose": {
        "lockedDate": "2009-07-22",
        "include": [
            {
                "system": "http://www.ihtsdo.org/",
                "version": "20230301",
                "concept": [
                    {
                        "code": "18165001",
                        "display": "Jaundice"
                    },
                    {
                        "code": "84387000",
                        "display": "Asymptomatic"
                    }
                ]
            }
        ]
    }
}
```

