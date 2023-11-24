# MQTT Pipeline Service

This repository contains the source code for MQTT Pipeline Service built using Golang, Redis and MQTT. The system is responsible, -
1) generating a token for a given email, 
2) the client can publish the speed data and later subscriber consumes the message and store it in redis, 
3) the client can get the get speed data from the redis.

## Prerequisites

Before running the MQTT Pipeline Service, make sure you have the following prerequisites installed on your system:

- Go programming language (go1.21.3)
- Redis

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/ankitchahal20/mqtt-pipeline.git
   ```

2. Navigate to the project directory:

   ```bash
   cd mqtt-pipeline
   ```

3. Install the required dependencies:

   ```bash
   go mod tidy
   ```

4. Defaults.toml
Add the values to defaults.toml and execute `go run main.go` from the cmd directory.

## APIs
There are three API's which this repo currently supports.

Generate Token
```
curl -i -k -X POST \
   http://127.0.0.1:8080/v1/ \
  -H "transaction-id: 288a59c1-b826-42f7-a3cd-bf2911a5c351" \
  -H "content-type: application/json" \
  -d '{
  "original_url": "ankitchahal20@gmail.com"
}'
```
Response
```
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImFua2l0Y2hhaGFsMjBAZ21haWwuY29tIiwiZXhwIjoxNzAwODM4MTA3fQ.Npv6dm0EMIi3WKMI1-llyLD-URMxi0obgPJEpClHKts"
}
```
Generated token will be valid only for 5 mins.


Publish the Speed Data
```
curl -i -k -X POST \
  http://127.0.0.1:8080/v1/publish \
  -H "transaction-id: 288a59c1-b826-42f7-a3cd-bf2911a5c351" \
  -H "authorization: <token>" \
  -H "content-type: application/json" \
  -d '{
  "speed": 18
}'
```

Response
```
{
  "message": "Published speed data to MQTT Pipeline"
}
```

Get Latest Data

```
curl -i -k -X GET \
  http://127.0.0.1:8080/v1/ \
  -H "transaction-id: 288a59c1-b826-42f7-a3cd-bf2911a5c351" \
  -H "content-type: application/json" \
  -H "authorization: <token>"
```
Response
```
{
  "latest_speed": 99
}
```


## Project Structure

The project follows a standard Go project structure:

- `config/`: Configuration file for the application.
- `internal/`: Contains the internal packages and modules of the application.
  - `config/`: Global configuration which can be used anywhere in the application.
  - `constants/`: Contains constant values used throughout the application.
  - `models/`: Contains the data models used in the application.
  - `middleware/`: Contains code for input and token validation
  - `mqtterror`: Defines the errors in the application
  - `service/`: Contains the business logic and services of the application.
  - `server/`: Contains the server logic of the application.
  - `utils/`: Contains utility functions and helpers.
- `cmd/`:  Contains command you want to build.
    - `main.go`: Main entry point of the application.
- `README.md`: README.md contains the description for the MQTT Pipeline Service.

## Contributing

Contributions to theMQTT Pipeline Service are welcome. If you find any issues or have suggestions for improvement, feel free to open an issue or submit a pull request.

## License

The MQTT Pipeline Service is open-source and released under the [MIT License](LICENSE).

## Contact

For any inquiries or questions, please contact:

- Ankit Chahal
- ankitchahal20@gmail.com

Feel free to reach out with any feedback or concerns.
