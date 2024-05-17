# Risque Server

Backend for the spatial interaction module project, name is still undef.

Be sure to have a Postgres DB Setup.

Start the server for testing:
```bash
 go run cmd/risque-server/main.go
```
 
Compile to exec:
```bash
go build cmd/risque-server/main.go
```
---

# Docs

**GET /cityclimate**

Responds with the Sensor dataset of the ZHAW Grid, currently only has access to about 50 Sensors and the Temperature data only.


**GET /meteoblue**

Responds with a 3h forecast from Meteoblue Data, also provides a 24h overview. Data contains Temperature, Wind, Rain and some more. Each request takes 8000 tokens and our free API is limited to 10M so i beg to make only as many requests as needed.

**POST /users**

Create a User with following format

```JSON
{
 "name": "John Shoe",
 "email": "john@example.com"
}
```

**GET /users**

Responds with JSON of all users.
