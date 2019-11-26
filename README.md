# forgerock-go

Provides simple login functionality to ForgeRock AM protected services.

## Installation
   
```
# Go Modules
require github.com/infonova/forgerock-go v0.1.0
```

## Usage

Import forgerock into your code and refer it as forgerock.

```go
import "github.com/infonova/forgerock-go/forgerock"
```

Create a ForeRock client and login.\
Requires:
* ForgeRock base login url
* SP application url
* Credentials

```go
// Prepare input
forgerockBaseUrl := os.Getenv("FORGEROCK_BASE_URL")
appUrl := os.Getenv("APP_URL")
credentials := Credentials{
    Username: os.Getenv("FORGEROCK_USERNAME"),
    Password: os.Getenv("FORGEROCK_PASSWORD"),
}

// Create ForgeRock client
fr, err := forgerock.New(forgerockBaseUrl)
if err != nil {
    return errors.New("failed to create ForgeRock client: " + err.Error())
}

// Login to a ServiceProvider via ForgeRock
client, err := fr.Login(appUrl, credentials)
if err != nil {
    return errors.New("failed to login: " + err.Error())
}

// Use the returned client to make API calls
resp, err := client.R().
    SetHeader("Accept", "application/json").
    Get(appUrl + "/api/tenants")
if err != nil {
    return errors.New("failed to request tenants: " + err)
}

fmt.Println(resp.String())
```

## Versioning

forgerock-go provides tags that follow [Semantic Versioning](https://semver.org/).

## License

forgerock-go is licensed under MIT. Refer to [LICENSE](LICENSE) for details.
