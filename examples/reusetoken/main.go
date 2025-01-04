package main

import (
	"fmt"
	"log"

	"github.com/joefitzgerald/quickbooks-go/v2"
)

func main() {
	clientId := "<your-client-id>"
	clientSecret := "<your-client-secret>"
	realmId := "<realm-id>"

	token := quickbooks.BearerToken{
		RefreshToken: "<saved-refresh-token>",
		AccessToken:  "<saved-access-token>",
	}

	qbClient, err := quickbooks.NewClient(clientId, clientSecret, realmId, false, "", &token)
	if err != nil {
		log.Fatal(err)
	}

	// Make a request!
	info, err := qbClient.FindCompanyInfo()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(info)
}
