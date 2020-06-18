package countrydb

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type ddbClient interface {
	UpdateItem(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error)
	ScanPages(input *dynamodb.ScanInput, fn func(*dynamodb.ScanOutput, bool) bool) error
}

type ddb struct {
	tableName string
	client ddbClient
}

func NewDynamoDB() DB {
	sess := session.Must(session.NewSession())

	log.Printf("Initializing DynamoDB client connected to table: %s\n", os.Getenv("TRAVELS_TABLE_NAME"))
	// Create a DynamoDB client from just a session.
	svc := dynamodb.New(sess)
	return &ddb{
		tableName: os.Getenv("TRAVELS_TABLE_NAME"),
		client: svc,
	}
}

func (db *ddb) Save(country string) (int, error) {
	out, err := db.client.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(db.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Country": {
				S:   aws.String(country),
			},
		},
		UpdateExpression: aws.String("SET Visit = Visit + :incr"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":incr": {
				N: aws.String("1"),
			},
		},
		ReturnValues: aws.String(dynamodb.ReturnValueUpdatedNew),
	})
	if err != nil {
		return 0, fmt.Errorf("countrydb: update item %s: %w", country, err)
	}
	newCount, err := strconv.Atoi(*out.Attributes["Visit"].N)
	if err != nil {
		return 0, fmt.Errorf("countrydb: strconv attributes Visit %s: %w", *out.Attributes["Visit"].N, err)
	}
	return newCount, nil
}

func (db *ddb) Results() ([]Country, error) {
	countries := []Country{} // Initialize non-empty so that the front-end renders.
	in := &dynamodb.ScanInput{
		TableName: aws.String(db.tableName),
	}
	err := db.client.ScanPages(in,
		func(page *dynamodb.ScanOutput, lastPage bool) bool {
			for _, item := range page.Items {
				visit, _ := strconv.Atoi(*item["Visit"].N)
				countries = append(countries, Country{
					Country: *item["Country"].S,
					Visit: visit,
				})
			}
			return !lastPage
		})
	if err != nil {
		return nil, fmt.Errorf("countrydb: scan countries: %w", err)
	}
	return countries, nil
}

func (db *ddb) UniqueTotal() (int, error) {
	countries, err := db.Results()
	if err != nil {
		return 0, err
	}
	var unique int
	for _, country := range countries {
		if country.Visit > 0 {
			unique += 1
		}
	}
	return unique, nil
}
