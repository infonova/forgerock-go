package forgerock

import (
	"fmt"
	"os"
	"testing"
)

func TestLogin(t *testing.T) {
	appUrl := os.Getenv("APP_URL")
	forgerockUsername := os.Getenv("FORGEROCK_USERNAME")
	forgerockPassword := os.Getenv("FORGEROCK_PASSWORD")

	type args struct {
		appUrl      string
		credentials Credentials
	}
	tests := []struct {
		name            string
		args            args
		wantRestyClient bool
		wantErr         bool
	}{
		{
			name: "Successful login",
			args: args{
				appUrl: appUrl,
				credentials: Credentials{
					Username: forgerockUsername,
					Password: forgerockPassword,
				},
			},
			wantRestyClient: true,
			wantErr:         false,
		},
		{
			name: "Wrong credentials",
			args: args{
				appUrl: appUrl,
				credentials: Credentials{
					Username: "forgerock-go-test-username",
					Password: "forgerock-go-test-password",
				},
			},
			wantRestyClient: false,
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			forgerock, err := New(os.Getenv("FORGEROCK_BASE_URL"))
			if err != nil {
				t.Errorf("New() error = %v", err)
				return
			}
			restyClient, err := forgerock.Login(tt.args.appUrl, tt.args.credentials)
			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (restyClient != nil) != tt.wantRestyClient {
				t.Errorf("Login() restyClient = %v, wantRestyClient %v", restyClient, tt.wantRestyClient)
			}
		})
	}
}

func TestZuulTenants(t *testing.T) {
	forgerockBaseUrl := os.Getenv("FORGEROCK_BASE_URL")
	appUrl := os.Getenv("APP_URL")
	credentials := Credentials{
		Username: os.Getenv("FORGEROCK_USERNAME"),
		Password: os.Getenv("FORGEROCK_PASSWORD"),
	}

	forgerock, err := New(forgerockBaseUrl)
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}

	client, err := forgerock.Login(appUrl, credentials)
	if err != nil {
		t.Errorf("Login() error = %v", err)
		return
	}

	resp, err := client.R().
		SetHeader("Accept", "application/json").
		Get(appUrl + "/api/tenants")
	if err != nil {
		t.Errorf("Tenants error = %v", err)
		return
	}

	fmt.Println(resp.String())
}
