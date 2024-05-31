# bbt-golang

## Technologies Used

- Go
- SQLite

## Exposed Endpoints

API Base URL := `https://pythonapi.mastersofterp.in/BBT/`

1. `GET /`
Endpoint to check if the API is accessible
2. `GET /customers`
Endpoint to fetch data for all users in the database
3. `DELETE /clearDB` - Requires a valid bearer token to be sent to reset the database
4. `POST /identify`
Endpoint  to identify, categorise, and process an new incoming request.

- Input:

```json
{
    "email": "your_email@example.com",
    "phoneNumber": "9090909090"
}
```

- Output Format:

```json
{
    "contact": {
        "primaryContactId": 1,
        "emails": ["your_email@example.com"],
        "phoneNumbers": ["9090909090"],
        "secondaryContactIds": []
    }
}
```

## Build Instructions

1. Clone the repository

    ```sh
    git clone https://github.com/swarnimcodes/bbt-golang
    ```

2. Change Directory

    ```sh
    cd bbt-golang
    ```

3. Build

    ```sh
    go build
    ```

4. Run

    ```sh
    ./bitespeed
    ```

5. Misc

    Refer to the `example.env` file for required environment variables
