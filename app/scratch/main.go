package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/open-policy-agent/opa/rego"
)

func main() {
	err := genToken()
	// err := genKey()
	if err != nil {
		log.Fatal(err)
	}
}

func genToken() error {
	// Generating a token requires defining a set of claims. In this applications
	// case, we only care about defining the subject and the user in question and
	// the roles they have on the database. This token will expire in a year.
	//
	// iss (issuer): Issuer of the JWT
	// sub (subject): Subject of the JWT (the user)
	// aud (audience): Recipient for which the JWT is intended
	// exp (expiration time): Time after which the JWT expires
	// nbf (not before time): Time before which the JWT must not be accepted for processing
	// iat (issued at time): Time at which the JWT was issued; can be used to determine the age of the JWT
	// jti (JWT ID): Unique identifier; can be used to prevent the JWT from being replayed (allows a token to be used only once)
	claims := struct {
		jwt.RegisteredClaims
		Roles []string
	}{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "12345678",
			Issuer:    "service project",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(8760 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: []string{"ADMIN"},
	}

	method := jwt.GetSigningMethod(jwt.SigningMethodRS256.Name)

	token := jwt.NewWithClaims(method, claims)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("signing token: %w", err)
	}

	tokenStr, err := token.SignedString(privateKey)
	if err != nil {
		return err
	}

	fmt.Println("******* TOKEN ********")
	fmt.Println(tokenStr)

	// -------------------------------------------------------------------------
	// Output public key to stdout.

	// Marshal the public key from the private key to PKIX.
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshaling public key: %w", err)
	}

	// Construct a PEM block for the public key.
	publicBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	// -------------------------------------------------------------------------
	// Validate the token.

	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}))

	var clm struct {
		jwt.RegisteredClaims
		Roles []string
	}
	kf := func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	}
	tkn, err := parser.ParseWithClaims(tokenStr, &clm, kf)
	if err != nil {
		return fmt.Errorf("parsing token: %w", err)
	}

	if !tkn.Valid {
		return fmt.Errorf("token is invalid")
	}

	fmt.Println("\nTOKEN VALIDATED")

	// -------------------------------------------------------------------------
	// OPA policy evaluation - authentication

	// Write the public key to the public key file.
	var buf bytes.Buffer
	if err := pem.Encode(&buf, &publicBlock); err != nil {
		return fmt.Errorf("encoding to public key: %w", err)
	}

	err = opaPolicyEvaluationAuthen(context.Background(), buf.String(), tokenStr, clm.Issuer)
	if err != nil {
		return fmt.Errorf("opa authentication policy failed: %w", err)
	}

	fmt.Println("TOKEN VALIDATED BY OPA")
	fmt.Printf("\n%#v\n", clm)

	// -------------------------------------------------------------------------
	// OPA policy evaluation - authorization

	err = opaPolicyEvaluationAuthor(context.Background())
	if err != nil {
		return fmt.Errorf("opa authorization policy failed: %w", err)
	}

	fmt.Println("\nTOKEN AUTHORIZED BY OPA")

	return nil
}

func genKey() error {
	// Generate a new private key.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("generating key: %w", err)
	}

	// Create a file for the private key information in PEM form.
	privateFile, err := os.Create("private.pem")
	if err != nil {
		return fmt.Errorf("creating private key file: %w", err)
	}
	defer privateFile.Close()

	// Construct a PEM block for the private key
	privateBlock := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Write the private key to the private key file.
	if err := pem.Encode(privateFile, &privateBlock); err != nil {
		return fmt.Errorf("encoding to private key: %w", err)
	}

	// -------------------------------------------------------------------------

	// Create a file for the public key information in PEM form.
	publicFile, err := os.Create("public.pem")
	if err != nil {
		return fmt.Errorf("creating public key file: %w", err)
	}
	defer publicFile.Close()

	// Marshal the public key from the private key to PKIX.
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshaling public key: %w", err)
	}

	// Construct a PEM block for the public key.
	publicBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	// Write the public key to the public key file.
	if err := pem.Encode(publicFile, &publicBlock); err != nil {
		return fmt.Errorf("encoding to public key: %w", err)
	}

	fmt.Println("private and public key files generated")

	return nil
}

// Core OPA policies
var (
	//go:embed rego/authentication.rego
	opaAuthentication string

	//go:embed rego/authorization.rego
	opaAuthorization string
)

func opaPolicyEvaluationAuthor(ctx context.Context) error {
	const opaPackage = "ardan.rego"
	const rule = "ruleAdminOnly"
	query := fmt.Sprintf("x = data.%s.%s", opaPackage, rule)

	q, err := rego.New(
		rego.Query(query),
		rego.Module("policy.rego", opaAuthorization),
	).PrepareForEval(ctx)
	if err != nil {
		return err
	}

	input := map[string]any{
		"Roles":   []string{"ADMIN"},
		"Subject": "1234567",
		"UserID":  "1234567",
	}

	results, err := q.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	if len(results) == 0 {
		return errors.New("no results")
	}

	result, ok := results[0].Bindings["x"].(bool)
	if !ok || !result {
		return fmt.Errorf("bindings result[%v] ok[%v]", result, ok)
	}

	return nil
}

func opaPolicyEvaluationAuthen(ctx context.Context, pem string, tokenString string, issuer string) error {
	const opaPackage = "ardan.rego"
	const rule = "auth"
	query := fmt.Sprintf("x = data.%s.%s", opaPackage, rule)

	q, err := rego.New(
		rego.Query(query),
		rego.Module("policy.rego", opaAuthentication),
	).PrepareForEval(ctx)
	if err != nil {
		return err
	}

	input := map[string]any{
		"Key":   pem,
		"Token": tokenString,
		"ISS":   issuer,
	}

	results, err := q.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	if len(results) == 0 {
		return errors.New("no results")
	}

	result, ok := results[0].Bindings["x"].(bool)
	if !ok || !result {
		return fmt.Errorf("bindings result[%v] ok[%v]", result, ok)
	}

	return nil
}
