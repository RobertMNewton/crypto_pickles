package utils

import (
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/crypto_pickle/internal/s3_client"
	"github.com/emirpasic/gods/sets/hashset"
)

func GetSymbolList(client s3_client.S3Client, bucketName string) []string {
	svc := s3.New(client.GetSession())

	symSet := hashset.New()

	err := svc.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: &bucketName,
	}, func(p *s3.ListObjectsOutput, last bool) (shouldContinue bool) {
		for _, obj := range p.Contents {
			newSym := strings.Split(*obj.Key, "/")[0]
			if !symSet.Contains(newSym) {
				symSet.Add(newSym)
			}
		}

		return true
	})

	if err != nil {
		log.Fatal(err)
	}

	res := make([]string, 0, symSet.Size())
	for _, sym := range symSet.Values() {
		res = append(res, sym.(string))
	}

	return res
}
