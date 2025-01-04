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

	qbClient, err := quickbooks.NewClient(clientId, clientSecret, realmId, false, "", nil)
	if err != nil {
		log.Fatal(err)
	}

	// To do first when you receive the authorization code from quickbooks callback
	authorizationCode := "<received-from-callback>"
	redirectURI := "https://developer.intuit.com/v2/OAuth2Playground/RedirectUrl"
	bearerToken, err := qbClient.RetrieveBearerToken(authorizationCode, redirectURI)
	if err != nil {
		log.Fatal(err)
	}
	// Save the bearer token inside a db

	// When the token expire, you can use the following function
	bearerToken, err = qbClient.RefreshToken(bearerToken.RefreshToken)
	if err != nil {
		log.Fatal(err)
	}

	// Make a request!
	info, err := qbClient.FindCompanyInfo()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(info)

	// Revoke the token, this should be done only if a user unsubscribe from your app
	err = qbClient.RevokeToken(bearerToken.RefreshToken)
	if err != nil {
		log.Fatal(err)
	}
}
